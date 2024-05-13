package models

type ParserError struct {
	Err    string `json:"error"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

func NewParserError(err string, file string, line int, column int) ParserError {
	return ParserError{
		Err:    err,
		File:   file,
		Line:   line,
		Column: column,
	}
}

func (parserError ParserError) Error() string {
	return parserError.Err
}

const (
	IncorrectReturnTuple string = "function return type should be error or (type, error)"
	InvalidType          string = "invalid type"
)
