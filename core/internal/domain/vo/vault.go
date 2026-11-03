package vo

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// VaultEntry maps one real value to its placeholder together with every form it
// was seen in, so egress can look for all of them.
type VaultEntry struct {
	Placeholder string   `json:"placeholder"`
	Type        PIIType  `json:"type"`
	Canonical   string   `json:"canonical"`
	Forms       []string `json:"forms"`
}

// Vault lives inside one run and never leaves the perimeter.
type Vault struct {
	entries  []*VaultEntry
	counters map[PIIType]int
}

func NewVault() *Vault {
	return &Vault{counters: map[PIIType]int{}}
}

func RestoreVault(entries []VaultEntry) *Vault {
	v := NewVault()
	for _, entry := range entries {
		copied := entry
		copied.Forms = slices.Clone(entry.Forms)
		v.entries = append(v.entries, &copied)
		var n int
		if _, err := fmt.Sscanf(entry.Placeholder, "["+string(entry.Type)+"_%d]", &n); err == nil {
			v.counters[entry.Type] = max(v.counters[entry.Type], n)
		}
	}
	return v
}

// Assign returns the placeholder for canonical, creating one on first sight, and
// remembers form as another spelling of it.
func (v *Vault) Assign(t PIIType, canonical, form string) string {
	canonical = strings.TrimSpace(canonical)
	if canonical == "" {
		canonical = strings.TrimSpace(form)
	}
	key := normalizeValue(t, canonical)
	for _, entry := range v.entries {
		if entry.Type == t && normalizeValue(t, entry.Canonical) == key {
			entry.addForm(form)
			return entry.Placeholder
		}
	}

	v.counters[t]++
	entry := &VaultEntry{
		Placeholder: fmt.Sprintf("[%s_%d]", t, v.counters[t]),
		Type:        t,
		Canonical:   canonical,
	}
	entry.addForm(form)
	entry.addForm(canonical)
	v.entries = append(v.entries, entry)

	return entry.Placeholder
}

func (e *VaultEntry) addForm(form string) {
	form = strings.TrimSpace(form)
	if form == "" || slices.Contains(e.Forms, form) {
		return
	}
	e.Forms = append(e.Forms, form)
}

func (v *Vault) Entries() []VaultEntry {
	out := make([]VaultEntry, 0, len(v.entries))
	for _, entry := range v.entries {
		copied := *entry
		copied.Forms = slices.Clone(entry.Forms)
		out = append(out, copied)
	}
	return out
}

func (v *Vault) Len() int {
	return len(v.entries)
}

func (v *Vault) Placeholders() []string {
	out := make([]string, 0, len(v.entries))
	for _, entry := range v.entries {
		out = append(out, entry.Placeholder)
	}
	return out
}

// Forms lists every spelling the vault knows, the material for the egress check.
func (v *Vault) Forms() []string {
	var out []string
	for _, entry := range v.entries {
		for _, form := range entry.Forms {
			if !slices.Contains(out, form) {
				out = append(out, form)
			}
		}
	}
	return out
}

// Reidentify puts real values back: names in the nominative canonical form, numbers
// as they were first written. Longer placeholders go first so that [PERSON_1] is not
// touched while replacing [PERSON_10].
func (v *Vault) Reidentify(text string) string {
	entries := slices.Clone(v.entries)
	slices.SortFunc(entries, func(a, b *VaultEntry) int {
		return len(b.Placeholder) - len(a.Placeholder)
	})
	for _, entry := range entries {
		text = strings.ReplaceAll(text, entry.Placeholder, entry.Display())
	}
	return text
}

// Display is the value shown to the operator in place of the placeholder.
func (e VaultEntry) Display() string {
	if e.Type.Numeric() && len(e.Forms) > 0 {
		return e.Forms[0]
	}
	return e.Canonical
}

// ApplySpans swaps every occurrence of each span for its placeholder, registering
// the value in the vault. Longer values go first so that a surname inside a full
// name is not cut twice.
func ApplySpans(text string, spans []Span, v *Vault) string {
	ordered := slices.Clone(spans)
	slices.SortStableFunc(ordered, func(a, b Span) int {
		return len(b.Text) - len(a.Text)
	})
	for _, span := range ordered {
		if strings.TrimSpace(span.Text) == "" {
			continue
		}
		placeholder := v.Assign(span.Type, span.Canonical, span.Text)
		text = strings.ReplaceAll(text, span.Text, placeholder)
	}
	return text
}

// normalizeValue makes two spellings of one value compare equal: digits only for
// numbers (with the Russian 8-prefix folded into 7), lower case and single spaces
// for everything else.
func normalizeValue(t PIIType, s string) string {
	if t.Numeric() {
		digits := Digits(s)
		if t == PIIPhone && len(digits) == 11 && digits[0] == '8' {
			digits = "7" + digits[1:]
		}
		return digits
	}
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

func Digits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
