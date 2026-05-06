package token

type Token struct {
	Text     string
	Lower    string
	Start    int
	End      int
	Capital  bool
	SentInit bool
}
