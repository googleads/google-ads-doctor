package oauth

import (
  "strings"
  "testing"
)

func TestGenAuthCode(t *testing.T) {

  var tests = []struct {
    desc  string
    input string
    want  string
  }{
    {
      desc:  "Windows",
      input: "windows",
      want:  "You are running Windows",
    },
    {
      desc:  "Linux",
      input: "linux",
      want:  "Copy",
    },
  }

  for _, tt := range tests {
    got := genAuthCodePrompt(tt.input)
    if !strings.HasPrefix(got, tt.want) {
      t.Errorf("\n%s\nprompt does not match operating system want=%s\ngot=%s\ninput=%s\n",
        tt.desc, tt.want, got, tt.input)
    }
  }
}
