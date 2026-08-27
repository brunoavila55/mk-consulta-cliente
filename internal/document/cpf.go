package document

import (
	"errors"
	"strings"
)

var (
	ErrInvalidCPF      = errors.New("CPF inválido")
	ErrInvalidDocument = errors.New("CPF ou CNPJ inválido")
)

// NormalizeDocument removes common formatting characters and validates the
// check digits of either an 11-digit CPF or a 14-digit CNPJ.
func NormalizeDocument(value string) (string, error) {
	digits, err := onlyDocumentDigits(value)
	if err != nil || allDigitsEqual(digits) {
		return "", ErrInvalidDocument
	}

	switch len(digits) {
	case 11:
		if !validCPF(digits) {
			return "", ErrInvalidDocument
		}
	case 14:
		if !validCNPJ(digits) {
			return "", ErrInvalidDocument
		}
	default:
		return "", ErrInvalidDocument
	}

	return digits, nil
}

// NormalizeCPF removes common formatting characters and validates both CPF
// check digits. It returns exactly 11 numeric digits.
func NormalizeCPF(value string) (string, error) {
	cpf, err := onlyDocumentDigits(value)
	if err != nil || len(cpf) != 11 || allDigitsEqual(cpf) || !validCPF(cpf) {
		return "", ErrInvalidCPF
	}
	return cpf, nil
}

func onlyDocumentDigits(value string) (string, error) {
	var digits strings.Builder
	digits.Grow(14)

	for _, char := range strings.TrimSpace(value) {
		switch {
		case char >= '0' && char <= '9':
			digits.WriteRune(char)
		case char == '.' || char == '-' || char == '/' || char == ' ' || char == '\t':
			// Common CPF/CNPJ formatting is accepted in any position.
		default:
			return "", ErrInvalidDocument
		}
	}
	return digits.String(), nil
}

func validCPF(cpf string) bool {
	return cpfDigit(cpf[:9], 10) == int(cpf[9]-'0') && cpfDigit(cpf[:10], 11) == int(cpf[10]-'0')
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

func validCNPJ(cnpj string) bool {
	firstWeights := [...]int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	secondWeights := [...]int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	return cnpjDigit(cnpj[:12], firstWeights[:]) == int(cnpj[12]-'0') &&
		cnpjDigit(cnpj[:13], secondWeights[:]) == int(cnpj[13]-'0')
}

func cnpjDigit(value string, weights []int) int {
	sum := 0
	for index := range value {
		sum += int(value[index]-'0') * weights[index]
	}
	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}
