package runpipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fluffy-giggle/core/internal/domain/vo"
)

func TestEgress(t *testing.T) {
	t.Parallel()

	vault := vo.NewVault()
	vault.Assign(vo.PIIPerson, "Иванов Пётр Сергеевич", "Иванова Петра Сергеевича")
	vault.Assign(vo.PIIPhone, "79161234567", "+7 (916) 123-45-67")
	vault.Assign(vo.PIIDate, "12.03.2024", "12.03.2024")

	cases := []struct {
		name    string
		payload string
		want    []vo.Finding
	}{
		{"clean", "Черновик для [PERSON_1], звонок на [PHONE_1] до 30 дней.", nil},
		{"name in another case", "Уважаемый Иванова Петра Сергеевича, ...",
			[]vo.Finding{{Type: vo.PIIPerson, Value: "Иванова Петра Сергеевича"}}},
		{"name in lower case", "ответ для иванов пётр сергеевич",
			[]vo.Finding{{Type: vo.PIIPerson, Value: "Иванов Пётр Сергеевич"}}},
		{"number without separators", "тел 7-916-123-45-67",
			[]vo.Finding{{Type: vo.PIIPhone, Value: "79161234567"}}},
		{"detector catches a new value", "звоните +7 999 000-11-22",
			[]vo.Finding{{Type: vo.PIIPhone, Value: "+7 999 000-11-22"}}},
		{"a fresh date is not a leak", "ответ до 15.11.2026", nil},
		{"the client's date is", "событие 12.03.2024",
			[]vo.Finding{{Type: vo.PIIDate, Value: "12.03.2024"}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, Egress(tt.payload, vault, stubDetector{}))
		})
	}
}

func TestCountCanaries(t *testing.T) {
	t.Parallel()

	canaries := []string{"Иванов Пётр Сергеевич", "+7 916 123-45-67", "4512 123456", ""}
	assert.Equal(t, 0, countCanaries(canaries, []string{"[PERSON_1] [PHONE_1]"}))
	assert.Equal(t, 2, countCanaries(canaries, []string{"иванов пётр сергеевич", "паспорт 4512123456"}))
	assert.Equal(t, 1, countCanaries(canaries, []string{"тел +7 (916) 123 45 67"}))
}
