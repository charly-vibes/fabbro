package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type Token struct {
	Text  string
	Color string
}

type Highlighter struct {
	lexer chroma.Lexer
	style *chroma.Style
}

func New(filename, content string) *Highlighter {
	var lexer chroma.Lexer

	if filename != "" {
		lexer = lexers.Match(filename)
	}
	if lexer == nil {
		lexer = lexers.Analyse(content)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	return &Highlighter{
		lexer: lexer,
		style: style,
	}
}

func (h *Highlighter) HighlightLine(line string) []Token {
	iterator, err := h.lexer.Tokenise(nil, line)
	if err != nil {
		return []Token{{Text: line, Color: ""}}
	}

	var tokens []Token
	for _, token := range iterator.Tokens() {
		entry := h.style.Get(token.Type)
		color := ""
		if entry.Colour.IsSet() {
			color = entry.Colour.String()
		}
		tokens = append(tokens, Token{
			Text:  token.Value,
			Color: color,
		})
	}
	return tokens
}

func (h *Highlighter) RenderLine(line string) string {
	tokens := h.HighlightLine(line)
	var b strings.Builder
	for _, t := range tokens {
		if t.Color != "" {
			b.WriteString(ansiColor(t.Color))
			b.WriteString(t.Text)
			b.WriteString("\033[0m")
		} else {
			b.WriteString(t.Text)
		}
	}
	return b.String()
}

func ansiColor(hexColor string) string {
	if len(hexColor) != 7 || hexColor[0] != '#' {
		return ""
	}
	r := hexToByte(hexColor[1:3])
	g := hexToByte(hexColor[3:5])
	b := hexToByte(hexColor[5:7])
	return "\033[38;2;" + itoa(r) + ";" + itoa(g) + ";" + itoa(b) + "m"
}

func hexToByte(s string) int {
	var val int
	for _, c := range s {
		val *= 16
		switch {
		case c >= '0' && c <= '9':
			val += int(c - '0')
		case c >= 'a' && c <= 'f':
			val += int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			val += int(c-'A') + 10
		}
	}
	return val
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [3]byte
	n := 2
	for i > 0 {
		b[n] = byte('0' + i%10)
		i /= 10
		n--
	}
	return string(b[n+1:])
}
