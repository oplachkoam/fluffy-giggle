package anonymizer

// luhnValid is the check on bank card numbers; digits is the number without separators.
func luhnValid(digits string) bool {
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}
	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

// snilsValid checks the two control digits of a Russian СНИЛС (11 digits).
func snilsValid(digits string) bool {
	if len(digits) != 11 {
		return false
	}
	sum := 0
	for i := range 9 {
		sum += int(digits[i]-'0') * (9 - i)
	}
	var control int
	switch {
	case sum < 100:
		control = sum
	case sum == 100 || sum == 101:
		control = 0
	default:
		control = sum % 101
		if control == 100 {
			control = 0
		}
	}
	want := int(digits[9]-'0')*10 + int(digits[10]-'0')
	return control == want
}

// innValid checks a 10-digit (company) or 12-digit (person) Russian ИНН.
func innValid(digits string) bool {
	switch len(digits) {
	case 10:
		return innControl(digits, []int{2, 4, 10, 3, 5, 9, 4, 6, 8}) == int(digits[9]-'0')
	case 12:
		return innControl(digits, []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}) == int(digits[10]-'0') &&
			innControl(digits, []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}) == int(digits[11]-'0')
	}
	return false
}

func innControl(digits string, weights []int) int {
	sum := 0
	for i, w := range weights {
		sum += int(digits[i]-'0') * w
	}
	return sum % 11 % 10
}
