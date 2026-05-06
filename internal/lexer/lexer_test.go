package lexer_test

import (
	"testing"

	"phi-redactor/internal/lexer"
	"phi-redactor/internal/token"

	"github.com/google/go-cmp/cmp"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []token.Token
	}{
		{
			name:  "empty input",
			input: "",
			want:  []token.Token{},
		},
		{
			name:  "single lowercase word",
			input: "hope",
			want: []token.Token{
				{Text: "hope", Lower: "hope", Start: 0, End: 4, Capital: false, SentInit: true},
			},
		},
		{
			name:  "single capitalized word",
			input: "Hope",
			want: []token.Token{
				{Text: "Hope", Lower: "hope", Start: 0, End: 4, Capital: true, SentInit: true},
			},
		},
		{
			name:  "multiple words with mixed case",
			input: "the quick Brown fox",
			want: []token.Token{
				{Text: "the", Lower: "the", Start: 0, End: 3, Capital: false, SentInit: true},
				{Text: " ", Lower: " ", Start: 3, End: 4, Capital: false, SentInit: false},
				{Text: "quick", Lower: "quick", Start: 4, End: 9, Capital: false, SentInit: false},
				{Text: " ", Lower: " ", Start: 9, End: 10, Capital: false, SentInit: false},
				{Text: "Brown", Lower: "brown", Start: 10, End: 15, Capital: true, SentInit: false},
				{Text: " ", Lower: " ", Start: 15, End: 16, Capital: false, SentInit: false},
				{Text: "fox", Lower: "fox", Start: 16, End: 19, Capital: false, SentInit: false},
			},
		},
		{
			name:  "sentence boundary after period",
			input: "Hope arrived. Will followed.",
			want: []token.Token{
				{Text: "Hope", Lower: "hope", Start: 0, End: 4, Capital: true, SentInit: true},
				{Text: " ", Lower: " ", Start: 4, End: 5, Capital: false, SentInit: false},
				{Text: "arrived", Lower: "arrived", Start: 5, End: 12, Capital: false, SentInit: false},
				{Text: ".", Lower: ".", Start: 12, End: 13, Capital: false, SentInit: false},
				{Text: " ", Lower: " ", Start: 13, End: 14, Capital: false, SentInit: false},
				{Text: "Will", Lower: "will", Start: 14, End: 18, Capital: true, SentInit: true},
				{Text: " ", Lower: " ", Start: 18, End: 19, Capital: false, SentInit: false},
				{Text: "followed", Lower: "followed", Start: 19, End: 27, Capital: false, SentInit: false},
				{Text: ".", Lower: ".", Start: 27, End: 28, Capital: false, SentInit: false},
			},
		},
		{
			name:  "sentence boundary after question mark",
			input: "Ready? Yes.",
			want: []token.Token{
				{Text: "Ready", Lower: "ready", Start: 0, End: 5, Capital: true, SentInit: true},
				{Text: "?", Lower: "?", Start: 5, End: 6, Capital: false, SentInit: false},
				{Text: " ", Lower: " ", Start: 6, End: 7, Capital: false, SentInit: false},
				{Text: "Yes", Lower: "yes", Start: 7, End: 10, Capital: true, SentInit: true},
				{Text: ".", Lower: ".", Start: 10, End: 11, Capital: false, SentInit: false},
			},
		},
		{
			name:  "sentence boundary after exclamation",
			input: "Stat! Now.",
			want: []token.Token{
				{Text: "Stat", Lower: "stat", Start: 0, End: 4, Capital: true, SentInit: true},
				{Text: "!", Lower: "!", Start: 4, End: 5, Capital: false, SentInit: false},
				{Text: " ", Lower: " ", Start: 5, End: 6, Capital: false, SentInit: false},
				{Text: "Now", Lower: "now", Start: 6, End: 9, Capital: true, SentInit: true},
				{Text: ".", Lower: ".", Start: 9, End: 10, Capital: false, SentInit: false},
			},
		},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		toks := make([]token.Token, 0, len(test.input)/2)

		for {
			tok := l.NextToken()
			if tok.Text == "" {
				break
			}
			toks = append(toks, tok)
		}
		if diff := cmp.Diff(test.want, toks); diff != "" {
			t.Logf("Error Detected on test: %s", test.name)
			t.Fatalf("lexer.NextToken() mismatch (-want +got):\n%s", diff)
		}
	}
}
