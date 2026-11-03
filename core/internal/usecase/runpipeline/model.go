package runpipeline

import "fluffy-giggle/core/internal/domain/vo"

type Config struct {
	Gate2MaxPasses int
}

type gate2Response struct {
	Spans []gate2Span `json:"spans"`
}

type gate2Span struct {
	Text      string `json:"text"`
	Type      string `json:"type"`
	Canonical string `json:"canonical"`
}

type draftResponse struct {
	Card  vo.Card `json:"card"`
	Draft string  `json:"draft"`
}

type adviseResponse struct {
	Remarks []string `json:"remarks"`
}

type judgeResponse struct {
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

// promptPayload is what a model step records as its input: the exact messages sent.
type promptPayload struct {
	System string `json:"system"`
	User   string `json:"user"`
	Image  string `json:"image,omitempty"`
}
