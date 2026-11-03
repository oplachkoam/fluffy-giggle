package entity

import (
	"encoding/json"
	"errors"
	"time"
	"uuid"

	"fluffy-giggle/core/internal/domain/vo"
)

// Step is one recorded action of the pipeline, with the exact input and output, so
// the operator can see what each side of the boundary saw.
type Step struct {
	id               uuid.UUID
	runID            uuid.UUID
	seq              int
	kind             vo.StepKind
	side             vo.Side
	status           vo.StepStatus
	startedAt        time.Time
	finishedAt       time.Time
	model            string
	promptTokens     int
	completionTokens int
	input            json.RawMessage
	output           json.RawMessage
	err              string
}

type StepState struct {
	ID               uuid.UUID
	RunID            uuid.UUID
	Seq              int
	Kind             vo.StepKind
	Side             vo.Side
	Status           vo.StepStatus
	StartedAt        time.Time
	FinishedAt       time.Time
	Model            string
	PromptTokens     int
	CompletionTokens int
	Input            json.RawMessage
	Output           json.RawMessage
	Error            string
}

func NewStep(state *StepState) (*Step, error) {
	copied := *state
	if copied.ID == uuid.Nil() {
		copied.ID = uuid.NewV7()
	}
	return RestoreStep(&copied)
}

func RestoreStep(state *StepState) (*Step, error) {
	if state.ID == uuid.Nil() {
		return nil, errors.New("step id is empty")
	}
	if state.RunID == uuid.Nil() {
		return nil, errors.New("step run id is empty")
	}
	if state.Seq < 1 {
		return nil, errors.New("step seq starts at 1")
	}
	if _, err := vo.ParseStepKind(string(state.Kind)); err != nil {
		return nil, err
	}
	if _, err := vo.ParseSide(string(state.Side)); err != nil {
		return nil, err
	}
	if _, err := vo.ParseStepStatus(string(state.Status)); err != nil {
		return nil, err
	}
	if state.StartedAt.IsZero() {
		return nil, errors.New("step start time is empty")
	}
	finishedAt := state.FinishedAt
	if finishedAt.IsZero() {
		finishedAt = state.StartedAt
	}

	return &Step{
		id:               state.ID,
		runID:            state.RunID,
		seq:              state.Seq,
		kind:             state.Kind,
		side:             state.Side,
		status:           state.Status,
		startedAt:        state.StartedAt.UTC(),
		finishedAt:       finishedAt.UTC(),
		model:            state.Model,
		promptTokens:     state.PromptTokens,
		completionTokens: state.CompletionTokens,
		input:            orEmptyObject(state.Input),
		output:           orEmptyObject(state.Output),
		err:              state.Error,
	}, nil
}

func orEmptyObject(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage("{}")
	}
	return json.RawMessage(append([]byte(nil), raw...))
}

func (s *Step) ID() uuid.UUID           { return s.id }
func (s *Step) RunID() uuid.UUID        { return s.runID }
func (s *Step) Seq() int                { return s.seq }
func (s *Step) Kind() vo.StepKind       { return s.kind }
func (s *Step) Side() vo.Side           { return s.side }
func (s *Step) Status() vo.StepStatus   { return s.status }
func (s *Step) StartedAt() time.Time    { return s.startedAt }
func (s *Step) FinishedAt() time.Time   { return s.finishedAt }
func (s *Step) Model() string           { return s.model }
func (s *Step) PromptTokens() int       { return s.promptTokens }
func (s *Step) CompletionTokens() int   { return s.completionTokens }
func (s *Step) Input() json.RawMessage  { return orEmptyObject(s.input) }
func (s *Step) Output() json.RawMessage { return orEmptyObject(s.output) }
func (s *Step) Error() string           { return s.err }
func (s *Step) Duration() time.Duration { return s.finishedAt.Sub(s.startedAt) }
