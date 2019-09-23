package oauth

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestReadCustomerID(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	original := readStdin
	defer func() { readStdin = original }()

	tests := []struct {
		desc  string
		input string
		want  string
	}{
		{
			desc:  "Return a valid customer ID",
			input: "123-456-7890",
			want:  "1234567890",
		},
		{
			desc:  "Return original string",
			input: "abc",
			want:  "abc",
		},
	}

	for _, test := range tests {
		readStdin = func() string {
			return test.input
		}

		got := ReadCustomerID()
		if got != test.want {
			t.Errorf("[%s] got: %s, want: %s\n", test.desc, got, test.want)
		}
	}
}
