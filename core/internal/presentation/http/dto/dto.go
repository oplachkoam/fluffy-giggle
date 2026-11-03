// Package dto is the JSON the API speaks, shared by handlers and the event stream.
package dto

import (
	"encoding/json"
	"time"
	"uuid"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/vo"
	"fluffy-giggle/core/internal/usecase/getbenchmark"
)

type Run struct {
	ID              string          `json:"id"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	Status          string          `json:"status"`
	Phase           string          `json:"phase"`
	AdvisorMode     string          `json:"advisor_mode"`
	AdvisorOutcome  string          `json:"advisor_outcome"`
	ReviewDecision  string          `json:"review_decision"`
	Source          string          `json:"source"`
	CaseID          string          `json:"case_id"`
	BenchmarkID     string          `json:"benchmark_id"`
	HasImage        bool            `json:"has_image"`
	InputText       string          `json:"input_text"`
	AnonymizedText  string          `json:"anonymized_text"`
	Vault           []vo.VaultEntry `json:"vault"`
	ReviewFindings  []vo.Span       `json:"review_findings"`
	Card            *vo.Card        `json:"card"`
	Draft           string          `json:"draft"`
	Advice          []string        `json:"advice"`
	FinalAnonymized string          `json:"final_anonymized"`
	FinalText       string          `json:"final_text"`
	JudgeScore      *int            `json:"judge_score"`
	JudgeReason     string          `json:"judge_reason"`
	Gate2Passes     int             `json:"gate2_passes"`
	Escalated       bool            `json:"escalated"`
	CanariesTotal   int             `json:"canaries_total"`
	CanariesLeaked  int             `json:"canaries_leaked"`
	Error           string          `json:"error"`
	Steps           []Step          `json:"steps,omitempty"`
}

type Step struct {
	ID               string          `json:"id"`
	RunID            string          `json:"run_id"`
	Seq              int             `json:"seq"`
	Kind             string          `json:"kind"`
	Side             string          `json:"side"`
	Status           string          `json:"status"`
	StartedAt        time.Time       `json:"started_at"`
	FinishedAt       time.Time       `json:"finished_at"`
	DurationMs       int64           `json:"duration_ms"`
	Model            string          `json:"model"`
	PromptTokens     int             `json:"prompt_tokens"`
	CompletionTokens int             `json:"completion_tokens"`
	Input            json.RawMessage `json:"input"`
	Output           json.RawMessage `json:"output"`
	Error            string          `json:"error"`
}

type Benchmark struct {
	ID        string        `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	Status    string        `json:"status"`
	Total     int           `json:"total"`
	Finished  int           `json:"finished"`
	Summary   []ModeSummary `json:"summary,omitempty"`
	Runs      []Run         `json:"runs,omitempty"`
}

type ModeSummary struct {
	Mode      string   `json:"mode"`
	Runs      int      `json:"runs"`
	Done      int      `json:"done"`
	Failed    int      `json:"failed"`
	Escalated int      `json:"escalated"`
	Blocked   int      `json:"blocked"`
	Scored    int      `json:"scored"`
	AvgScore  *float64 `json:"avg_score"`
	Leaked    int      `json:"leaked"`
	Gate2Avg  float64  `json:"gate2_avg"`
}

type Case struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	HasImage bool   `json:"has_image"`
	Canaries int    `json:"canaries"`
}

func FromRun(run *entity.Run, steps []*entity.Step) Run {
	out := Run{
		ID:              run.ID().String(),
		CreatedAt:       run.CreatedAt(),
		UpdatedAt:       run.UpdatedAt(),
		Status:          string(run.Status()),
		Phase:           string(run.Phase()),
		AdvisorMode:     string(run.AdvisorMode()),
		AdvisorOutcome:  string(run.AdvisorOutcome()),
		ReviewDecision:  string(run.ReviewDecision()),
		Source:          string(run.Source()),
		CaseID:          run.CaseID(),
		BenchmarkID:     idOrEmpty(run.BenchmarkID()),
		HasImage:        run.HasImage(),
		InputText:       run.InputText(),
		AnonymizedText:  run.Anonymized(),
		Vault:           orEmpty(run.Vault().Entries()),
		ReviewFindings:  orEmpty(run.Findings()),
		Card:            run.Card(),
		Draft:           run.Draft(),
		Advice:          orEmpty(run.Advice()),
		FinalAnonymized: run.FinalAnonymous(),
		FinalText:       run.FinalText(),
		JudgeReason:     run.JudgeReason(),
		Gate2Passes:     run.Gate2Passes(),
		Escalated:       run.Escalated(),
		CanariesTotal:   len(run.Canaries()),
		CanariesLeaked:  run.CanariesLeaked(),
		Error:           run.Error(),
	}
	if score := run.JudgeScore(); score > 0 {
		out.JudgeScore = &score
	}
	for _, step := range steps {
		out.Steps = append(out.Steps, FromStep(step))
	}
	return out
}

func FromStep(step *entity.Step) Step {
	return Step{
		ID:               step.ID().String(),
		RunID:            step.RunID().String(),
		Seq:              step.Seq(),
		Kind:             string(step.Kind()),
		Side:             string(step.Side()),
		Status:           string(step.Status()),
		StartedAt:        step.StartedAt(),
		FinishedAt:       step.FinishedAt(),
		DurationMs:       step.Duration().Milliseconds(),
		Model:            step.Model(),
		PromptTokens:     step.PromptTokens(),
		CompletionTokens: step.CompletionTokens(),
		Input:            step.Input(),
		Output:           step.Output(),
		Error:            step.Error(),
	}
}

func FromBenchmark(b *entity.Benchmark, summary []getbenchmark.ModeSummary, runs []*entity.Run) Benchmark {
	out := Benchmark{
		ID:        b.ID().String(),
		CreatedAt: b.CreatedAt(),
		Status:    string(b.Status()),
		Total:     b.Total(),
		Finished:  b.Finished(),
	}
	for _, s := range summary {
		item := ModeSummary{
			Mode: string(s.Mode), Runs: s.Runs, Done: s.Done, Failed: s.Failed, Escalated: s.Escalated,
			Blocked: s.Blocked, Scored: s.Scored, Leaked: s.Leaked,
		}
		if s.Scored > 0 {
			avg := s.AvgScore
			item.AvgScore = &avg
		}
		if s.Runs > 0 {
			item.Gate2Avg = float64(s.Gate2Total) / float64(s.Runs)
		}
		out.Summary = append(out.Summary, item)
	}
	for _, run := range runs {
		out.Runs = append(out.Runs, FromRun(run, nil))
	}
	return out
}

func FromCase(c *entity.Case) Case {
	return Case{ID: c.ID(), Title: c.Title(), HasImage: c.HasImage(), Canaries: len(c.Canaries())}
}

func idOrEmpty(id uuid.UUID) string {
	if id == uuid.Nil() {
		return ""
	}
	return id.String()
}

func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
