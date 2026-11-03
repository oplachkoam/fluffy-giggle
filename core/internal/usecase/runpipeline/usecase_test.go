package runpipeline

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/domain/vo"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.DiscardHandler))
	os.Exit(m.Run())
}

const (
	phone  = "+7 916 123-45-67"
	letter = "Я, Иванов Пётр Сергеевич, тел. " + phone + ", требую выплату 120 000 рублей."
)

// stubDetector is gate 1 reduced to one phone pattern.
type stubDetector struct{}

var phoneRe = regexp.MustCompile(`\+7 \d{3} \d{3}-\d{2}-\d{2}`)

func (stubDetector) Detect(text string) []vo.Span {
	var spans []vo.Span
	for _, m := range phoneRe.FindAllString(text, -1) {
		spans = append(spans, vo.Span{Text: m, Type: vo.PIIPhone, Canonical: vo.Digits(m)})
	}
	return spans
}

// fakeModel answers from a script and remembers what it was asked.
type fakeModel struct {
	mu        sync.Mutex
	name      string
	responses []string
	errs      []error
	requests  []port.ChatRequest
}

func (m *fakeModel) Complete(_ context.Context, req port.ChatRequest) (*port.ChatResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, req)
	i := len(m.requests) - 1
	if i < len(m.errs) && m.errs[i] != nil {
		return nil, m.errs[i]
	}
	if i >= len(m.responses) {
		return nil, errors.New("unexpected call to " + m.name)
	}
	return &port.ChatResponse{
		Content:          m.responses[i],
		Model:            m.name,
		PromptTokens:     10,
		CompletionTokens: 5,
	}, nil
}

type memory struct {
	mu         sync.Mutex
	runs       map[uuid.UUID]*entity.Run
	steps      []*entity.Step
	benchmarks map[uuid.UUID]*entity.Benchmark
	events     []port.Event
}

func newMemory() *memory {
	return &memory{runs: map[uuid.UUID]*entity.Run{}, benchmarks: map[uuid.UUID]*entity.Benchmark{}}
}

func (m *memory) Create(ctx context.Context, run *entity.Run) error { return m.Update(ctx, run) }
func (m *memory) Update(_ context.Context, run *entity.Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[run.ID()] = run
	return nil
}
func (m *memory) Get(_ context.Context, id uuid.UUID) (*entity.Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.runs[id], nil
}
func (m *memory) List(context.Context, int, uuid.UUID) ([]*entity.Run, error) { return nil, nil }
func (m *memory) ListByBenchmark(context.Context, uuid.UUID) ([]*entity.Run, error) {
	return nil, nil
}

type stepStore struct{ m *memory }

func (s stepStore) Create(_ context.Context, step *entity.Step) error {
	s.m.mu.Lock()
	defer s.m.mu.Unlock()
	s.m.steps = append(s.m.steps, step)
	return nil
}
func (s stepStore) ListByRun(_ context.Context, runID uuid.UUID) ([]*entity.Step, error) {
	s.m.mu.Lock()
	defer s.m.mu.Unlock()
	var out []*entity.Step
	for _, step := range s.m.steps {
		if step.RunID() == runID {
			out = append(out, step)
		}
	}
	return out, nil
}

type benchStore struct{ m *memory }

func (b benchStore) Create(ctx context.Context, bm *entity.Benchmark) error { return b.Update(ctx, bm) }
func (b benchStore) Update(_ context.Context, bm *entity.Benchmark) error {
	b.m.mu.Lock()
	defer b.m.mu.Unlock()
	b.m.benchmarks[bm.ID()] = bm
	return nil
}
func (b benchStore) Get(_ context.Context, id uuid.UUID) (*entity.Benchmark, error) {
	b.m.mu.Lock()
	defer b.m.mu.Unlock()
	return b.m.benchmarks[id], nil
}
func (b benchStore) List(context.Context) ([]*entity.Benchmark, error) { return nil, nil }

type bus struct{ m *memory }

func (b bus) Publish(event port.Event) {
	b.m.mu.Lock()
	defer b.m.mu.Unlock()
	b.m.events = append(b.m.events, event)
}
func (b bus) Subscribe(context.Context) <-chan port.Event { return nil }

type harness struct {
	m     *memory
	local *fakeModel
	cloud *fakeModel
	uc    *Usecase
}

func newHarness(t *testing.T, local, cloud []string) *harness {
	t.Helper()
	m := newMemory()
	h := &harness{
		m:     m,
		local: &fakeModel{name: "weak", responses: local},
		cloud: &fakeModel{name: "strong", responses: cloud},
	}
	h.uc = New(m, stepStore{m}, benchStore{m}, port.Models{
		Local: h.local, Vision: &fakeModel{name: "vision"}, Cloud: h.cloud,
	}, stubDetector{}, bus{m}, Config{Gate2MaxPasses: 3})
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	h.uc.now = func() time.Time { now = now.Add(time.Second); return now }
	return h
}

func (h *harness) run(t *testing.T, mode vo.AdvisorMode, canaries []string) *entity.Run {
	t.Helper()
	run, err := entity.NewRun(entity.NewRunInput{
		Source: vo.SourceText, AdvisorMode: mode, Text: letter, Canaries: canaries,
	}, h.uc.now())
	require.NoError(t, err)
	require.NoError(t, h.m.Create(t.Context(), run))
	return run
}

func (h *harness) kinds(t *testing.T, runID uuid.UUID) []string {
	t.Helper()
	steps, _ := stepStore{h.m}.ListByRun(t.Context(), runID)
	out := make([]string, 0, len(steps))
	for _, s := range steps {
		out = append(out, string(s.Kind())+":"+string(s.Side())+":"+string(s.Status()))
	}
	return out
}

const (
	gate2Name  = `{"spans":[{"text":"Иванов Пётр Сергеевич","type":"PERSON","canonical":"Иванов Пётр Сергеевич"}]}`
	gate2Clean = `{"spans":[]}`
	draftJSON  = `{"card":{"category":"выплата","subject":"задержка выплаты","amount_range":"50–200 тыс. ₽",` +
		`"event_period":"не указан","requested_action":"выплатить","summary":"Клиент ждёт выплату."},` +
		`"draft":"Уважаемый [PERSON_1], мы получили ваше обращение."}`
	adviceJSON = `{"remarks":["Укажите срок рассмотрения.","Добавьте контакт для связи."]}`
	judgeJSON  = `{"score":4,"reason":"Полный ответ."}`
)

func TestExecute(t *testing.T) {
	t.Parallel()

	t.Run("advisor on: two gates, critic, finalize, judge, reidentify", func(t *testing.T) {
		t.Parallel()
		h := newHarness(
			t,
			[]string{
				gate2Name,
				gate2Clean,
				draftJSON,
				"Уважаемый [PERSON_1], срок рассмотрения 10 дней, контакт [PHONE_1].",
			},
			[]string{adviceJSON, judgeJSON},
		)
		run := h.run(t, vo.AdvisorOn, []string{"Иванов Пётр Сергеевич", phone})

		require.NoError(t, h.uc.Execute(t.Context(), run.ID()))

		stored, _ := h.m.Get(t.Context(), run.ID())
		assert.Equal(t, vo.RunDone, stored.Status())
		assert.Equal(t, vo.OutcomeApplied, stored.AdvisorOutcome())
		assert.Equal(t, "Я, [PERSON_1], тел. [PHONE_1], требую выплату 120 000 рублей.", stored.Anonymized())
		assert.Equal(t, 2, stored.Gate2Passes())
		assert.Equal(t, 4, stored.JudgeScore())
		assert.Equal(t, 0, stored.CanariesLeaked())
		assert.Equal(
			t,
			"Уважаемый Иванов Пётр Сергеевич, срок рассмотрения 10 дней, контакт "+phone+".",
			stored.FinalText(),
		)
		assert.Equal(
			t,
			[]string{"Укажите срок рассмотрения.", "Добавьте контакт для связи."},
			stored.Advice(),
		)
		assert.Equal(t, []string{
			"gate1:system:ok", "gate2:local:ok", "gate2:local:ok", "draft:local:ok",
			"egress:system:ok", "advise:cloud:ok", "finalize:local:ok",
			"egress:system:ok", "judge:cloud:ok", "reidentify:system:ok",
		}, h.kinds(t, run.ID()))

		for _, req := range h.cloud.requests {
			assert.NotContains(t, req.User, "Иванов")
			assert.NotContains(t, req.User, "123-45-67")
		}
		assert.Contains(t, h.cloud.requests[0].User, "Черновик ответа")
		assert.Contains(t, h.local.requests[3].User, "1. Укажите срок рассмотрения.")
	})

	t.Run("egress blocks the cloud when the draft carries a real value", func(t *testing.T) {
		t.Parallel()
		leaky := strings.Replace(draftJSON, "Уважаемый [PERSON_1]", "Уважаемый Иванов Пётр Сергеевич", 1)
		h := newHarness(t, []string{gate2Name, gate2Clean, leaky}, nil)
		run := h.run(t, vo.AdvisorOn, []string{"Иванов Пётр Сергеевич"})

		require.NoError(t, h.uc.Execute(t.Context(), run.ID()))

		stored, _ := h.m.Get(t.Context(), run.ID())
		assert.Equal(t, vo.RunDone, stored.Status())
		assert.Equal(t, vo.OutcomeBlocked, stored.AdvisorOutcome())
		assert.Equal(t, 0, stored.JudgeScore())
		assert.Empty(t, h.cloud.requests, "the strong model must never be called")
		assert.Equal(t, []string{
			"gate1:system:ok", "gate2:local:ok", "gate2:local:ok", "draft:local:ok",
			"egress:system:blocked", "advise:cloud:blocked", "finalize:system:skipped",
			"egress:system:blocked", "judge:cloud:blocked", "reidentify:system:ok",
		}, h.kinds(t, run.ID()))
	})

	t.Run("gate 2 that keeps finding things escalates", func(t *testing.T) {
		t.Parallel()
		again := `{"spans":[{"text":"рублей","type":"OTHER","canonical":"рублей"}]}`
		h := newHarness(t, []string{gate2Name, again, again}, nil)
		run := h.run(t, vo.AdvisorOn, nil)

		require.NoError(t, h.uc.Execute(t.Context(), run.ID()))

		stored, _ := h.m.Get(t.Context(), run.ID())
		assert.Equal(t, vo.RunNeedsReview, stored.Status())
		assert.True(t, stored.Escalated())
		assert.Equal(t, 2, stored.Gate2Passes(), "same count twice in a row means the model is stuck")
		assert.Equal(
			t,
			[]vo.Span{{Text: "рублей", Type: vo.PIIOther, Canonical: "рублей"}},
			stored.Findings(),
		)
		assert.Len(t, h.local.requests, 2)
	})

	t.Run("gate 2 escalates after the last pass", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, []string{
			`{"spans":[{"text":"Иванов Пётр Сергеевич","type":"PERSON","canonical":"Иванов Пётр Сергеевич"},` +
				`{"text":"рублей","type":"OTHER","canonical":"рублей"},{"text":"выплату","type":"OTHER","canonical":"выплата"}]}`,
			`{"spans":[{"text":"требую","type":"OTHER","canonical":"требую"},{"text":"тел.","type":"OTHER","canonical":"тел."}]}`,
			`{"spans":[{"text":"120 000","type":"OTHER","canonical":"120000"}]}`,
		}, nil)
		run := h.run(t, vo.AdvisorOn, nil)

		require.NoError(t, h.uc.Execute(t.Context(), run.ID()))

		stored, _ := h.m.Get(t.Context(), run.ID())
		assert.Equal(t, vo.RunNeedsReview, stored.Status())
		assert.Equal(t, 3, stored.Gate2Passes())
	})

	t.Run("a failed gate 2 pass is not a clean one", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, []string{"", gate2Clean, draftJSON}, []string{judgeJSON})
		h.local.errs = []error{errors.New("timeout")}
		run := h.run(t, vo.AdvisorOff, nil)

		require.NoError(t, h.uc.Execute(t.Context(), run.ID()))

		stored, _ := h.m.Get(t.Context(), run.ID())
		assert.Equal(t, vo.RunDone, stored.Status())
		assert.Equal(t, 2, stored.Gate2Passes())
		assert.Equal(t, "gate2:local:failed", h.kinds(t, run.ID())[1])
	})

	t.Run("placebo advises without the cloud and still gets judged", func(t *testing.T) {
		t.Parallel()
		h := newHarness(
			t,
			[]string{gate2Clean, draftJSON, "Итоговый текст для [PERSON_1]."},
			[]string{judgeJSON},
		)
		run := h.run(t, vo.AdvisorPlacebo, nil)

		require.NoError(t, h.uc.Execute(t.Context(), run.ID()))

		stored, _ := h.m.Get(t.Context(), run.ID())
		assert.Equal(t, vo.OutcomePlacebo, stored.AdvisorOutcome())
		assert.Len(t, stored.Advice(), 3)
		assert.Equal(t, 4, stored.JudgeScore())
		assert.Len(t, h.cloud.requests, 1, "only the judge talks to the cloud")
		assert.Contains(t, h.kinds(t, run.ID()), "advise:placebo:ok")
	})

	t.Run("advisor off skips advice and finalize but keeps the judge", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, []string{gate2Clean, draftJSON}, []string{judgeJSON})
		run := h.run(t, vo.AdvisorOff, nil)

		require.NoError(t, h.uc.Execute(t.Context(), run.ID()))

		stored, _ := h.m.Get(t.Context(), run.ID())
		assert.Equal(t, vo.OutcomeSkipped, stored.AdvisorOutcome())
		assert.Equal(t, stored.Draft(), stored.FinalAnonymous())
		assert.Equal(t, []string{
			"gate1:system:ok", "gate2:local:ok", "draft:local:ok", "advise:system:skipped",
			"finalize:system:skipped", "egress:system:ok", "judge:cloud:ok", "reidentify:system:ok",
		}, h.kinds(t, run.ID()))
	})

	t.Run("a model failure fails the run with the step recorded", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, []string{gate2Clean, ""}, nil)
		h.local.errs = []error{nil, errors.New("502 bad gateway")}
		run := h.run(t, vo.AdvisorOn, nil)

		require.NoError(t, h.uc.Execute(t.Context(), run.ID()))

		stored, _ := h.m.Get(t.Context(), run.ID())
		assert.Equal(t, vo.RunFailed, stored.Status())
		assert.Contains(t, stored.Error(), "draft: 502 bad gateway")
		assert.Equal(t, "draft:local:failed", h.kinds(t, run.ID())[2])
	})

	t.Run("resumes after the operator finishes without the adviser", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, []string{draftJSON}, nil)
		run := h.run(t, vo.AdvisorOn, nil)
		require.NoError(t, run.Start(h.uc.now()))
		vault := vo.NewVault()
		text := vo.ApplySpans(letter, stubDetector{}.Detect(letter), vault)
		run.Escalate(text, vault, []vo.Span{{Text: "x", Type: vo.PIIOther}}, 3, h.uc.now())
		require.NoError(t, run.Resolve(vo.ReviewFinishWithoutAdvisor, h.uc.now()))
		require.NoError(t, h.m.Update(t.Context(), run))

		require.NoError(t, h.uc.Execute(t.Context(), run.ID()))

		stored, _ := h.m.Get(t.Context(), run.ID())
		assert.Equal(t, vo.RunDone, stored.Status())
		assert.Equal(t, vo.OutcomeSkipped, stored.AdvisorOutcome())
		assert.Equal(t, 0, stored.JudgeScore())
		assert.Empty(t, h.cloud.requests)
		assert.Equal(t, []string{
			"draft:local:ok", "advise:system:skipped", "finalize:system:skipped", "reidentify:system:ok",
		}, h.kinds(t, run.ID()))
	})

	t.Run("a finished run advances its benchmark", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t, []string{gate2Clean, draftJSON}, []string{judgeJSON})
		benchmark, err := entity.NewBenchmark(1, h.uc.now())
		require.NoError(t, err)
		require.NoError(t, benchStore{h.m}.Create(t.Context(), benchmark))
		run, err := entity.NewRun(entity.NewRunInput{
			Source: vo.SourceText, AdvisorMode: vo.AdvisorOff, Text: letter, BenchmarkID: benchmark.ID(),
		}, h.uc.now())
		require.NoError(t, err)
		require.NoError(t, h.m.Create(t.Context(), run))

		require.NoError(t, h.uc.Execute(t.Context(), run.ID()))

		stored, _ := benchStore{h.m}.Get(t.Context(), benchmark.ID())
		assert.Equal(t, entity.BenchmarkDone, stored.Status())
		assert.Equal(t, 1, stored.Finished())
	})
}

func TestPlaceboIsStablePerRun(t *testing.T) {
	t.Parallel()

	id := uuid.NewV7()
	assert.Equal(t, placebo(id), placebo(id))
	assert.Len(t, placebo(id), 3)
}

func TestDecodeJSON(t *testing.T) {
	t.Parallel()

	var out judgeResponse
	require.NoError(t, decodeJSON("```json\n{\"score\": 5, \"reason\": \"ok\"}\n```", &out))
	assert.Equal(t, 5, out.Score)
	assert.Error(t, decodeJSON("no json here", &out))
}
