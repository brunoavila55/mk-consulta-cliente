package document

import (
	"errors"
	"strings"
)

var ErrInvalidCPF = errors.New("CPF inválido")

// NormalizeCPF removes common formatting characters and validates both CPF
// check digits. It returns exactly 11 numeric digits.
func NormalizeCPF(value string) (string, error) {
	var digits strings.Builder
	digits.Grow(11)

	for _, char := range strings.TrimSpace(value) {
		switch {
		case char >= '0' && char <= '9':
			digits.WriteRune(char)
		case char == '.' || char == '-' || char == ' ' || char == '\t':
			// Accepted CPF formatting.
		default:
			return "", ErrInvalidCPF
		}
	}

	cpf := digits.String()
	if len(cpf) != 11 || allDigitsEqual(cpf) {
		return "", ErrInvalidCPF
	}

	if cpfDigit(cpf[:9], 10) != int(cpf[9]-'0') || cpfDigit(cpf[:10], 11) != int(cpf[10]-'0') {
		return "", ErrInvalidCPF
	}

	return cpf, nil
}

func allDigitsEqual(value string) bool {
	for index := 1; index < len(value); index++ {
		if value[index] != value[0] {
			return false
		}
	}
	return true
}

func cpfDigit(value string, weight int) int {
	sum := 0
	for index := range value {
		sum += int(value[index]-'0') * (weight - index)
	}
	remainder := (sum * 10) % 11
	if remainder == 10 {
		return 0
	}
	return remainder
}
