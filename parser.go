package gotreescript

// Modes
type ScopeType int

const (
	SCOPE_TEXT ScopeType = iota
	SCOPE_TAG_START
	SCOPE_COMMENT
	SCOPE_TAG_NAME
	SCOPE_ATTRIBUTE_START
	SCOPE_ATTRIBUTE
	SCOPE_ASSIGNMENT // Search for `:` or `=`
	SCOPE_VALUE_START
	SCOPE_VALUE
	SCOPE_VALUE_SIMPLE_QUOTED
	SCOPE_VALUE_DOUBLE_QUOTED
)

type Parser struct {
	scope  ScopeType
	tokens []*Token
	token  *Token
}

func NewParser() *Parser {
	return &Parser{}
}

func Parse(code string) *[]*Token {

	p := NewParser()
	return p.Parse(code)
}

func (this *Parser) Parse(code string) *[]*Token {

	this.scope = SCOPE_TEXT
	this.token = &Token{}
	this.tokens = []*Token{}

	attribute := ""
	value := ""

	n := len(code)

	for i := 0; i < n; i++ {
		c := string(code[i])
		cc := c
		if i+1 < n {
			cc += string(code[i+1])
		}

		switch this.scope {

		case SCOPE_TEXT:

			if "[[" == cc {
				i++
				this.add_token()
				this.token.Partial = "[["
				this.scope = SCOPE_TAG_START
			} else {
				this.token.Partial += c
			}

		case SCOPE_TAG_START:

			if " " == c || "\n" == c || "\t" == c {
				this.token.Partial += c
				this.token.Type = COMMENT
				this.scope = SCOPE_COMMENT
			} else if "]]" == cc {
				i++
				this.token.Partial += cc
				this.token.Type = COMMENT
				this.scope = SCOPE_COMMENT
				this.add_token()
				this.scope = SCOPE_TEXT
			} else {
				this.token.Type = TAG
				this.token.Name += c
				this.token.Partial += c
				this.token.Flags = []string{}
				this.token.Args = map[string]string{}
				this.scope = SCOPE_TAG_NAME
			}

		case SCOPE_COMMENT:

			if "]]" == cc {
				i++
				this.token.Partial += cc
				this.add_token()
				this.scope = SCOPE_TEXT
			} else {
				this.token.Partial += c
			}

		case SCOPE_TAG_NAME:

			if "]]" == cc {
				i++
				this.token.Partial += "]]"
				this.add_token()
				this.scope = SCOPE_TEXT
			} else if " " == c || "\n" == c || "\t" == c {
				this.token.Partial += c
				this.scope = SCOPE_ATTRIBUTE_START
			} else {
				this.token.Name += c
				this.token.Partial += c
			}

		case SCOPE_ATTRIBUTE_START:
			/*
				Look for attribute start or tag end
			*/

			if "]]" == cc {
				i++
				this.token.Partial += "]]"
				this.add_token()
				this.scope = SCOPE_TEXT
			} else if " " == c || "\n" == c || "\t" == c {
				this.token.Partial += c
			} else {
				this.token.Partial += c
				attribute = c
				this.scope = SCOPE_ATTRIBUTE
			}

		case SCOPE_ATTRIBUTE:
			/*
				First letter of the attribute has been found, now join new
				letters and look for:
				tag end: `]]`
				separator: ` ` | `\n` | `\t`
			*/

			if "]]" == cc {
				i++
				this.token.Partial += "]]"
				this.token.Flags = append(this.token.Flags, attribute)
				this.add_token()
				this.scope = SCOPE_TEXT
			} else if ":" == c || "=" == c {
				// Move to look for value :)
				this.token.Partial += c
				this.scope = SCOPE_VALUE_START
			} else if " " == c || "\n" == c || "\t" == c {
				this.token.Partial += c
				this.scope = SCOPE_ASSIGNMENT
			} else {
				this.token.Partial += c
				attribute += c
			}

		case SCOPE_ASSIGNMENT:
			/*
				The attribute is stored, so, I look for a separator or ending
			*/

			if "]]" == cc {
				i++
				this.token.Partial += "]]"
				this.token.Flags = append(this.token.Flags, attribute)
				this.add_token()
				this.scope = SCOPE_TEXT
			} else if " " == c || "\n" == c || "\t" == c {
				this.token.Partial += c
			} else if ":" == c || "=" == c {
				// Move to look for value :)
				this.token.Partial += c
				this.scope = SCOPE_VALUE_START
			} else {
				// Move to another attribute
				this.token.Partial += c
				this.token.Flags = append(this.token.Flags, attribute)
				attribute = c
				this.scope = SCOPE_ATTRIBUTE
			}

		case SCOPE_VALUE_START:

			if "]]" == cc {
				i++
				this.token.Partial += "]]"
				this.token.Args[attribute] = ""
				this.add_token()
				this.scope = SCOPE_TEXT
			} else if " " == c || "\n" == c || "\t" == c {
				this.token.Partial += c
			} else if "'" == c {
				this.token.Partial += c
				value = ""
				this.scope = SCOPE_VALUE_SIMPLE_QUOTED
			} else if "\"" == c {
				this.token.Partial += c
				value = ""
				this.scope = SCOPE_VALUE_DOUBLE_QUOTED
			} else {
				this.token.Partial += c
				value = c
				this.scope = SCOPE_VALUE
			}

		case SCOPE_VALUE:

			if "]]" == cc {
				i++
				this.token.Partial += "]]"
				this.token.Args[attribute] = value
				this.add_token()
				this.scope = SCOPE_TEXT
			} else if " " == c || "\n" == c || "\t" == c {
				this.token.Partial += c
				this.token.Args[attribute] = value
				this.scope = SCOPE_ATTRIBUTE_START
			} else {
				this.token.Partial += c
				value += c
			}

		case SCOPE_VALUE_SIMPLE_QUOTED:

			if "\\\\" == cc {
				i++
				this.token.Partial += cc
				value += "\\"
			} else if "\\'" == cc {
				i++
				this.token.Partial += cc
				value += "'"
			} else if "\\\"" == cc {
				i++
				this.token.Partial += cc
				value += "\""
			} else if "'" == c {
				this.token.Partial += c
				this.token.Args[attribute] = value
				this.scope = SCOPE_ATTRIBUTE_START
			} else {
				this.token.Partial += c
				value += c
			}

		case SCOPE_VALUE_DOUBLE_QUOTED:

			if "\\\\" == cc {
				i++
				this.token.Partial += cc
				value += "\\"
			} else if "\\'" == cc {
				i++
				this.token.Partial += cc
				value += "'"
			} else if "\\\"" == cc {
				i++
				this.token.Partial += cc
				value += "\""
			} else if "\"" == c {
				this.token.Partial += c
				this.token.Args[attribute] = value
				this.scope = SCOPE_ATTRIBUTE_START
			} else {
				this.token.Partial += c
				value += c
			}
		}
	}
	this.add_token()

	_ = value // TODO: remove this line

	return &this.tokens
}

func (this *Parser) add_token() {
	this.tokens = append(this.tokens, this.token)
	this.token = &Token{}
}
