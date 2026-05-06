package detectors

import (
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

var AmbiguousNames = stringset.New("hope", "will", "grace", "faith", "joy", "charity",
	"mark", "pat", "may", "april", "june", "rose", "summer", "autumn", "dawn",
	"iris", "violet", "daisy", "sunny", "ray", "art", "earl", "frank", "drew",
	"wade", "brook", "brooke", "reed", "reid")

var Titles = stringset.New("dr", "mr", "ms", "mrs", "mz")
