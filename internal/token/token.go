package token

type Token struct {
	Text     string
	Lower    string
	Start    int
	End      int
	Capital  bool
	SentInit bool
}

var Auxiliaries = NewStringSet("is", "am", "are", "was", "were", "be", "been", "being",
	"have", "has", "had", "do", "does", "did", "will", "would", "shall", "should",
	"may", "might", "can", "could", "must")

var Pronouns = NewStringSet("i", "you", "he", "she", "it", "we", "they", "me", "him", "her",
	"us", "them", "my", "your", "his", "our", "their", "myself", "yourself")

var Determiners = NewStringSet("the", "a", "an", "this", "that", "these", "those",
	"some", "any", "no", "every", "each", "another")

var Prepositions = NewStringSet("of", "in", "on", "at", "by", "for", "with", "to", "from",
	"about", "into", "onto", "over", "under", "through", "between", "against")

var AmbiguousNames = NewStringSet("hope", "will", "grace", "faith", "joy", "charity",
	"mark", "pat", "may", "april", "june", "rose", "summer", "autumn", "dawn",
	"iris", "violet", "daisy", "sunny", "ray", "art", "earl", "frank", "drew",
	"wade", "brook", "brooke", "reed", "reid")

var Titles = NewStringSet("dr", "mr", "ms", "mrs", "mz")
