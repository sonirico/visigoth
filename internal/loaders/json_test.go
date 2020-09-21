package loaders

import (
	"bytes"
	"strings"
	"testing"
)

func assertAllTokensInString(t *testing.T, actual, expectedTokens string) (expectedButNotPresent []string, ok bool) {
	t.Helper()

	ok = true
	allActual := strings.Split(actual, " ")
	allActualMap := make(map[string]bool)
	for _, actualToken := range allActual {
		allActualMap[actualToken] = true
	}

	allExpected := strings.Split(expectedTokens, " ")

	for _, expected := range allExpected {
		if _, in := allActualMap[expected]; !in {
			expectedButNotPresent = append(expectedButNotPresent, expected)
			ok = false
		}
	}

	return
}

func TestJSONLoader_Load(t *testing.T) {
	tests := []struct {
		payload        string
		expectedTokens string
	}{
		{
			payload: `{
				"name": "Pikachu",
				"color": "yellow",
				"type":"electric", 
				"position": 25, 
				"evolutions": ["raychu", "pichu"],
				"attackPoints": {
					"ray": 60,
					"thunder": 120
                }
			}`,
			expectedTokens: "Pikachu yellow electric 25 raychu pichu 60 120",
		},
	}

	for _, test := range tests {
		src := bytes.NewBufferString(test.payload)
		dst := bytes.NewBuffer(nil)
		loader := NewJSONLoader(false)
		if err := loader.Load(src, dst); err != nil {
			t.Fatal(err)
		}
		actual := dst.String()
		missingTokens, ok := assertAllTokensInString(t, actual, test.expectedTokens)
		if !ok {
			t.Error("bad json loading. these words have not been loaded: ", missingTokens)
		}
	}
}
