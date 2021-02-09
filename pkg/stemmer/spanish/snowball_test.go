package spanish

import (
	"bytes"
	"testing"
)

func TestRV(t *testing.T) {
	tests := []struct {
		word string
		r1   string
		r2   string
		rv   string
	}{
		{
			word: "macho",
			r1:   "ho",
			r2:   "",
			rv:   "ho",
		},
		{
			word: "oliva",
			r1:   "iva",
			r2:   "a",
			rv:   "va",
		},
		{
			word: "trabajo",
			r1:   "ajo",
			r2:   "o",
			rv:   "bajo",
		},
		{
			word: "áureo",
			r1:   "eo",
			r2:   "",
			rv:   "eo",
		},
	}

	for _, test := range tests {
		r1, r2, rv := r1r2rv([]byte(test.word))

		if r2 != nil && test.r2 != "" {
			if bytes.Compare(r2.data, []byte(test.r2)) != 0 {
				t.Fatalf("unexpected R2, want '%s', have '%s'",
					test.r2, string(r2.data))
			}
		}

		if r1 == nil {
			t.Fatalf("unexpected R1, want '%s', have nil",
				test.r1)
		}

		if bytes.Compare(r1.data, []byte(test.r1)) != 0 {
			t.Fatalf("unexpected R1, want '%s', have '%s'",
				test.r1, string(r1.data))
		}
		if rv == nil {
			t.Fatalf("unexpected RV, want '%s', have nil",
				test.rv)
		}

		if bytes.Compare(rv.data, []byte(test.rv)) != 0 {
			t.Fatalf("unexpected RV, want '%s', have '%s'",
				test.rv, string(rv.data))
		}
	}
}
