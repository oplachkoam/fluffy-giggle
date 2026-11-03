package vo

import "fmt"

type StepKind string

const (
	StepIngest     StepKind = "ingest"
	StepVision     StepKind = "vision"
	StepGate1      StepKind = "gate1"
	StepGate2      StepKind = "gate2"
	StepReview     StepKind = "review"
	StepDraft      StepKind = "draft"
	StepEgress     StepKind = "egress"
	StepAdvise     StepKind = "advise"
	StepFinalize   StepKind = "finalize"
	StepJudge      StepKind = "judge"
	StepReidentify StepKind = "reidentify"
)

func ParseStepKind(s string) (StepKind, error) {
	switch StepKind(s) {
	case StepIngest, StepVision, StepGate1, StepGate2, StepReview, StepDraft, StepEgress,
		StepAdvise, StepFinalize, StepJudge, StepReidentify:
		return StepKind(s), nil
	}
	return "", fmt.Errorf("unknown step kind %q", s)
}

// Side is who saw the step's data: local models and code never leave the perimeter,
// cloud is the strong model, placebo is advice no model produced.
type Side string

const (
	SideLocal   Side = "local"
	SideCloud   Side = "cloud"
	SideSystem  Side = "system"
	SidePlacebo Side = "placebo"
)

func ParseSide(s string) (Side, error) {
	switch Side(s) {
	case SideLocal, SideCloud, SideSystem, SidePlacebo:
		return Side(s), nil
	}
	return "", fmt.Errorf("unknown side %q", s)
}

type StepStatus string

const (
	StepOK      StepStatus = "ok"
	StepFailed  StepStatus = "failed"
	StepBlocked StepStatus = "blocked"
	StepSkipped StepStatus = "skipped"
)

func ParseStepStatus(s string) (StepStatus, error) {
	switch StepStatus(s) {
	case StepOK, StepFailed, StepBlocked, StepSkipped:
		return StepStatus(s), nil
	}
	return "", fmt.Errorf("unknown step status %q", s)
}
