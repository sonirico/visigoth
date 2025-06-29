package visigoth

import "testing"

func intSliceCompare(t *testing.T, act, exp []int) bool {
	if len(act) != len(exp) {
		t.Errorf("unexpected slice length, want %d, have %d",
			len(exp), len(act))
		return false
	}
	for i, a := range act {
		e := exp[i]
		if a != e {
			t.Errorf("unexpected element at position %d, want %d, have %d",
				i, e, a)
			return false
		}
	}
	return true
}

type intersectionPayload struct {
	L1, L2   []int
	Expected []int
}

func Test_intersection(t *testing.T) {
	tests := []intersectionPayload{
		{
			L1:       nil,
			L2:       nil,
			Expected: nil,
		},
		{
			L1:       []int{0, 1, 2, 3, 4, 5},
			L2:       nil,
			Expected: nil,
		},
		{
			L1:       nil,
			L2:       []int{0, 1, 2, 3, 4, 5},
			Expected: nil,
		},
		{
			L1:       []int{0, 1, 2, 3, 4, 5},
			L2:       []int{0, 1, 2, 3, 4, 5},
			Expected: []int{0, 1, 2, 3, 4, 5},
		},
		{
			L1:       []int{0, 1, 2, 3, 4, 5},
			L2:       []int{0, 5, 6},
			Expected: []int{0, 5},
		},
		{
			L1:       []int{0},
			L2:       []int{1},
			Expected: nil,
		},
		{
			L1:       []int{0},
			L2:       []int{0},
			Expected: []int{0},
		},
	}

	for _, test := range tests {
		actual := intersection(test.L1, test.L2)
		if !intSliceCompare(t, actual, test.Expected) {
			t.Fatalf("unexpected intersection result. want %v, have %v",
				test.Expected, actual)
		}
	}
}
