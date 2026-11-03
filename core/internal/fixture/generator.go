package fixture

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// values are the PII planted in one letter. Every field is a canary.
type values struct {
	person    person
	phone     string
	email     string
	passport  string
	snils     string
	inn       string
	card      string
	account   string
	address   string
	birthDate string
	eventDate string
	org       string
	policy    string
	plate     string
	amount    string
}

// generate is deterministic in the case number, so the fixture is the same on every
// machine and the canaries are known without a file on disk.
func generate(n int) values {
	r := rand.New(rand.NewPCG(uint64(n), 42))
	p := people[(n-1)%len(people)]
	v := values{person: p}

	v.phone = fmt.Sprintf("+7 (9%02d) %03d-%02d-%02d", r.IntN(100), r.IntN(1000), r.IntN(100), r.IntN(100))
	v.email = p.email + "@" + []string{"mail.ru", "yandex.ru", "gmail.com", "bk.ru"}[r.IntN(4)]
	v.passport = fmt.Sprintf("%02d %02d %06d", 40+r.IntN(40), 10+r.IntN(14), 100000+r.IntN(900000))
	v.snils = snils(r)
	v.inn = inn12(r)
	v.card = card(r)
	v.account = "4081781" + digits(r, 13)
	v.address = fmt.Sprintf("г. %s, ул. %s, д. %d, кв. %d",
		cities[r.IntN(len(cities))], streets[r.IntN(len(streets))], 1+r.IntN(120), 1+r.IntN(200))
	v.birthDate = fmt.Sprintf("%02d.%02d.%d", 1+r.IntN(28), 1+r.IntN(12), 1960+r.IntN(40))
	v.eventDate = fmt.Sprintf("%d %s 2026", 1+r.IntN(28), months[r.IntN(10)])
	v.org = orgs[r.IntN(len(orgs))]
	v.policy = fmt.Sprintf("%s %010d", []string{"ХХХ", "ТТТ", "ААС", "РРР"}[r.IntN(4)], r.IntN(1_000_000_000))
	letters := []rune("АВЕКМНОРСТУХ")
	v.plate = fmt.Sprintf("%c%03d%c%c %d", letters[r.IntN(12)], 1+r.IntN(999),
		letters[r.IntN(12)], letters[r.IntN(12)], 71+r.IntN(100))
	v.amount = fmt.Sprintf("%d 000", 20+r.IntN(400))

	return v
}

func (v values) canaries() []string {
	return []string{
		v.person.nom, v.person.gen, v.person.dat, v.phone, v.email, v.passport, v.snils, v.inn,
		v.card, v.account, v.address, v.birthDate, v.eventDate, v.org, v.policy, v.plate,
	}
}

// fill replaces the tokens of a template; an unknown token is a bug and stays visible.
func (v values) fill(template string) string {
	replacer := strings.NewReplacer(
		"{{FIO}}", v.person.nom, "{{FIO_GEN}}", v.person.gen, "{{FIO_DAT}}", v.person.dat,
		"{{PHONE}}", v.phone, "{{EMAIL}}", v.email, "{{PASSPORT}}", v.passport,
		"{{SNILS}}", v.snils, "{{INN}}", v.inn, "{{CARD}}", v.card, "{{ACCOUNT}}", v.account,
		"{{ADDRESS}}", v.address, "{{BIRTHDATE}}", v.birthDate, "{{DATE}}", v.eventDate,
		"{{ORG}}", v.org, "{{POLICY}}", v.policy, "{{PLATE}}", v.plate, "{{AMOUNT}}", v.amount,
	)
	return strings.TrimSpace(replacer.Replace(template))
}

func digits(r *rand.Rand, n int) string {
	var b strings.Builder
	for range n {
		b.WriteByte(byte('0' + r.IntN(10)))
	}
	return b.String()
}

func snils(r *rand.Rand) string {
	d := digits(r, 9)
	sum := 0
	for i := range 9 {
		sum += int(d[i]-'0') * (9 - i)
	}
	control := sum
	switch {
	case sum == 100 || sum == 101:
		control = 0
	case sum > 101:
		control = sum % 101
		if control == 100 {
			control = 0
		}
	}
	return fmt.Sprintf("%s-%s-%s %02d", d[0:3], d[3:6], d[6:9], control)
}

func inn12(r *rand.Rand) string {
	d := digits(r, 10)
	c1 := innControl(d, []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8})
	d += fmt.Sprint(c1)
	c2 := innControl(d, []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8})
	return d + fmt.Sprint(c2)
}

func innControl(d string, weights []int) int {
	sum := 0
	for i, w := range weights {
		sum += int(d[i]-'0') * w
	}
	return sum % 11 % 10
}

func card(r *rand.Rand) string {
	d := "4276" + digits(r, 11)
	sum := 0
	double := true
	for i := len(d) - 1; i >= 0; i-- {
		x := int(d[i] - '0')
		if double {
			x *= 2
			if x > 9 {
				x -= 9
			}
		}
		sum += x
		double = !double
	}
	d += fmt.Sprint((10 - sum%10) % 10)
	return d[0:4] + " " + d[4:8] + " " + d[8:12] + " " + d[12:16]
}
