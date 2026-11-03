package anonymizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fluffy-giggle/core/internal/domain/vo"
)

func TestDetect(t *testing.T) {
	t.Parallel()

	d := New()
	cases := []struct {
		name string
		text string
		want []vo.Span
	}{
		{
			"email",
			"пишите на ivan.petrov@example.com срочно",
			[]vo.Span{
				{Text: "ivan.petrov@example.com", Type: vo.PIIEmail, Canonical: "ivan.petrov@example.com"},
			},
		},
		{"phone with brackets", "тел. +7 (916) 123-45-67.",
			[]vo.Span{{Text: "+7 (916) 123-45-67", Type: vo.PIIPhone, Canonical: "79161234567"}}},
		{"phone with eight", "звоните 8 916 123 45 67",
			[]vo.Span{{Text: "8 916 123 45 67", Type: vo.PIIPhone, Canonical: "89161234567"}}},
		{"card by luhn", "карта 4111 1111 1111 1111",
			[]vo.Span{{Text: "4111 1111 1111 1111", Type: vo.PIICard, Canonical: "4111111111111111"}}},
		{"card with bad luhn is not a card", "номер 4111 1111 1111 1112", nil},
		{
			"account",
			"счёт 40817810099910004312 в банке",
			[]vo.Span{
				{Text: "40817810099910004312", Type: vo.PIIAccount, Canonical: "40817810099910004312"},
			},
		},
		{"snils", "СНИЛС 112-233-445 95",
			[]vo.Span{{Text: "112-233-445 95", Type: vo.PIISNILS, Canonical: "11223344595"}}},
		{"inn person", "ИНН 500100732259",
			[]vo.Span{{Text: "500100732259", Type: vo.PIIINN, Canonical: "500100732259"}}},
		{"passport with words", "паспорт РФ серия 45 12 № 123456, выдан",
			[]vo.Span{{Text: "45 12 № 123456", Type: vo.PIIPassport, Canonical: "4512123456"}}},
		{"passport bare", "документ 4512 123456",
			[]vo.Span{{Text: "4512 123456", Type: vo.PIIPassport, Canonical: "4512123456"}}},
		{"date numeric", "произошло 12.03.2024 утром",
			[]vo.Span{{Text: "12.03.2024", Type: vo.PIIDate, Canonical: "12.03.2024"}}},
		{"date words", "родился 5 мая 1987 г.",
			[]vo.Span{{Text: "5 мая 1987", Type: vo.PIIDate, Canonical: "5 мая 1987"}}},
		{
			"full name with patronymic",
			"Я, Иванов Пётр Сергеевич, прошу",
			[]vo.Span{
				{Text: "Иванов Пётр Сергеевич", Type: vo.PIIPerson, Canonical: "Иванов Пётр Сергеевич"},
			},
		},
		{
			"name patronymic surname",
			"Уважаемая Анна Павловна Смирнова!",
			[]vo.Span{
				{Text: "Анна Павловна Смирнова", Type: vo.PIIPerson, Canonical: "Анна Павловна Смирнова"},
			},
		},
		{"initials", "Директор Кузнецов А. В. подписал",
			[]vo.Span{{Text: "Кузнецов А. В.", Type: vo.PIIPerson, Canonical: "Кузнецов А. В."}}},
		{"first name from dictionary", "менеджер Елена Соколова ответила",
			[]vo.Span{{Text: "Елена Соколова", Type: vo.PIIPerson, Canonical: "Елена Соколова"}}},
		{
			"address",
			"проживаю по адресу г. Тула, ул. Ленина, д. 5, кв. 12.",
			[]vo.Span{
				{
					Text:      "г. Тула, ул. Ленина, д. 5, кв. 12",
					Type:      vo.PIIAddress,
					Canonical: "г. Тула, ул. Ленина, д. 5, кв. 12",
				},
			},
		},
		{"policy number", "полис КАСКО ХХХ 0264857094, прошу",
			[]vo.Span{{Text: "ХХХ 0264857094", Type: vo.PIIOther, Canonical: "ХХХ 0264857094"}}},
		{"contract number after the word", "по договору № 123456789 выплата",
			[]vo.Span{{Text: "123456789", Type: vo.PIIOther, Canonical: "123456789"}}},
		{"plain words are not names", "Прошу рассмотреть претензию в срок.", nil},
		{"amount is not an id", "сумма 125 000 рублей", nil},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, d.Detect(tt.text))
		})
	}
}

func TestDetectResolvesOverlaps(t *testing.T) {
	t.Parallel()

	spans := New().Detect("карта 4111 1111 1111 1111, тел +7 916 123-45-67, Иванов Иван Иванович")
	require.Len(t, spans, 3)
	assert.Equal(t, vo.PIICard, spans[0].Type)
	assert.Equal(t, vo.PIIPhone, spans[1].Type)
	assert.Equal(t, vo.PIIPerson, spans[2].Type)
}

func TestApplySpans(t *testing.T) {
	t.Parallel()

	text := "Иванов Пётр Сергеевич, тел. +7 916 123-45-67. Повторно: +7 916 123-45-67. Иванов Пётр Сергеевич."
	vault := vo.NewVault()
	out := vo.ApplySpans(text, New().Detect(text), vault)

	assert.Equal(t, "[PERSON_1], тел. [PHONE_1]. Повторно: [PHONE_1]. [PERSON_1].", out)
	assert.Equal(t, 2, vault.Len())
	assert.Equal(t, text, vault.Reidentify(out))
}
