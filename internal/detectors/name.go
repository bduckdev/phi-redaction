package detectors

import (
	"fmt"

	"github.com/cloudflare/ahocorasick"

	"phi-redactor/internal/candidate"
	"phi-redactor/internal/token"
	"phi-redactor/pkg/stringset"
)

var Auxiliaries = stringset.New("is", "am", "are", "was", "were", "be", "been", "being",
	"have", "has", "had", "do", "does", "did", "will", "would", "shall", "should",
	"may", "might", "can", "could", "must")

var Pronouns = stringset.New("i", "you", "he", "she", "it", "we", "they", "me", "him", "her",
	"us", "them", "my", "your", "his", "our", "their", "myself", "yourself")

var Determiners = stringset.New("the", "a", "an", "this", "that", "these", "those",
	"some", "any", "no", "every", "each", "another")

var Prepositions = stringset.New("of", "in", "on", "at", "by", "for", "with", "to", "from",
	"about", "into", "onto", "over", "under", "through", "between", "against")

var Titles = stringset.New("dr", "mr", "ms", "mrs", "mz")

// === THIS IS WHERE THE SSA/CENSUS BUREAU DATA WILL GO ===
// Raw name lists for building AC matcher
var firstNamesList = []string{
	"james", "mary", "john", "patricia", "robert", "jennifer", "michael", "linda",
	"david", "elizabeth", "william", "barbara", "richard", "susan", "joseph", "jessica",
	"thomas", "sarah", "charles", "karen", "christopher", "lisa", "daniel", "nancy",
	"matthew", "betty", "anthony", "margaret", "mark", "sandra", "donald", "ashley",
	"steven", "kimberly", "paul", "emily", "andrew", "donna", "joshua", "michelle",
	"kenneth", "dorothy", "kevin", "carol", "brian", "amanda", "george", "melissa",
	"timothy", "deborah", "ronald", "stephanie", "edward", "rebecca", "jason", "sharon",
	"jeffrey", "laura", "ryan", "cynthia", "jacob", "kathleen", "gary", "amy",
	"nicholas", "angela", "eric", "shirley", "jonathan", "anna", "stephen", "brenda",
}

var lastNamesList = []string{
	"smith", "johnson", "williams", "brown", "jones", "garcia", "miller", "davis",
	"rodriguez", "martinez", "hernandez", "lopez", "gonzalez", "wilson", "anderson",
	"thomas", "taylor", "moore", "jackson", "martin", "lee", "perez", "thompson",
	"white", "harris", "sanchez", "clark", "ramirez", "lewis", "robinson", "walker",
	"young", "allen", "king", "wright", "scott", "torres", "nguyen", "hill", "flores",
	"green", "adams", "nelson", "baker", "hall", "rivera", "campbell", "mitchell",
	"carter", "roberts", "gomez", "phillips", "evans", "turner", "diaz", "parker",
	"cruz", "edwards", "collins", "reyes", "stewart", "morris", "morales", "murphy",
}

var ambiguousNamesList = []string{
	"hope", "will", "grace", "faith", "joy", "charity",
	"mark", "pat", "may", "april", "june", "rose", "summer", "autumn", "dawn",
	"iris", "violet", "daisy", "sunny", "ray", "art", "earl", "frank", "drew",
	"wade", "brook", "brooke", "reed", "reid",
}

// nameCategory tracks which name lists a pattern belongs to
type nameCategory struct {
	isFirst     bool
	isLast      bool
	isAmbiguous bool
}

// NameMatcher uses Aho-Corasick for fast multi-pattern matching
type NameMatcher struct {
	ac         *ahocorasick.Matcher
	patterns   []string       // original patterns for exact match check
	categories []nameCategory // category per pattern index
}

// NewNameMatcher builds an AC matcher from all name lists
func NewNameMatcher() *NameMatcher {
	// Deduplicate and track categories
	seen := make(map[string]int) // pattern -> index in patterns
	var patterns []string
	var categories []nameCategory

	addNames := func(names []string, setFirst, setLast, setAmbig bool) {
		for _, name := range names {
			if idx, exists := seen[name]; exists {
				// Update existing category
				if setFirst {
					categories[idx].isFirst = true
				}
				if setLast {
					categories[idx].isLast = true
				}
				if setAmbig {
					categories[idx].isAmbiguous = true
				}
			} else {
				// Add new pattern
				seen[name] = len(patterns)
				patterns = append(patterns, name)
				categories = append(categories, nameCategory{
					isFirst:     setFirst,
					isLast:      setLast,
					isAmbiguous: setAmbig,
				})
			}
		}
	}

	addNames(firstNamesList, true, false, false)
	addNames(lastNamesList, false, true, false)
	addNames(ambiguousNamesList, false, false, true)

	return &NameMatcher{
		ac:         ahocorasick.NewStringMatcher(patterns),
		patterns:   patterns,
		categories: categories,
	}
}

// Classify checks if a token matches any name pattern exactly
// Returns which categories the name belongs to
func (nm *NameMatcher) Classify(tokenLower string) (isFirst, isLast, isAmbiguous bool) {
	hits := nm.ac.MatchThreadSafe([]byte(tokenLower))

	for _, idx := range hits {
		// Exact match only - pattern must equal full token
		if nm.patterns[idx] == tokenLower {
			cat := nm.categories[idx]
			isFirst = isFirst || cat.isFirst
			isLast = isLast || cat.isLast
			isAmbiguous = isAmbiguous || cat.isAmbiguous
		}
	}
	return
}

// DefaultNameMatcher is the singleton matcher instance
var DefaultNameMatcher = NewNameMatcher()

// FindNameCandidates scans tokens for potential name matches.
// Uses Aho-Corasick for fast pattern matching, then applies filters.
// Emits candidates for resolver to handle ambiguity.
func FindNameCandidates(tokens []token.Token) []candidate.Candidate {
	return FindNameCandidatesWithMatcher(tokens, DefaultNameMatcher)
}

// FindNameCandidatesWithMatcher allows injecting a custom matcher for testing
func FindNameCandidatesWithMatcher(tokens []token.Token, nm *NameMatcher) []candidate.Candidate {
	candidates := make([]candidate.Candidate, 0)

	for i, tok := range tokens {
		// Skip whitespace and punctuation
		if tok.Text == "" || !isWord(tok.Text) {
			continue
		}

		// Stage 1: Aho-Corasick lookup
		isFirst, isLast, isAmbig := nm.Classify(tok.Lower)
		if !isFirst && !isLast && !isAmbig {
			continue
		}

		// Require capitalization only for ambiguous names
		// Unambiguous names (jonathan, rodriguez) match regardless of case
		if !tok.Capital && isAmbig {
			continue
		}

		// Stage 2: Apply filters for ambiguous names
		// Skip if word appears to be used as verb or common noun
		if isAmbig && !isFirst && !isLast {
			// Pure ambiguous word (not also a known first/last name)
			if looksLikeVerb(tokens, i) {
				continue
			}
			if looksLikeNoun(tokens, i) {
				continue
			}
		}

		// Stage 3: Build reasons and context signals
		var reasons []string

		if isFirst {
			reasons = append(reasons, "firstname")
		}
		if isLast {
			reasons = append(reasons, "lastname")
		}
		if isAmbig {
			reasons = append(reasons, "ambiguous")
		}

		// Add context signals
		if tok.Capital {
			reasons = append(reasons, "capitalized")
		} else {
			reasons = append(reasons, "lowercase")
		}
		if tok.SentInit {
			reasons = append(reasons, "sent_init")
		}

		// Check for title prefix (previous non-whitespace token)
		if hasTitlePrefix(tokens, i) {
			reasons = append(reasons, "has_title")
		}

		candidates = append(candidates, candidate.Candidate{
			Kind:       candidate.Name,
			Text:       tok.Text,
			Start:      tok.Start,
			End:        tok.End,
			StartToken: i,
			EndToken:   i + 1,
			Reason:     buildReason(reasons),
		})
	}

	return candidates
}

// hasTitlePrefix checks if token at idx has a title (Dr, Mr, etc.) before it
func hasTitlePrefix(tokens []token.Token, idx int) bool {
	// Walk backward skipping whitespace and period
	for i := idx - 1; i >= 0; i-- {
		tok := tokens[i]
		if isWhitespaceToken(tok.Text) {
			continue
		}
		if tok.Text == "." {
			continue
		}
		return Titles.Has(tok.Lower)
	}
	return false
}

func isWord(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != '\'' {
			return false
		}
	}
	return true
}

func isWhitespaceToken(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}

func buildReason(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result = fmt.Sprintf("%s:%s", result, parts[i])
	}
	return result
}

// nextWordToken returns the next non-whitespace token after idx, or nil
func nextWordToken(tokens []token.Token, idx int) *token.Token {
	for i := idx + 1; i < len(tokens); i++ {
		if !isWhitespaceToken(tokens[i].Text) {
			return &tokens[i]
		}
	}
	return nil
}

// prevWordToken returns the previous non-whitespace token before idx, or nil
func prevWordToken(tokens []token.Token, idx int) *token.Token {
	for i := idx - 1; i >= 0; i-- {
		if !isWhitespaceToken(tokens[i].Text) {
			return &tokens[i]
		}
	}
	return nil
}

// Grammatical auxiliary pairs - first aux can be followed by second aux
// "will be", "will have", "would be", "would have", "may be", "may have", etc.
var validAuxPairs = map[string]stringset.StringSet{
	"will":   stringset.New("be", "have"),
	"would":  stringset.New("be", "have"),
	"shall":  stringset.New("be", "have"),
	"should": stringset.New("be", "have"),
	"may":    stringset.New("be", "have"),
	"might":  stringset.New("be", "have"),
	"can":    stringset.New("be", "have"),
	"could":  stringset.New("be", "have"),
	"must":   stringset.New("be", "have"),
}

// looksLikeVerb checks if an ambiguous word appears to be used as an auxiliary verb
// Example: "Will go" -> true, "Will Smith" -> false, "May will come" -> false (May is subject)
func looksLikeVerb(tokens []token.Token, idx int) bool {
	tok := tokens[idx]

	// Check if it's a known auxiliary that could be a name
	if !Auxiliaries.Has(tok.Lower) {
		return false
	}

	// Check if preceded by an auxiliary - if so, this is likely a subject (name)
	// "Will May come?" - "Will" is auxiliary, "May" is subject
	// "Did Grace help?" - "Did" is auxiliary, "Grace" is subject
	prev := prevWordToken(tokens, idx)
	if prev != nil && Auxiliaries.Has(prev.Lower) {
		return false
	}

	// Look at next word
	next := nextWordToken(tokens, idx)
	if next == nil {
		return false
	}

	// If followed by a capitalized word, likely name + name pattern
	// "Will Smith" - both are names
	// "Will May come?" - ambiguous, let resolver handle (emit both as candidates)
	if next.Capital {
		return false
	}

	// If next word is a lowercase auxiliary, check if it's a valid auxiliary pair
	// "Will be going" - "will be" is valid pair → Will is auxiliary
	// "May will come" - "may will" is NOT valid pair → May is likely name
	if Auxiliaries.Has(next.Lower) {
		validFollowers, exists := validAuxPairs[tok.Lower]
		if exists && validFollowers.Has(next.Lower) {
			// Valid auxiliary pair like "will be", "may have"
			return true
		}
		// Invalid pair like "may will" - first word is likely a name
		return false
	}

	// If followed by a verb-like word (lowercase, not auxiliary), likely an auxiliary verb
	// "Will go" - "go" is lowercase verb, "Will" is auxiliary
	if !next.Capital && isWord(next.Text) {
		return true
	}

	return false
}

// looksLikeNoun checks if an ambiguous word appears to be used as a common noun
// Example: "the rose" -> true, "Rose Smith" -> false
func looksLikeNoun(tokens []token.Token, idx int) bool {
	// Check if preceded by determiner
	prev := prevWordToken(tokens, idx)
	if prev != nil && Determiners.Has(prev.Lower) {
		return true
	}

	return false
}
