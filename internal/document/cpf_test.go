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
