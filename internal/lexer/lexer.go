package lexer

import (
	"strings"
	"unicode"

	"phi-redactor/internal/token"
)

type Lexer struct {
	input        string // the text being tokenized
	position     int
	readPosition int
	ch           byte

	sentInit bool // whether or not the current char is a sentence initializer
}

// Instantiate lexer on a given text input
func New(input string) *Lexer {
	l := &Lexer{
		input:    input,
		sentInit: true,
	}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	switch {
	case l.ch == 0:
		return token.Token{
			Text:  "",
			Lower: "",
			Start: l.position,
			End:   l.position,
		}
	case isWhitespace(l.ch):
		start := l.position
		text := l.readWhitespace()
		return createToken(text, start, false)
	case isLetter(l.ch):
		start := l.position
		sentInit := l.sentInit
		text := l.readWord()
		l.sentInit = false
		return createToken(text, start, sentInit)
	case isDigit(l.ch):
		start := l.position
		sentInit := l.sentInit
		text := l.readNumber()
		l.sentInit = false
		return createToken(text, start, sentInit)
	case isSentenceTerminator(l.ch):
		start := l.position
		ch := l.ch
		l.readChar()
		l.sentInit = true
		return createToken(string(ch), start, false)
	case isPunctuation(l.ch):
		start := l.position
		ch := l.ch
		l.readChar()
		return createToken(string(ch), start, l.sentInit)
	default:
		start := l.position
		ch := l.ch
		l.readChar()
		return createToken(string(ch), start, l.sentInit)
	}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) readWord() string {
	start := l.position
	for isLetter(l.ch) || isWordJoiner(l.ch, l.peekChar()) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) {
		l.readChar()
	}

	return l.input[start:l.position]
}

func (l *Lexer) readWhitespace() string {
	start := l.position

	for isWhitespace(l.ch) {
		l.readChar()
	}

	return l.input[start:l.position]
}

// Look ahead without advancing lexer state
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func createToken(s string, start int, sentInit bool) token.Token {
	return token.Token{
		Text:     s,
		Lower:    strings.ToLower(s),
		Start:    start,
		End:      start + len(s),
		Capital:  unicode.IsUpper([]rune(s)[0]),
		SentInit: sentInit,
	}
}

func isLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\n' || ch == '\t' || ch == '\r'
}

func isSentenceTerminator(ch byte) bool {
	return ch == '.' || ch == '!' || ch == '?'
}

func isPunctuation(ch byte) bool {
	switch ch {
	case ',', ';', ':', '"', '\'', '(', ')', '[', ']', '{', '}', '/', '\\', '-':
		return true
	default:
		return false
	}
}

func isWordJoiner(ch byte, next byte) bool {
	if ch != '\'' && ch != '!' {
		return false
	}
	return isLetter(next)
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
