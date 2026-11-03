package entity

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"uuid"

	"fluffy-giggle/core/internal/domain/vo"
)

const MaxImageBytes = 10 << 20

// Run is one pass of the pipeline over one input. Everything the pipeline learns
// lives here, so a run interrupted for review resumes from the stored state.
type Run struct {
	id             uuid.UUID
	createdAt      time.Time
	updatedAt      time.Time
	status         vo.RunStatus
	phase          vo.RunPhase
	advisorMode    vo.AdvisorMode
	advisorOutcome vo.AdvisorOutcome
	reviewDecision vo.ReviewDecision
	source         vo.Source
	caseID         string
	benchmarkID    uuid.UUID
	inputText      string
	image          []byte
	imageSize      int
	imageMIME      string
	canaries       []string
	anonymized     string
	vault          *vo.Vault
	findings       []vo.Span
	card           *vo.Card
	draft          string
	advice         []string
	finalAnonymous string
	finalText      string
	judgeScore     int
	judgeReason    string
	gate2Passes    int
	escalated      bool
	canariesLeaked int
	err            string
}

type RunState struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Status         vo.RunStatus
	Phase          vo.RunPhase
	AdvisorMode    vo.AdvisorMode
	AdvisorOutcome vo.AdvisorOutcome
	ReviewDecision vo.ReviewDecision
	Source         vo.Source
	CaseID         string
	BenchmarkID    uuid.UUID
	InputText      string
	Image          []byte
	// ImageSize stands in for Image when a listing leaves the bytes out.
	ImageSize      int
	ImageMIME      string
	Canaries       []string
	Anonymized     string
	Vault          []vo.VaultEntry
	Findings       []vo.Span
	Card           *vo.Card
	Draft          string
	Advice         []string
	FinalAnonymous string
	FinalText      string
	// JudgeScore is 0 while there is no score.
	JudgeScore     int
	JudgeReason    string
	Gate2Passes    int
	Escalated      bool
	CanariesLeaked int
	Error          string
}

type NewRunInput struct {
	Source      vo.Source
	AdvisorMode vo.AdvisorMode
	Text        string
	Image       []byte
	ImageMIME   string
	CaseID      string
	Canaries    []string
	BenchmarkID uuid.UUID
}

func NewRun(in NewRunInput, now time.Time) (*Run, error) {
	return RestoreRun(&RunState{
		ID:             uuid.NewV7(),
		CreatedAt:      now,
		UpdatedAt:      now,
		Status:         vo.RunQueued,
		Phase:          vo.PhaseAnonymize,
		AdvisorMode:    in.AdvisorMode,
		AdvisorOutcome: vo.OutcomePending,
		Source:         in.Source,
		CaseID:         in.CaseID,
		BenchmarkID:    in.BenchmarkID,
		InputText:      in.Text,
		Image:          in.Image,
		ImageMIME:      in.ImageMIME,
		Canaries:       in.Canaries,
	})
}

func RestoreRun(state *RunState) (*Run, error) {
	if state.ID == uuid.Nil() {
		return nil, errors.New("run id is empty")
	}
	if state.CreatedAt.IsZero() {
		return nil, errors.New("run creation time is empty")
	}
	if _, err := vo.ParseRunStatus(string(state.Status)); err != nil {
		return nil, err
	}
	if _, err := vo.ParseRunPhase(string(state.Phase)); err != nil {
		return nil, err
	}
	if _, err := vo.ParseAdvisorMode(string(state.AdvisorMode)); err != nil {
		return nil, err
	}
	if _, err := vo.ParseAdvisorOutcome(string(state.AdvisorOutcome)); err != nil {
		return nil, err
	}
	if _, err := vo.ParseStoredReviewDecision(string(state.ReviewDecision)); err != nil {
		return nil, err
	}
	if _, err := vo.ParseSource(string(state.Source)); err != nil {
		return nil, err
	}
	switch state.Source {
	case vo.SourceText:
		if strings.TrimSpace(state.InputText) == "" {
			return nil, errors.New("run text is empty")
		}
	case vo.SourceImage:
		if len(state.Image) == 0 && state.ImageSize == 0 {
			return nil, errors.New("run image is empty")
		}
		if len(state.Image) > MaxImageBytes {
			return nil, fmt.Errorf("run image is larger than %d bytes", MaxImageBytes)
		}
		if !vo.IsImageMIME(state.ImageMIME) {
			return nil, fmt.Errorf("run image type %q is not supported", state.ImageMIME)
		}
	}
	if state.JudgeScore < 0 || state.JudgeScore > 5 {
		return nil, errors.New("judge score is out of range")
	}
	updatedAt := state.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = state.CreatedAt
	}

	return &Run{
		id:             state.ID,
		createdAt:      state.CreatedAt.UTC(),
		updatedAt:      updatedAt.UTC(),
		status:         state.Status,
		phase:          state.Phase,
		advisorMode:    state.AdvisorMode,
		advisorOutcome: state.AdvisorOutcome,
		reviewDecision: state.ReviewDecision,
		source:         state.Source,
		caseID:         state.CaseID,
		benchmarkID:    state.BenchmarkID,
		inputText:      state.InputText,
		image:          slices.Clone(state.Image),
		imageSize:      max(state.ImageSize, len(state.Image)),
		imageMIME:      state.ImageMIME,
		canaries:       slices.Clone(state.Canaries),
		anonymized:     state.Anonymized,
		vault:          vo.RestoreVault(state.Vault),
		findings:       slices.Clone(state.Findings),
		card:           cloneCard(state.Card),
		draft:          state.Draft,
		advice:         slices.Clone(state.Advice),
		finalAnonymous: state.FinalAnonymous,
		finalText:      state.FinalText,
		judgeScore:     state.JudgeScore,
		judgeReason:    state.JudgeReason,
		gate2Passes:    state.Gate2Passes,
		escalated:      state.Escalated,
		canariesLeaked: state.CanariesLeaked,
		err:            state.Error,
	}, nil
}

func cloneCard(card *vo.Card) *vo.Card {
	if card == nil {
		return nil
	}
	copied := *card
	return &copied
}

func (r *Run) ID() uuid.UUID                     { return r.id }
func (r *Run) CreatedAt() time.Time              { return r.createdAt }
func (r *Run) UpdatedAt() time.Time              { return r.updatedAt }
func (r *Run) Status() vo.RunStatus              { return r.status }
func (r *Run) Phase() vo.RunPhase                { return r.phase }
func (r *Run) AdvisorMode() vo.AdvisorMode       { return r.advisorMode }
func (r *Run) AdvisorOutcome() vo.AdvisorOutcome { return r.advisorOutcome }
func (r *Run) ReviewDecision() vo.ReviewDecision { return r.reviewDecision }
func (r *Run) Source() vo.Source                 { return r.source }
func (r *Run) CaseID() string                    { return r.caseID }
func (r *Run) BenchmarkID() uuid.UUID            { return r.benchmarkID }
func (r *Run) InputText() string                 { return r.inputText }
func (r *Run) Image() []byte                     { return slices.Clone(r.image) }
func (r *Run) ImageMIME() string                 { return r.imageMIME }
func (r *Run) ImageSize() int                    { return r.imageSize }
func (r *Run) HasImage() bool                    { return r.imageSize > 0 }
func (r *Run) Canaries() []string                { return slices.Clone(r.canaries) }
func (r *Run) Anonymized() string                { return r.anonymized }
func (r *Run) Vault() *vo.Vault                  { return vo.RestoreVault(r.vault.Entries()) }
func (r *Run) Findings() []vo.Span               { return slices.Clone(r.findings) }
func (r *Run) Card() *vo.Card                    { return cloneCard(r.card) }
func (r *Run) Draft() string                     { return r.draft }
func (r *Run) Advice() []string                  { return slices.Clone(r.advice) }
func (r *Run) FinalAnonymous() string            { return r.finalAnonymous }
func (r *Run) FinalText() string                 { return r.finalText }
func (r *Run) JudgeScore() int                   { return r.judgeScore }
func (r *Run) JudgeReason() string               { return r.judgeReason }
func (r *Run) Gate2Passes() int                  { return r.gate2Passes }
func (r *Run) Escalated() bool                   { return r.escalated }
func (r *Run) CanariesLeaked() int               { return r.canariesLeaked }
func (r *Run) Error() string                     { return r.err }

// AdvisorSkippedByReview is the operator's choice to finish without the cloud.
func (r *Run) AdvisorSkippedByReview() bool {
	return r.reviewDecision == vo.ReviewFinishWithoutAdvisor
}

func (r *Run) Start(now time.Time) error {
	if r.status != vo.RunQueued {
		return fmt.Errorf("run is %s, not queued", r.status)
	}
	r.status = vo.RunRunning
	r.touch(now)
	return nil
}

// SetTranscript replaces the text of an image run with what the vision model read.
func (r *Run) SetTranscript(text string, now time.Time) {
	r.inputText = text
	r.touch(now)
}

// SetAnonymized closes anonymization: the text is clean, composition can start.
func (r *Run) SetAnonymized(text string, vault *vo.Vault, passes int, now time.Time) {
	r.anonymized = text
	r.vault = vo.RestoreVault(vault.Entries())
	r.gate2Passes = passes
	r.findings = nil
	r.phase = vo.PhaseCompose
	r.touch(now)
}

// Escalate parks the run for the operator with what gate 2 still finds.
func (r *Run) Escalate(text string, vault *vo.Vault, findings []vo.Span, passes int, now time.Time) {
	r.anonymized = text
	r.vault = vo.RestoreVault(vault.Entries())
	r.findings = slices.Clone(findings)
	r.gate2Passes = passes
	r.escalated = true
	r.status = vo.RunNeedsReview
	r.touch(now)
}

func (r *Run) Resolve(decision vo.ReviewDecision, now time.Time) error {
	if r.status != vo.RunNeedsReview {
		return fmt.Errorf("run is %s, not awaiting review", r.status)
	}
	if _, err := vo.ParseReviewDecision(string(decision)); err != nil {
		return err
	}
	r.reviewDecision = decision
	r.status = vo.RunRunning
	r.phase = vo.PhaseCompose
	r.findings = nil
	r.touch(now)
	return nil
}

func (r *Run) SetDraft(card vo.Card, draft string, now time.Time) {
	r.card = &card
	r.draft = draft
	r.touch(now)
}

func (r *Run) SetAdvice(advice []string, outcome vo.AdvisorOutcome, now time.Time) {
	r.advice = slices.Clone(advice)
	r.advisorOutcome = outcome
	r.touch(now)
}

func (r *Run) SetFinal(anonymous, text string, now time.Time) {
	r.finalAnonymous = anonymous
	r.finalText = text
	r.touch(now)
}

func (r *Run) SetJudgement(j vo.Judgement, now time.Time) error {
	if j.Score < 1 || j.Score > 5 {
		return fmt.Errorf("judge score %d is out of 1..5", j.Score)
	}
	r.judgeScore = j.Score
	r.judgeReason = j.Reason
	r.touch(now)
	return nil
}

func (r *Run) Finish(canariesLeaked int, now time.Time) error {
	if r.status != vo.RunRunning {
		return fmt.Errorf("run is %s, not running", r.status)
	}
	r.canariesLeaked = canariesLeaked
	r.status = vo.RunDone
	r.phase = vo.PhaseFinished
	r.touch(now)
	return nil
}

func (r *Run) Fail(reason string, now time.Time) {
	r.err = reason
	r.status = vo.RunFailed
	r.phase = vo.PhaseFinished
	r.touch(now)
}

func (r *Run) touch(now time.Time) {
	r.updatedAt = now.UTC()
}
