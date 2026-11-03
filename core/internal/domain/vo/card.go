package vo

// Card is the generalized view of a claim: enough for the adviser to judge the
// draft, nothing that names the client.
type Card struct {
	Category        string `json:"category"`
	Subject         string `json:"subject"`
	AmountRange     string `json:"amount_range"`
	EventPeriod     string `json:"event_period"`
	RequestedAction string `json:"requested_action"`
	Summary         string `json:"summary"`
}

type Judgement struct {
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

// Finding is what the egress check caught in a payload bound for the cloud.
type Finding struct {
	Type  PIIType `json:"type"`
	Value string  `json:"value"`
}
