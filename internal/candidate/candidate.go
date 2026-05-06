package candidate

type Kind int

const Name Kind = iota

type Candidate struct {
	Kind  Kind   // what kind of thing is this. name, phone, etc.
	Text  string // text content
	Start int    // index of first byte in original text
	End   int    // index of last byte in original text

	StartToken int // index of first token covered in input
	EndToken   int // index of last token covered, exlcusive

	Reason string
}
