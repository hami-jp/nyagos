package parser

import "bytes"

type RedirectT struct {
	path     string
	isAppend bool
}

type StatementT struct {
	argv     []string
	redirect [3]RedirectT
	isAppend [3]bool
	term     string
}

var prefix []string = []string{" 0<", " 1>", " 2>"}

func (this StatementT) String() string {
	var buffer bytes.Buffer
	for _, arg := range this.argv {
		buffer.WriteRune('[')
		buffer.WriteString(arg)
		buffer.WriteRune(']')
	}
	for i := 0; i < len(prefix); i++ {
		if len(this.redirect[i].path) > 0 {
			buffer.WriteString(prefix[i])
			buffer.WriteString("[")
			buffer.WriteString(this.redirect[i].path)
			buffer.WriteString("]")
		}
	}
	buffer.WriteString(" ")
	buffer.WriteString(this.term)
	return buffer.String()
}

func chomp(buffer *bytes.Buffer) {
	original := buffer.String()
	buffer.Reset()
	var lastchar rune
	for i, ch := range original {
		if i > 0 {
			buffer.WriteRune(lastchar)
		}
		lastchar = ch
	}
}

func dequote(source *bytes.Buffer) string {
	var buffer bytes.Buffer

	lastchar := '\000'
	quote := false
	for {
		ch, _, ok := source.ReadRune()
		if ok != nil {
			break
		}
		if ch == '"' {
			quote = !quote
			if lastchar == '"' && quote {
				buffer.WriteRune('"')
				lastchar = '\000'
			}
		} else {
			buffer.WriteRune(ch)
		}
		lastchar = ch
	}
	return buffer.String()
}

func terminate(statements *[]StatementT,
	nextword *int,
	redirect *[3]RedirectT,
	buffer *bytes.Buffer,
	argv *[]string,
	term string) {
	var statement1 StatementT
	if buffer.Len() > 0 {
		if *nextword == WORD_ARGV {
			statement1.argv = append(*argv, dequote(buffer))
		} else {
			statement1.argv = *argv
			(*redirect)[*nextword].path = dequote(buffer)
			*nextword = WORD_ARGV
		}
		buffer.Reset()
	} else {
		statement1.argv = *argv
	}
	statement1.redirect[0] = redirect[0]
	statement1.redirect[1] = redirect[1]
	statement1.redirect[2] = redirect[2]
	redirect[0].path = ""
	redirect[0].isAppend = false
	redirect[1].path = ""
	redirect[1].isAppend = false
	redirect[2].path = ""
	redirect[2].isAppend = false
	*argv = make([]string, 0)
	statement1.term = term
	*statements = append(*statements, statement1)
}

const (
	WORD_ARGV   = -1
	WORD_STDIN  = 0
	WORD_STDOUT = 1
	WORD_STDERR = 2
)

func Parse1(text string) []StatementT {
	isQuoted := false
	statements := make([]StatementT, 0)
	argv := make([]string, 0)
	lastchar := ' '
	lastredirected := -1
	var buffer bytes.Buffer
	nextword := WORD_ARGV
	var redirect [3]RedirectT
	for _, ch := range text {
		if ch == '"' {
			isQuoted = !isQuoted
		}
		if isQuoted {
			buffer.WriteRune(ch)
		} else {
			if ch == ' ' {
				if buffer.Len() > 0 {
					if nextword == WORD_ARGV {
						argv = append(argv, dequote(&buffer))
					} else {
						redirect[nextword].path = dequote(&buffer)
					}
					buffer.Reset()
					nextword = WORD_ARGV
				}
			} else if lastchar == ' ' && ch == ';' {
				terminate(&statements, &nextword, &redirect, &buffer, &argv, ";")
			} else if ch == '|' {
				if lastchar == '|' {
					statements[len(statements)-1].term = "||"
				} else {
					terminate(&statements, &nextword, &redirect, &buffer, &argv, "|")
				}
			} else if ch == '&' {
				if lastchar == '&' {
					statements[len(statements)-1].term = "&&"
				} else {
					terminate(&statements, &nextword, &redirect, &buffer, &argv, "&")
				}
			} else if ch == '>' {
				if lastchar == '1' {
					chomp(&buffer)
					nextword = WORD_STDOUT
					redirect[1].isAppend = false
					lastredirected = 1
				} else if lastchar == '2' {
					chomp(&buffer)
					nextword = WORD_STDERR
					redirect[2].isAppend = false
					lastredirected = 2
				} else if lastchar == '>' && lastredirected >= 0 {
					redirect[lastredirected].isAppend = true
				} else {
					nextword = WORD_STDOUT
					lastredirected = 1
				}
			} else if ch == '<' {
				nextword = WORD_STDIN
				redirect[0].isAppend = false
				lastredirected = 0
			} else {
				buffer.WriteRune(ch)
			}
		}
		lastchar = ch
	}
	terminate(&statements, &nextword, &redirect, &buffer, &argv, " ")
	return statements
}

func Parse2(statements []StatementT) [][]StatementT {
	result := make([][]StatementT, 1)
	for _, statement1 := range statements {
		result[len(result)-1] = append(result[len(result)-1], statement1)
		if statement1.term != "|" {
			result = append(result, make([]StatementT, 0))
		}
	}
	if len(result[len(result)-1]) <= 0 {
		result = result[0 : len(result)-1]
	}
	return result
}

func Parse(text string) [][]StatementT {
	result1 := Parse1(text)
	result2 := Parse2(result1)
	return result2
}