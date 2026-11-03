package vo

import "strings"

type PIIType string

const (
	PIIPerson   PIIType = "PERSON"
	PIIPhone    PIIType = "PHONE"
	PIIEmail    PIIType = "EMAIL"
	PIIPassport PIIType = "PASSPORT"
	PIISNILS    PIIType = "SNILS"
	PIIINN      PIIType = "INN"
	PIICard     PIIType = "CARD"
	PIIAccount  PIIType = "ACCOUNT"
	PIIAddress  PIIType = "ADDRESS"
	PIIDate     PIIType = "DATE"
	PIIOrg      PIIType = "ORG"
	PIIOther    PIIType = "OTHER"
)

var piiTypes = []PIIType{
	PIIPerson, PIIPhone, PIIEmail, PIIPassport, PIISNILS, PIIINN,
	PIICard, PIIAccount, PIIAddress, PIIDate, PIIOrg, PIIOther,
}

// ParsePIIType is lenient: a model may invent a label, and an unknown label still
// marks something to cut, so it becomes OTHER.
func ParsePIIType(s string) PIIType {
	upper := PIIType(strings.ToUpper(strings.TrimSpace(s)))
	for _, t := range piiTypes {
		if t == upper {
			return t
		}
	}
	return PIIOther
}

func (t PIIType) Numeric() bool {
	switch t {
	case PIIPhone, PIIPassport, PIISNILS, PIIINN, PIICard, PIIAccount:
		return true
	default:
		return false
	}
}

// Span is one found value. Canonical is the form the vault keys on: digits only for
// numbers, nominative case for names; Text is how it appeared.
type Span struct {
	Text      string  `json:"text"`
	Type      PIIType `json:"type"`
	Canonical string  `json:"canonical"`
}
