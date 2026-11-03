// Package runpipeline executes one run from the stored state to the end, or to the
// point where the operator has to look. It is the only place that calls the cloud.
package runpipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"math/rand/v2"
	"strings"
	"sync"
	"time"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/domain/vo"
)

const (
	EventRun       = "run"
	EventStep      = "step"
	EventBenchmark = "benchmark"
)

type Usecase struct {
	runs       port.RunRepository
	steps      port.StepRepository
	benchmarks port.BenchmarkRepository
	models     port.Models
	detector   port.Detector
	events     port.EventBus
	cfg        Config
	now        func() time.Time
	benchmarkM sync.Mutex
}

func New(
	runs port.RunRepository,
	steps port.StepRepository,
	benchmarks port.BenchmarkRepository,
	models port.Models,
	detector port.Detector,
	events port.EventBus,
	cfg Config,
) *Usecase {
	return &Usecase{
		runs: runs, steps: steps, benchmarks: benchmarks, models: models,
		detector: detector, events: events, cfg: cfg, now: time.Now,
	}
}

// modelError is a failed model call: the run fails, the pipeline does not.
type modelError struct {
	cause error
}

func (e *modelError) Error() string { return e.cause.Error() }
func (e *modelError) Unwrap() error { return e.cause }

// Execute runs a queued run to the end, or resumes one the operator released.
func (u *Usecase) Execute(ctx context.Context, runID uuid.UUID) error {
	run, err := u.runs.Get(ctx, runID)
	if err != nil {
		return fmt.Errorf("load run: %w", err)
	}
	if run == nil {
		return fmt.Errorf("run %s: not found", runID)
	}
	existing, err := u.steps.ListByRun(ctx, runID)
	if err != nil {
		return fmt.Errorf("load steps: %w", err)
	}
	exec := &execution{u: u, run: run, seq: len(existing)}

	err = exec.proceed(ctx)
	var failed *modelError
	if errors.As(err, &failed) {
		run.Fail(failed.Error(), u.now())
		slog.WarnContext(ctx, "run failed", slog.String("run_id", runID.String()),
			slog.String("error", failed.Error()))
		err = nil
	}
	if err != nil {
		return err
	}
	if err := u.save(ctx, run); err != nil {
		return err
	}
	if run.Status() == vo.RunDone || run.Status() == vo.RunFailed {
		return u.advanceBenchmark(ctx, run)
	}

	return nil
}

func (u *Usecase) save(ctx context.Context, run *entity.Run) error {
	if err := u.runs.Update(ctx, run); err != nil {
		return fmt.Errorf("save run: %w", err)
	}
	u.events.Publish(port.Event{Name: EventRun, Data: run})
	return nil
}

func (u *Usecase) advanceBenchmark(ctx context.Context, run *entity.Run) error {
	if run.BenchmarkID() == uuid.Nil() {
		return nil
	}
	u.benchmarkM.Lock()
	defer u.benchmarkM.Unlock()

	benchmark, err := u.benchmarks.Get(ctx, run.BenchmarkID())
	if err != nil {
		return fmt.Errorf("load benchmark: %w", err)
	}
	if benchmark == nil {
		return nil
	}
	benchmark.Advance()
	if err := u.benchmarks.Update(ctx, benchmark); err != nil {
		return fmt.Errorf("save benchmark: %w", err)
	}
	u.events.Publish(port.Event{Name: EventBenchmark, Data: benchmark})
	return nil
}

// execution is one Execute call: the run, the step counter and what went to the cloud.
type execution struct {
	u             *Usecase
	run           *entity.Run
	seq           int
	cloudPayloads []string
}

func (e *execution) proceed(ctx context.Context) error {
	if e.run.Status() == vo.RunQueued {
		if err := e.run.Start(e.u.now()); err != nil {
			return err
		}
		if err := e.u.save(ctx, e.run); err != nil {
			return err
		}
	}
	if e.run.Phase() == vo.PhaseAnonymize {
		stopped, err := e.anonymize(ctx)
		if err != nil || stopped {
			return err
		}
	}
	if e.run.Phase() == vo.PhaseCompose {
		return e.compose(ctx)
	}
	return nil
}

// anonymize is vision, gate 1 and gate 2. It reports true when the run was parked
// for review.
func (e *execution) anonymize(ctx context.Context) (bool, error) {
	text := e.run.InputText()
	if e.run.Source() == vo.SourceImage {
		transcript, err := e.transcribe(ctx)
		if err != nil {
			return false, err
		}
		text = transcript
		e.run.SetTranscript(text, e.u.now())
	}

	started := e.u.now()
	vault := vo.NewVault()
	spans := e.u.detector.Detect(text)
	before := text
	text = vo.ApplySpans(text, spans, vault)
	err := e.record(ctx, stepInput{
		kind: vo.StepGate1, side: vo.SideSystem, status: vo.StepOK, startedAt: started,
		input:  map[string]any{"text": before},
		output: map[string]any{"spans": spans, "text": text, "vault": vault.Entries()},
	})
	if err != nil {
		return false, err
	}

	previous := -1
	var lastSpans []vo.Span
	for pass := 1; pass <= e.u.cfg.Gate2MaxPasses; pass++ {
		found, replaced, err := e.gate2Pass(ctx, text, vault, pass)
		if err != nil {
			return false, err
		}
		if found == 0 {
			e.run.SetAnonymized(text, vault, pass, e.u.now())
			return false, e.u.save(ctx, e.run)
		}
		text = replaced.text
		lastSpans = replaced.spans
		stuck := previous >= 0 && found >= 0 && found >= previous
		if pass == e.u.cfg.Gate2MaxPasses || stuck {
			e.run.Escalate(text, vault, lastSpans, pass, e.u.now())
			return true, e.u.save(ctx, e.run)
		}
		previous = found
	}

	return false, nil
}

type gate2Result struct {
	text  string
	spans []vo.Span
}

// gate2Pass asks the weak model once. It returns how many spans were found, -1
// when the model failed: an unknown answer is not a clean one.
func (e *execution) gate2Pass(
	ctx context.Context,
	text string,
	vault *vo.Vault,
	pass int,
) (int, gate2Result, error) {
	started := e.u.now()
	user := "Плейсхолдеры, которые уже есть в тексте: " + strings.Join(vault.Placeholders(), ", ") +
		"\n\nТекст:\n" + text
	input := map[string]any{"pass": pass, "prompt": promptPayload{System: systemGate2, User: user}}

	resp, err := e.u.models.Local.Complete(ctx, port.ChatRequest{System: systemGate2, User: user, JSON: true})
	if err != nil {
		return -1, gate2Result{text: text}, e.record(ctx, stepInput{
			kind: vo.StepGate2, side: vo.SideLocal, status: vo.StepFailed, startedAt: started,
			input: input, err: err,
		})
	}
	var parsed gate2Response
	if err := decodeJSON(resp.Content, &parsed); err != nil {
		return -1, gate2Result{text: text}, e.record(ctx, stepInput{
			kind: vo.StepGate2, side: vo.SideLocal, status: vo.StepFailed, startedAt: started,
			input: input, output: map[string]any{"raw": resp.Content}, resp: resp, err: err,
		})
	}

	var spans []vo.Span
	for _, s := range parsed.Spans {
		if strings.TrimSpace(s.Text) == "" || !strings.Contains(text, s.Text) {
			continue
		}
		kind := vo.ParsePIIType(s.Type)
		if kind.Numeric() && len(vo.Digits(s.Text)) < 4 {
			continue
		}
		spans = append(spans, vo.Span{Text: s.Text, Type: kind, Canonical: s.Canonical})
	}
	replaced := vo.ApplySpans(text, spans, vault)
	err = e.record(ctx, stepInput{
		kind: vo.StepGate2, side: vo.SideLocal, status: vo.StepOK, startedAt: started,
		input:  input,
		output: map[string]any{"spans": spans, "text": replaced, "vault": vault.Entries()},
		resp:   resp,
	})
	return len(spans), gate2Result{text: replaced, spans: spans}, err
}

func (e *execution) transcribe(ctx context.Context) (string, error) {
	started := e.u.now()
	user := "Распознай текст на изображении."
	input := map[string]any{"prompt": promptPayload{System: systemVision, User: user,
		Image: fmt.Sprintf("%s, %d bytes", e.run.ImageMIME(), e.run.ImageSize())}}
	resp, err := e.u.models.Vision.Complete(ctx, port.ChatRequest{
		System: systemVision, User: user,
		Image: &port.ImageData{Data: e.run.Image(), MIME: e.run.ImageMIME()},
	})
	if err != nil {
		return "", e.fail(
			ctx,
			stepInput{kind: vo.StepVision, side: vo.SideLocal, startedAt: started, input: input},
			err,
		)
	}
	text := strings.TrimSpace(resp.Content)
	if text == "" {
		return "", e.fail(
			ctx,
			stepInput{kind: vo.StepVision, side: vo.SideLocal, startedAt: started, input: input},
			errors.New("vision model returned no text"),
		)
	}
	err = e.record(ctx, stepInput{
		kind: vo.StepVision, side: vo.SideLocal, status: vo.StepOK, startedAt: started,
		input: input, output: map[string]any{"text": text}, resp: resp,
	})
	return text, err
}

// compose is draft, adviser, finalize, judge and re-identification.
func (e *execution) compose(ctx context.Context) error {
	card, draft, err := e.draft(ctx)
	if err != nil {
		return err
	}
	e.run.SetDraft(card, draft, e.u.now())
	if err := e.u.save(ctx, e.run); err != nil {
		return err
	}

	advice, err := e.advise(ctx, card, draft)
	if err != nil {
		return err
	}

	final := draft
	if len(advice) > 0 {
		if final, err = e.finalize(ctx, draft, advice); err != nil {
			return err
		}
	} else {
		err := e.record(ctx, stepInput{
			kind: vo.StepFinalize, side: vo.SideSystem, status: vo.StepSkipped, startedAt: e.u.now(),
			input: map[string]any{"reason": "no advice to apply"}, output: map[string]any{"text": final},
		})
		if err != nil {
			return err
		}
	}

	if !e.run.AdvisorSkippedByReview() {
		if err := e.judge(ctx, card, final); err != nil {
			return err
		}
	}

	started := e.u.now()
	vault := e.run.Vault()
	text := vault.Reidentify(final)
	e.run.SetFinal(final, text, e.u.now())
	err = e.record(ctx, stepInput{
		kind: vo.StepReidentify, side: vo.SideSystem, status: vo.StepOK, startedAt: started,
		input: map[string]any{"text": final}, output: map[string]any{"text": text},
	})
	if err != nil {
		return err
	}

	return e.run.Finish(countCanaries(e.run.Canaries(), e.cloudPayloads), e.u.now())
}

func (e *execution) draft(ctx context.Context) (vo.Card, string, error) {
	started := e.u.now()
	user := "Обезличенный текст претензии:\n" + e.run.Anonymized()
	input := map[string]any{"prompt": promptPayload{System: systemDraft, User: user}}
	base := stepInput{kind: vo.StepDraft, side: vo.SideLocal, startedAt: started, input: input}

	resp, err := e.u.models.Local.Complete(ctx, port.ChatRequest{System: systemDraft, User: user, JSON: true})
	if err != nil {
		return vo.Card{}, "", e.fail(ctx, base, err)
	}
	var parsed draftResponse
	if err := decodeJSON(resp.Content, &parsed); err != nil {
		base.output = map[string]any{"raw": resp.Content}
		return vo.Card{}, "", e.fail(ctx, base, err)
	}
	if strings.TrimSpace(parsed.Draft) == "" {
		base.output = map[string]any{"raw": resp.Content}
		return vo.Card{}, "", e.fail(ctx, base, errors.New("draft is empty"))
	}
	base.status, base.resp = vo.StepOK, resp
	base.output = map[string]any{"card": parsed.Card, "draft": parsed.Draft}
	return parsed.Card, parsed.Draft, e.record(ctx, base)
}

func (e *execution) advise(ctx context.Context, card vo.Card, draft string) ([]string, error) {
	if e.run.AdvisorSkippedByReview() || e.run.AdvisorMode() == vo.AdvisorOff {
		reason := "advisor mode is off"
		if e.run.AdvisorSkippedByReview() {
			reason = "operator chose to finish without the adviser"
		}
		e.run.SetAdvice(nil, vo.OutcomeSkipped, e.u.now())
		return nil, e.record(ctx, stepInput{
			kind: vo.StepAdvise, side: vo.SideSystem, status: vo.StepSkipped, startedAt: e.u.now(),
			input: map[string]any{"reason": reason},
		})
	}

	if e.run.AdvisorMode() == vo.AdvisorPlacebo {
		remarks := placebo(e.run.ID())
		e.run.SetAdvice(remarks, vo.OutcomePlacebo, e.u.now())
		return remarks, e.record(ctx, stepInput{
			kind: vo.StepAdvise, side: vo.SidePlacebo, status: vo.StepOK, startedAt: e.u.now(),
			input:  map[string]any{"reason": "placebo: generic remarks, nothing sent anywhere"},
			output: map[string]any{"remarks": remarks},
		})
	}

	user := "Карточка обращения:\n" + encodeJSON(card) + "\n\nЧерновик ответа:\n" + draft
	resp, blocked, err := e.callCloud(ctx, vo.StepAdvise, systemAdvise, user)
	if err != nil {
		return nil, err
	}
	if blocked {
		e.run.SetAdvice(nil, vo.OutcomeBlocked, e.u.now())
		return nil, nil
	}
	var parsed adviseResponse
	if err := decodeJSON(resp.content, &parsed); err != nil {
		resp.base.output = map[string]any{"raw": resp.content}
		return nil, e.fail(ctx, resp.base, err)
	}
	remarks := make([]string, 0, len(parsed.Remarks))
	for _, r := range parsed.Remarks {
		if strings.TrimSpace(r) != "" && len(remarks) < 5 {
			remarks = append(remarks, strings.TrimSpace(r))
		}
	}
	e.run.SetAdvice(remarks, vo.OutcomeApplied, e.u.now())
	resp.base.status = vo.StepOK
	resp.base.output = map[string]any{"remarks": remarks}
	return remarks, e.record(ctx, resp.base)
}

func (e *execution) finalize(ctx context.Context, draft string, advice []string) (string, error) {
	started := e.u.now()
	var b strings.Builder
	b.WriteString("Черновик:\n" + draft + "\n\nПравки эксперта:\n")
	for i, remark := range advice {
		fmt.Fprintf(&b, "%d. %s\n", i+1, remark)
	}
	user := b.String()
	base := stepInput{kind: vo.StepFinalize, side: vo.SideLocal, startedAt: started,
		input: map[string]any{"prompt": promptPayload{System: systemFinalize, User: user}}}

	resp, err := e.u.models.Local.Complete(ctx, port.ChatRequest{System: systemFinalize, User: user})
	if err != nil {
		return "", e.fail(ctx, base, err)
	}
	text := strings.TrimSpace(resp.Content)
	if text == "" {
		return "", e.fail(ctx, base, errors.New("final text is empty"))
	}
	base.status, base.resp, base.output = vo.StepOK, resp, map[string]any{"text": text}
	return text, e.record(ctx, base)
}

func (e *execution) judge(ctx context.Context, card vo.Card, final string) error {
	user := "Карточка обращения:\n" + encodeJSON(card) + "\n\nОтвет клиенту:\n" + final
	resp, blocked, err := e.callCloud(ctx, vo.StepJudge, systemJudge, user)
	if err != nil || blocked {
		return err
	}
	var parsed judgeResponse
	if err := decodeJSON(resp.content, &parsed); err != nil {
		resp.base.output = map[string]any{"raw": resp.content}
		return e.fail(ctx, resp.base, err)
	}
	judgement := vo.Judgement{Score: parsed.Score, Reason: strings.TrimSpace(parsed.Reason)}
	if err := e.run.SetJudgement(judgement, e.u.now()); err != nil {
		resp.base.output = map[string]any{"raw": resp.content}
		return e.fail(ctx, resp.base, err)
	}
	resp.base.status = vo.StepOK
	resp.base.output = judgement
	return e.record(ctx, resp.base)
}

type cloudResponse struct {
	content string
	base    stepInput
}

// callCloud is the only path to the strong model. The egress check runs first and
// records its own step; a finding means the call is never made.
func (e *execution) callCloud(
	ctx context.Context,
	kind vo.StepKind,
	system, user string,
) (*cloudResponse, bool, error) {
	started := e.u.now()
	findings := Egress(user, e.run.Vault(), e.u.detector)
	status := vo.StepOK
	if len(findings) > 0 {
		status = vo.StepBlocked
	}
	err := e.record(ctx, stepInput{
		kind: vo.StepEgress, side: vo.SideSystem, status: status, startedAt: started,
		input:  map[string]any{"target": kind, "payload": user},
		output: map[string]any{"findings": findings},
	})
	if err != nil {
		return nil, false, err
	}
	if len(findings) > 0 {
		return nil, true, e.record(ctx, stepInput{
			kind: kind, side: vo.SideCloud, status: vo.StepBlocked, startedAt: e.u.now(),
			input: map[string]any{"reason": "egress check found personal data, nothing was sent"},
		})
	}

	started = e.u.now()
	e.cloudPayloads = append(e.cloudPayloads, system, user)
	base := stepInput{kind: kind, side: vo.SideCloud, startedAt: started,
		input: map[string]any{"prompt": promptPayload{System: system, User: user}}}
	resp, err := e.u.models.Cloud.Complete(ctx, port.ChatRequest{System: system, User: user, JSON: true})
	if err != nil {
		return nil, false, e.fail(ctx, base, err)
	}
	base.resp = resp
	return &cloudResponse{content: resp.Content, base: base}, false, nil
}

type stepInput struct {
	kind      vo.StepKind
	side      vo.Side
	status    vo.StepStatus
	startedAt time.Time
	input     any
	output    any
	resp      *port.ChatResponse
	err       error
}

// record stores one step and announces it.
func (e *execution) record(ctx context.Context, in stepInput) error {
	e.seq++
	state := &entity.StepState{
		RunID:      e.run.ID(),
		Seq:        e.seq,
		Kind:       in.kind,
		Side:       in.side,
		Status:     in.status,
		StartedAt:  in.startedAt,
		FinishedAt: e.u.now(),
		Input:      encodeRaw(in.input),
		Output:     encodeRaw(in.output),
	}
	if in.resp != nil {
		state.Model = in.resp.Model
		state.PromptTokens = in.resp.PromptTokens
		state.CompletionTokens = in.resp.CompletionTokens
	}
	if in.err != nil {
		state.Error = in.err.Error()
	}
	step, err := entity.NewStep(state)
	if err != nil {
		return fmt.Errorf("build step: %w", err)
	}
	if err := e.u.steps.Create(ctx, step); err != nil {
		return fmt.Errorf("save step: %w", err)
	}
	e.u.events.Publish(port.Event{Name: EventStep, Data: step})
	return nil
}

// fail records a failed model step and turns the cause into a run failure.
func (e *execution) fail(ctx context.Context, in stepInput, cause error) error {
	in.status = vo.StepFailed
	in.err = cause
	if err := e.record(ctx, in); err != nil {
		return err
	}
	return &modelError{cause: fmt.Errorf("%s: %w", in.kind, cause)}
}

func placebo(runID uuid.UUID) []string {
	h := fnv.New64a()
	_, _ = h.Write(runID[:])
	r := rand.New(rand.NewPCG(h.Sum64(), 0))
	perm := r.Perm(len(placeboRemarks))
	out := make([]string, 0, 3)
	for _, i := range perm[:3] {
		out = append(out, placeboRemarks[i])
	}
	return out
}

// decodeJSON tolerates code fences and chatter around the object.
func decodeJSON(content string, dst any) error {
	trimmed := strings.TrimSpace(content)
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start < 0 || end < start {
		return errors.New("model returned no JSON object")
	}
	if err := json.Unmarshal([]byte(trimmed[start:end+1]), dst); err != nil {
		return fmt.Errorf("model returned invalid JSON: %w", err)
	}
	return nil
}

func encodeJSON(v any) string {
	raw, _ := json.MarshalIndent(v, "", "  ")
	return string(raw)
}

func encodeRaw(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		raw, _ = json.Marshal(map[string]string{"error": err.Error()})
	}
	return raw
}
