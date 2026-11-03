// Package anonymizer is gate 1: deterministic detectors that find PII by shape.
// It knows nothing about models; what it misses, the model gate catches.
package anonymizer

import (
	"regexp"
	"sort"
	"strings"

	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/domain/vo"
)

type Detector struct {
	rules []rule
}

var _ port.Detector = (*Detector)(nil)

// rule finds candidates with re; group is the capture that is the value (0 for the
// whole match); valid rejects false positives; priority breaks ties on overlap.
type rule struct {
	kind     vo.PIIType
	re       *regexp.Regexp
	group    int
	valid    func(value string) bool
	priority int
}

const (
	cyrUpper = `[А-ЯЁ]`
	cyrLower = `[а-яё]`
	cyrWord  = cyrUpper + cyrLower + `+`
	// notLetter guards the left edge of Cyrillic patterns: \b is ASCII-only in RE2.
	notLetter  = `(?:^|[^\p{L}\p{N}])`
	patronymic = cyrUpper + cyrLower + `+(?:ович|евич|ьич|овна|евна|ична|инична)`
	months     = `(?:января|февраля|марта|апреля|мая|июня|июля|августа|сентября|октября|ноября|декабря)`
)

func New() *Detector {
	names := firstNamesAlternation()
	return &Detector{rules: []rule{
		// Policy and contract numbers: three letters and ten digits, or digits after «№» near the word.
		{
			kind:     vo.PIIOther,
			priority: 0,
			group:    1,
			re: regexp.MustCompile(
				`(?i)` + notLetter + `([А-ЯЁA-Z]{3} ?\d{10})\b|(?:полис\p{L}*|договор\p{L}*)[^\d\n№]{0,12}№ ?(\d{6,12})\b`,
			),
		},
		{kind: vo.PIIEmail, priority: 1,
			re: regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)},
		{kind: vo.PIICard, priority: 2, valid: luhnValid,
			re: regexp.MustCompile(`\b(?:\d{4}[ -]?){3}\d{4}\b`)},
		{kind: vo.PIIAccount, priority: 3,
			re: regexp.MustCompile(`\b\d{20}\b`)},
		{kind: vo.PIISNILS, priority: 4, valid: snilsValid,
			re: regexp.MustCompile(`\b\d{3}-\d{3}-\d{3}[ -]\d{2}\b`)},
		{
			kind:     vo.PIIPassport,
			priority: 5,
			group:    1,
			re: regexp.MustCompile(
				`(?i)(?:паспорт\p{L}*|серия|серии)[^\d\n]{0,24}(\d{2} ?\d{2} ?(?:№ ?|n ?)?\d{6})\b`,
			),
		},
		{kind: vo.PIIPassport, priority: 6,
			re: regexp.MustCompile(`\b\d{2} \d{2} \d{6}\b|\b\d{4} \d{6}\b|\b\d{4} ?№ ?\d{6}\b`)},
		{kind: vo.PIIINN, priority: 7, valid: innValid,
			re: regexp.MustCompile(`\b\d{12}\b|\b\d{10}\b`)},
		{kind: vo.PIIPhone, priority: 8,
			re: regexp.MustCompile(`(?:\+7|\b8)[ (]*\d{3}[ )]*[ -]*\d{3}[ -]*\d{2}[ -]*\d{2}\b`)},
		{kind: vo.PIIDate, priority: 9,
			re: regexp.MustCompile(`\b\d{1,2}[./]\d{1,2}[./](?:\d{4}|\d{2})\b|\b\d{4}-\d{2}-\d{2}\b`)},
		{kind: vo.PIIDate, priority: 10, group: 1,
			re: regexp.MustCompile(`(?i)\b(\d{1,2} ` + months + ` \d{4})(?: ?г\.?)?`)},
		{
			kind:     vo.PIIAddress,
			priority: 11,
			group:    1,
			re: regexp.MustCompile(
				`(?i)` + notLetter + `((?:(?:г\.|город) ?[\p{L}-]+,? ?)?(?:ул\.|улица|пр-т|проспект|пер\.|переулок|б-р|бульвар|шоссе|наб\.|набережная) ?[\p{L}\p{N}-]+(?: [\p{L}-]+){0,2},? ?(?:д\.|дом) ?\d+\p{L}?(?:,? ?(?:корп\.|к\.|стр\.) ?\d+)?(?:,? ?(?:кв\.|квартира|оф\.|офис) ?\d+)?)`,
			),
		},
		// Фамилия Имя Отчество, Имя Отчество Фамилия, Имя Отчество.
		{
			kind:     vo.PIIPerson,
			priority: 12,
			group:    1,
			re: regexp.MustCompile(
				notLetter + `(` + cyrWord + `(?:-` + cyrWord + `)? ` + cyrWord + ` ` + patronymic + `)`,
			),
		},
		{kind: vo.PIIPerson, priority: 13, group: 1,
			re: regexp.MustCompile(notLetter + `(` + cyrWord + ` ` + patronymic + `(?: ` + cyrWord + `)?)`)},
		// Фамилия И. О. and И. О. Фамилия.
		{kind: vo.PIIPerson, priority: 14, group: 1,
			re: regexp.MustCompile(notLetter + `(` + cyrWord + ` ` + cyrUpper + `\. ?` + cyrUpper + `\.)`)},
		{kind: vo.PIIPerson, priority: 15, group: 1,
			re: regexp.MustCompile(notLetter + `(` + cyrUpper + `\. ?` + cyrUpper + `\. ?` + cyrWord + `)`)},
		// Имя Фамилия and Фамилия Имя with a known first name.
		{
			kind:     vo.PIIPerson,
			priority: 16,
			group:    1,
			re: regexp.MustCompile(
				notLetter + `(` + names + ` ` + cyrUpper + cyrLower + `{2,})` + `(?:$|[^\p{L}])`,
			),
		},
		{
			kind:     vo.PIIPerson,
			priority: 17,
			group:    1,
			re: regexp.MustCompile(
				notLetter + `(` + cyrUpper + cyrLower + `{2,} ` + names + `)` + `(?:$|[^\p{L}])`,
			),
		},
	}}
}

type match struct {
	start, end int
	kind       vo.PIIType
	priority   int
}

// salutations are capitalized words that stand before a name but are not part of it.
var salutations = map[string]bool{
	"Уважаемый": true, "Уважаемая": true, "Уважаемые": true, "Здравствуйте": true,
	"Господин": true, "Госпожа": true, "Гражданин": true, "Гражданка": true,
	"Заявитель": true, "Клиент": true, "Директор": true, "Менеджер": true,
	"Руководитель": true, "Начальник": true, "Специалист": true, "Сотрудник": true,
	"Страхователь": true, "Выгодоприобретатель": true, "Представитель": true,
}

// trimSalutation moves the start of a name match past a leading salutation; a match
// that loses all but one word is dropped.
func trimSalutation(text string, m match) (match, bool) {
	for {
		rest := text[m.start:m.end]
		word, _, found := strings.Cut(rest, " ")
		if !found || !salutations[word] {
			return m, strings.Count(strings.TrimSpace(rest), " ") >= 1
		}
		m.start += len(word) + 1
	}
}

// Detect returns every value found, in order of appearance, without overlaps: a
// card number is not also an account, a full name is not also a first name.
func (d *Detector) Detect(text string) []vo.Span {
	var matches []match
	for _, r := range d.rules {
		for _, loc := range r.re.FindAllStringSubmatchIndex(text, -1) {
			start, end := loc[2*r.group], loc[2*r.group+1]
			if start < 0 && len(loc) > 2*r.group+3 {
				// The alternative's group did not match; take the next one.
				start, end = loc[2*r.group+2], loc[2*r.group+3]
			}
			if start < 0 || end <= start {
				continue
			}
			if r.valid != nil && !r.valid(vo.Digits(text[start:end])) {
				continue
			}
			m := match{start: start, end: end, kind: r.kind, priority: r.priority}
			if r.kind == vo.PIIPerson {
				var ok bool
				if m, ok = trimSalutation(text, m); !ok {
					continue
				}
			}
			matches = append(matches, m)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].start != matches[j].start {
			return matches[i].start < matches[j].start
		}
		li, lj := matches[i].end-matches[i].start, matches[j].end-matches[j].start
		if li != lj {
			return li > lj
		}
		return matches[i].priority < matches[j].priority
	})

	var spans []vo.Span
	last := -1
	for _, m := range matches {
		if m.start < last {
			continue
		}
		last = m.end
		value := strings.TrimSpace(text[m.start:m.end])
		spans = append(spans, vo.Span{Text: value, Type: m.kind, Canonical: canonical(m.kind, value)})
	}

	return spans
}

func canonical(kind vo.PIIType, value string) string {
	if kind.Numeric() {
		return vo.Digits(value)
	}
	return strings.Join(strings.Fields(value), " ")
}
