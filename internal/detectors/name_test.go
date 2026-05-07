package detectors_test

import (
	"testing"

	"phi-redactor/internal/candidate"
	"phi-redactor/internal/detectors"
	"phi-redactor/internal/lexer"
	"phi-redactor/internal/token"

	"github.com/google/go-cmp/cmp"
)

func tokenize(input string) []token.Token {
	l := lexer.New(input)
	tokens := make([]token.Token, 0)
	for {
		tok := l.NextToken()
		if tok.Text == "" {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func TestFindNameCandidates(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []candidate.Candidate
	}{
		{
			name:  "no names lowercase",
			input: "the quick red fox",
			want:  []candidate.Candidate{},
		},
		{
			name:  "first name mid-sentence capitalized",
			input: "I saw James today",
			want: []candidate.Candidate{
				{
					Kind:       candidate.Name,
					Text:       "James",
					Start:      6,
					End:        11,
					StartToken: 4,
					EndToken:   5,
					Reason:     "firstname:capitalized",
				},
			},
		},
		{
			name:  "last name mid-sentence capitalized",
			input: "I met Smith yesterday",
			want: []candidate.Candidate{
				{
					Kind:       candidate.Name,
					Text:       "Smith",
					Start:      6,
					End:        11,
					StartToken: 4,
					EndToken:   5,
					Reason:     "lastname:capitalized",
				},
			},
		},
		{
			name:  "ambiguous name sentence-initial",
			input: "Hope arrived early.",
			want: []candidate.Candidate{
				{
					Kind:       candidate.Name,
					Text:       "Hope",
					Start:      0,
					End:        4,
					StartToken: 0,
					EndToken:   1,
					Reason:     "ambiguous:capitalized:sent_init",
				},
			},
		},
		{
			name:  "ambiguous name mid-sentence",
			input: "I saw Hope today.",
			want: []candidate.Candidate{
				{
					Kind:       candidate.Name,
					Text:       "Hope",
					Start:      6,
					End:        10,
					StartToken: 4,
					EndToken:   5,
					Reason:     "ambiguous:capitalized",
				},
			},
		},
		{
			name:  "lowercase ambiguous word not detected",
			input: "I will go home.",
			want:  []candidate.Candidate{},
		},
		{
			name:  "title prefix boosts candidate",
			input: "Dr. Smith called.",
			want: []candidate.Candidate{
				{
					Kind:       candidate.Name,
					Text:       "Smith",
					Start:      4,
					End:        9,
					StartToken: 3,
					EndToken:   4,
					Reason:     "lastname:capitalized:sent_init:has_title",
				},
			},
		},
		{
			name:  "multiple names in sentence",
			input: "James and Mary talked.",
			want: []candidate.Candidate{
				{
					Kind:       candidate.Name,
					Text:       "James",
					Start:      0,
					End:        5,
					StartToken: 0,
					EndToken:   1,
					Reason:     "firstname:capitalized:sent_init",
				},
				{
					Kind:       candidate.Name,
					Text:       "Mary",
					Start:      10,
					End:        14,
					StartToken: 4,
					EndToken:   5,
					Reason:     "firstname:capitalized",
				},
			},
		},
		{
			name:  "first and last name together",
			input: "I called John Smith.",
			want: []candidate.Candidate{
				{
					Kind:       candidate.Name,
					Text:       "John",
					Start:      9,
					End:        13,
					StartToken: 4,
					EndToken:   5,
					Reason:     "firstname:capitalized",
				},
				{
					Kind:       candidate.Name,
					Text:       "Smith",
					Start:      14,
					End:        19,
					StartToken: 6,
					EndToken:   7,
					Reason:     "lastname:capitalized",
				},
			},
		},
		{
			name:  "name that is both first and last",
			input: "I saw Thomas today.",
			want: []candidate.Candidate{
				{
					Kind:       candidate.Name,
					Text:       "Thomas",
					Start:      6,
					End:        12,
					StartToken: 4,
					EndToken:   5,
					Reason:     "firstname:lastname:capitalized",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			got := detectors.FindNameCandidates(tokens)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("FindNameCandidates() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestAmbiguousNameFiltering tests verb/noun detection for ambiguous names
// These are adversarial cases designed to require looksLikeVerb/looksLikeNoun
func TestAmbiguousNameFiltering(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTexts []string // nil means expect no candidates
	}{
		// === Verb-like usage - should NOT emit candidates ===
		{
			name:      "will_followed_by_verb",
			input:     "Will go tomorrow.",
			wantTexts: nil,
		},
		{
			name:      "will_be_auxiliary_chain",
			input:     "Will be there soon.",
			wantTexts: nil,
		},
		{
			name:      "may_as_auxiliary",
			input:     "May come later.",
			wantTexts: nil,
		},
		{
			name:      "will_have_auxiliary_chain",
			input:     "Will have finished.",
			wantTexts: nil,
		},

		// === Name usage - SHOULD emit candidates ===
		{
			name:      "will_smith_both_names",
			input:     "Will Smith arrived.",
			wantTexts: []string{"Will", "Smith"},
		},
		{
			name:      "grace_as_subject_will_auxiliary",
			input:     "Grace will help.",
			wantTexts: []string{"Grace"},
		},
		{
			name:      "may_as_name_will_auxiliary",
			input:     "May will come.",
			wantTexts: []string{"May"},
		},
		{
			name:      "will_as_auxiliary_may_as_name",
			input:     "Will May come?",
			wantTexts: []string{"Will", "May"}, // Both emitted as candidates; resolver handles
		},

		// === Noun-like usage - should NOT emit candidates ===
		{
			name:      "the_rose_determiner",
			input:     "The rose was red.",
			wantTexts: nil,
		},
		{
			name:      "a_violet_determiner",
			input:     "A violet bloomed.",
			wantTexts: nil,
		},
		{
			name:      "the_iris_in_phrase",
			input:     "I gave the iris to her.",
			wantTexts: nil,
		},
		{
			name:      "this_dawn_determiner",
			input:     "This dawn was beautiful.",
			wantTexts: nil,
		},

		// === Name usage with potential noun confusion ===
		{
			name:      "rose_smith_no_determiner",
			input:     "Rose Smith called.",
			wantTexts: []string{"Rose", "Smith"},
		},
		{
			name:      "violet_as_name_mid_sentence",
			input:     "I saw Violet yesterday.",
			wantTexts: []string{"Violet"},
		},

		// === Complex mixed cases ===
		{
			name:      "the_may_flowers_noun",
			input:     "The May flowers bloom.",
			wantTexts: nil, // "May" preceded by "The"
		},
		{
			name:      "hope_as_subject_remains_verb",
			input:     "Hope remains strong.",
			wantTexts: []string{"Hope"}, // "remains" is not auxiliary, Hope is subject
		},
		{
			name:      "faith_will_prevail",
			input:     "Faith will prevail.",
			wantTexts: []string{"Faith"}, // Faith is subject, will is auxiliary
		},

		// === Sentence boundary cases ===
		{
			name:      "sentence_init_will_verb",
			input:     "Go now. Will follow.",
			wantTexts: nil, // "Will" followed by verb despite sent_init
		},
		{
			name:      "sentence_init_hope_subject",
			input:     "Stop. Hope arrived.",
			wantTexts: []string{"Hope"}, // "Hope" followed by verb, but as subject
		},

		// === Lowercase unambiguous names ===
		{
			name:      "lowercase_unambiguous_firstname",
			input:     "i texted jonathan",
			wantTexts: []string{"jonathan"}, // jonathan is only a name, matches lowercase
		},
		{
			name:      "lowercase_unambiguous_lastname",
			input:     "called rodriguez yesterday",
			wantTexts: []string{"rodriguez"}, // rodriguez is only a name
		},
		{
			name:      "lowercase_ambiguous_still_filtered",
			input:     "i will go",
			wantTexts: nil, // will is ambiguous, requires capitalization
		},
		{
			name:      "mixed_unambiguous_and_ambiguous",
			input:     "i saw jennifer and will",
			wantTexts: []string{"jennifer"}, // jennifer matches lowercase, will doesn't
		},
		{
			name:      "multiple_lowercase_names",
			input:     "email from james to patricia",
			wantTexts: []string{"james", "patricia"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			got := detectors.FindNameCandidates(tokens)

			// Extract just the text from candidates for easier comparison
			var gotTexts []string
			for _, c := range got {
				gotTexts = append(gotTexts, c.Text)
			}

			if diff := cmp.Diff(tt.wantTexts, gotTexts); diff != "" {
				t.Errorf("FindNameCandidates() texts mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
