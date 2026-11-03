package vo

import "fmt"

type RunStatus string

const (
	RunQueued      RunStatus = "queued"
	RunRunning     RunStatus = "running"
	RunNeedsReview RunStatus = "needs_review"
	RunDone        RunStatus = "done"
	RunFailed      RunStatus = "failed"
)

func ParseRunStatus(s string) (RunStatus, error) {
	switch RunStatus(s) {
	case RunQueued, RunRunning, RunNeedsReview, RunDone, RunFailed:
		return RunStatus(s), nil
	}
	return "", fmt.Errorf("unknown run status %q", s)
}

// RunPhase is where the pipeline resumes: anonymization can stop for a review,
// composition runs to the end.
type RunPhase string

const (
	PhaseAnonymize RunPhase = "anonymize"
	PhaseCompose   RunPhase = "compose"
	PhaseFinished  RunPhase = "finished"
)

func ParseRunPhase(s string) (RunPhase, error) {
	switch RunPhase(s) {
	case PhaseAnonymize, PhaseCompose, PhaseFinished:
		return RunPhase(s), nil
	}
	return "", fmt.Errorf("unknown run phase %q", s)
}

type AdvisorMode string

const (
	AdvisorOff     AdvisorMode = "off"
	AdvisorOn      AdvisorMode = "on"
	AdvisorPlacebo AdvisorMode = "placebo"
)

func ParseAdvisorMode(s string) (AdvisorMode, error) {
	switch AdvisorMode(s) {
	case AdvisorOff, AdvisorOn, AdvisorPlacebo:
		return AdvisorMode(s), nil
	}
	return "", fmt.Errorf("unknown advisor mode %q", s)
}

// AdvisorOutcome says what happened to the advice: pending until the run gets
// there, blocked when egress stopped the call, skipped by mode or by the operator.
type AdvisorOutcome string

const (
	OutcomePending AdvisorOutcome = "pending"
	OutcomeApplied AdvisorOutcome = "applied"
	OutcomePlacebo AdvisorOutcome = "placebo"
	OutcomeSkipped AdvisorOutcome = "skipped"
	OutcomeBlocked AdvisorOutcome = "blocked"
)

func ParseAdvisorOutcome(s string) (AdvisorOutcome, error) {
	switch AdvisorOutcome(s) {
	case OutcomePending, OutcomeApplied, OutcomePlacebo, OutcomeSkipped, OutcomeBlocked:
		return AdvisorOutcome(s), nil
	}
	return "", fmt.Errorf("unknown advisor outcome %q", s)
}

type Source string

const (
	SourceText  Source = "text"
	SourceImage Source = "image"
)

func ParseSource(s string) (Source, error) {
	switch Source(s) {
	case SourceText, SourceImage:
		return Source(s), nil
	}
	return "", fmt.Errorf("unknown source %q", s)
}

type ReviewDecision string

const (
	ReviewNone                 ReviewDecision = ""
	ReviewContinue             ReviewDecision = "continue"
	ReviewFinishWithoutAdvisor ReviewDecision = "finish_without_advisor"
)

func ParseReviewDecision(s string) (ReviewDecision, error) {
	switch ReviewDecision(s) {
	case ReviewContinue, ReviewFinishWithoutAdvisor:
		return ReviewDecision(s), nil
	default:
		return "", fmt.Errorf("unknown review decision %q", s)
	}
}

func ParseStoredReviewDecision(s string) (ReviewDecision, error) {
	if s == "" {
		return ReviewNone, nil
	}
	return ParseReviewDecision(s)
}

var ImageMIMETypes = []string{"image/png", "image/jpeg"}

func IsImageMIME(mime string) bool {
	for _, allowed := range ImageMIMETypes {
		if allowed == mime {
			return true
		}
	}
	return false
}
