package stringset

type StringSet map[string]struct{}

func New(items ...string) StringSet {
	out := make(map[string]struct{})
	for _, item := range items {
		out[item] = struct{}{}
	}

	return out
}

func (s StringSet) Has(item string) bool {
	_, ok := s[item]
	return ok
}
