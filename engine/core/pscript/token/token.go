package token

type Type string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers & Literals
	IDENT = "IDENT" // variable or function names
	INT   = "INT"   // primitive numbers for scaling

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"
	EQ       = "=="
	NOT_EQ   = "!="

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"

	// Keywords
	FN     = "FN"
	LET    = "LET"
	IF     = "IF"
	ELSE   = "ELSE"
	WHILE  = "WHILE"
	RETURN = "RETURN"
	TRUE   = "TRUE"
	FALSE  = "FALSE"
)

type Token struct {
	Type    Type
	Literal string
}

var keywords = map[string]Type{
	"fn":     FN,
	"let":    LET,
	"if":     IF,
	"else":   ELSE,
	"while":  WHILE,
	"return": RETURN,
	"true":   TRUE,
	"false":  FALSE,
}

func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
