package document

import "testing"

func TestNormalizeCPF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
		valid bool
	}{
		{name: "formatted", input: "529.982.247-25", want: "52998224725", valid: true},
		{name: "digits", input: "52998224725", want: "52998224725", valid: true},
		{name: "wrong check digit", input: "52998224724", valid: false},
		{name: "repeated digits", input: "11111111111", valid: false},
		{name: "letters", input: "529A98224725", valid: false},
		{name: "empty", input: "", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeCPF(test.input)
			if test.valid && err != nil {
				t.Fatalf("NormalizeCPF() error = %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("NormalizeCPF() expected error")
			}
			if got != test.want {
				t.Errorf("NormalizeCPF() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeDocument(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
		valid bool
	}{
		{name: "formatted CPF", input: "529.982.247-25", want: "52998224725", valid: true},
		{name: "CPF with misplaced punctuation", input: "529982.24725", want: "52998224725", valid: true},
		{name: "formatted CNPJ", input: "04.252.011/0001-10", want: "04252011000110", valid: true},
		{name: "CNPJ digits", input: "04252011000110", want: "04252011000110", valid: true},
		{name: "CNPJ with misplaced punctuation", input: "04252.0110001/10", want: "04252011000110", valid: true},
		{name: "invalid CNPJ check digit", input: "04.252.011/0001-11", valid: false},
		{name: "repeated CNPJ", input: "11.111.111/1111-11", valid: false},
		{name: "wrong length", input: "123456789012", valid: false},
		{name: "letters", input: "04.252.01A/0001-10", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeDocument(test.input)
			if test.valid && err != nil {
				t.Fatalf("NormalizeDocument() error = %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("NormalizeDocument() expected error")
			}
			if got != test.want {
				t.Errorf("NormalizeDocument() = %q, want %q", got, test.want)
			}
		})
	}
}
