package mhparser

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

const (
	EOFRune = -1 // serve al parser per riconoscere la fine del source
)

// / RUN Stack used in Rewind
type runeNode struct {
	r    rune
	next *runeNode
}

type runeStack struct {
	start *runeNode
}

func newRuneStack() runeStack {
	return runeStack{}
}

func (s *runeStack) push(r rune) {
	node := &runeNode{r: r}
	if s.start == nil {
		s.start = node
	} else {
		node.next = s.start
		s.start = node
	}
}

func (s *runeStack) pop() rune {
	if s.start == nil {
		return EOFRune
	} else {
		n := s.start
		s.start = n.next
		return n.r
	}
}

func (s *runeStack) clear() {
	s.start = nil
}

// ********** TokenType ******************

type TokenType int

type Token struct {
	Type  TokenType
	ID    int
	Value string
}

func (tk *Token) String() string {
	switch tk.Type {
	case itemEOF:
		return "EOF"
	case itemError:
		return tk.Value
	}
	if len(tk.Value) > 30 {
		return fmt.Sprintf("%.10q...", tk.Value)
	}
	return fmt.Sprintf("%q", tk.Value)
}

// ********** Lexer ******************
type StateFunc func(*L) StateFunc

type PropInfo struct {
	Keyword   string
	TokenType TokenType
}

type L struct {
	source          string
	start, position int
	state           StateFunc
	tokens          chan Token
	runstack        runeStack
	// custom
	descrFns   DescrFns
	scriptLine int
}

// NewL creates a returns a lexer ready to parse the given source code.
func NewL(src string, start StateFunc) *L {
	l := L{
		source:     src,
		state:      start,
		start:      0,
		position:   0,
		scriptLine: 1,
		runstack:   newRuneStack(),
	}
	buffSize := len(l.source) / 2
	if buffSize <= 0 {
		buffSize = 1
	}
	l.tokens = make(chan Token, buffSize)

	return &l
}

func (l *L) current() string {
	return l.source[l.start:l.position]
}

func (l *L) emit(t TokenType) {
	tok := Token{
		Type:  t,
		Value: l.current(),
	}
	l.tokens <- tok
	l.start = l.position
	l.runstack.clear()
}

func (l *L) emitCustFn(t TokenType, id int) {
	tok := Token{
		Type:  t,
		ID:    id,
		Value: l.current(),
	}
	l.tokens <- tok
	l.start = l.position
	l.runstack.clear()
}

func (l *L) inc_line(r rune) {
	if r == '\n' {
		l.scriptLine += 1
	}
}

func (l *L) ignore() {
	l.runstack.clear()
	l.start = l.position
}

func (l *L) peek() rune {
	r := l.next()
	l.rewind()
	return r
}

func (l *L) rewind() {
	r := l.runstack.pop()
	if r > EOFRune {
		size := utf8.RuneLen(r)
		l.position -= size
		if l.position < l.start {
			l.position = l.start
		}
	}
}

func (l *L) next() rune {
	var (
		r rune
		s int
	)
	str := l.source[l.position:]
	if len(str) == 0 {
		r, s = EOFRune, 0
	} else {
		r, s = utf8.DecodeRuneInString(str)
		if r == utf8.RuneError && s == 1 {
			r, s = EOFRune, 0
		}
	}
	l.position += s
	l.runstack.push(r)

	return r
}

func (l *L) nextItem() Token {
	for {
		select {
		case item := <-l.tokens:
			return item
		default:
			if l.state != nil {
				l.state = l.state(l)
			} else {
				return Token{Type: itemEOF, Value: ""}
			}
		}
	}
}

func (l *L) errorf(format string, args ...interface{}) StateFunc {
	l.tokens <- Token{
		Type:  itemError,
		Value: fmt.Sprintf(format, args...),
	}
	return nil
}

// //////////////////////////////////////////////////////////////////////
// / Lexer Custom Part for specific task

// To add a new keyword follow:
// 1) add a new item into the DescrFnItem
// 2) build and run a test with the new keyword
// TokenType to String is manually done into the file token_type_string.go (no stringer anymore because has changed its usage)
const (
	// Internal Types
	itemText TokenType = iota
	itemBuiltinFunction
	itemVarValue
	itemAssign
	itemComment
	itemEmptyString
	itemEndOfStatement
	itemError
	itemFunctionName
	itemFunctionStartBlock
	itemFunctionEnd
	itemArrayBegin
	itemArrayEnd
	itemVarName
	itemParamString
	itemBegMdHtml
	itemMdHtmlBlock
	itemMdHtmlBlockLine
	itemSeparator
	itemEndOfBlock
	itemFigStack
	itemLinkSimple
	itemLinkCaption
	itemYouTubeEmbed
	itemLatestPosts
	itemArchivePosts
	itemTagPosts
	itemSingleTaggedPosts
	itemEOF
)

type DescrFnItem struct {
	KeyName       string
	ItemTokenType TokenType
	NumParam      int
	CustomID      int
	Internal      bool
	VariableArgs  bool
	InfoDet       string
	Labels        []string
	IsMultiline   bool
}
type DescrFns []DescrFnItem

// ///////////////// State functions ///////////////////////////////////
func lexStateEndOfStatement(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case r == EOFRune:
			return nil
		case r == '\r' || r == '\n':
			l.ignore()
			l.emit(itemEndOfStatement)
			l.inc_line(r)
			return lexStateInit
		case r == '#':
			l.rewind()
			l.emit(itemEndOfStatement)
			return lexStateInComment
		case unicode.IsSpace(r):
			l.ignore()
		default:
			return l.errorf("[lexStateEndOfStatement] Expected only one statement per line: '%s'", l.source[l.start:l.position])
		}
	}
}

func lexStateInComment(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case r == EOFRune:
			l.rewind()
			l.emit(itemComment)
			return nil
		case r == '\r' || r == '\n':
			l.rewind()
			l.emit(itemComment)
			l.inc_line(r)
			return lexStateInit
		}
	}
}

func lexStateAssignInValue(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case r == EOFRune:
			l.rewind()
			l.emit(itemVarValue)
			return nil
		case r == '\n':
			l.rewind()
			l.emit(itemVarValue)
			return lexStateEndOfStatement
		case r == '\r':
			l.rewind()
			l.emit(itemVarValue)
			return lexStateEndOfStatement
		case r == '#':
			l.rewind()
			l.emit(itemVarValue)
			return lexStateInComment
		}
	}
}

func lexStateAssignRight(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case r == ':':
			l.emit(itemAssign)
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			l.rewind()
			return lexStateAssignInValue
		case r == '/':
			l.rewind()
			return lexStateAssignInValue
		case r == EOFRune || r == '\n' || r == '\r':
			l.rewind()
			l.emit(itemVarValue)
			return lexStateEndOfStatement
		case unicode.IsSpace(r):
			l.ignore()
		default:
			return l.errorf("[lexStateAssignRight] expect string assign (lexStateAssignRight) or a known function name:  %q (Line %d)", l.source[l.start:l.position], l.scriptLine)
		}
	}
}

func lexStateInVariableAssign(l *L) StateFunc {
	for {
		switch r := l.next(); {
		case unicode.IsSpace(r):
			l.rewind()
			l.emit(itemVarName)
			return lexStateAssignRight
		case r == ':':
			l.rewind()
			l.emit(itemVarName)
			return lexStateAssignRight
		case r == '(':
			return l.errorf("Unexpected function declaration. Expected variable assignment. Function spelling? %q", l.source[l.start:l.position])
		case !unicode.IsDigit(r) && !unicode.IsLetter(r):
			return l.errorf("[lexStateInVariableAssign] Unexpected charcter '%c' in variable name. Name found '%s'. Please use only digit or letter for a variable name", r, l.source[l.start:l.position])
		}
	}
}

func lexStateInit(l *L) StateFunc {
	for {
		rlf := l.peek()
		//fmt.Print("*** peek ", rlf)
		for rlf == '\r' || rlf == '\n' {
			l.inc_line(rlf)
			l.next()
			l.ignore()
			rlf = l.peek()
			//fmt.Print("*** peek - next ", rlf)
		}
		if nextstate, ok := lexMatchMdHtmlKey(l); ok {
			return nextstate
		}
		switch r := l.next(); {
		case r == EOFRune:
			return nil
		case r == '\r' || r == '\n':
			l.inc_line(r)
			l.ignore()
		case unicode.IsSpace(r):
			l.ignore()
		case unicode.IsLetter(r):
			return lexStateInVariableAssign
		case r == '#':
			return lexStateInComment
		default:
			return l.errorf("Character is not suitable for any statement %s", l.source[l.start:])
		}
	}
}
