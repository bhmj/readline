package readline

import (
	"fmt"
	"strings"
)

const (
	Escape = 27

	keyTab       = 9
	keyUp        = 65
	keyDown      = 66
	keyRight     = 67
	keyLeft      = 68
	keyEnter     = 10
	keyDelete    = 51
	keyBackspace = 127
	keyHome      = 72
	keyEnd       = 70
	keyCtrlW     = 23
	keyCtrlA     = 1
	keyCtrlE     = 5
	keyCtrlK     = 11
	keyCtrlU     = 21
	keyCtrlN     = 14
	keyCtrlP     = 16
	keyCtrlB     = 2
	keyCtrlF     = 6
	keyCtrlArrow = 49
	keyCtrlLeft  = 68
	keyCtrlRight = 67
)

var (
	globalHistory map[string][]*string
)

func init() {
	globalHistory = make(map[string][]*string, 0)
}

type editable struct {
	scope             string    // input scope. Allows to have different input history for different inputs.
	term              *Term     // UNIX terminal attributes
	str               []rune    // current input
	pos               int       // current cursor position
	historyLines      []*string // history for current input scope
	modifiedHistory   []*string // modified history for current input session
	historyLineNumber int       // current history line number
}

// Use optional scope argument to set distinct history for each input type.
func Read(scope ...string) (string, error) {
	term := Open()
	defer term.Close()

	// get input scope
	inputScope := "default"
	for _, s := range scope {
		inputScope = s
		break
	}

	// prepare global history record if needed
	var hist []*string
	var found bool
	if hist, found = globalHistory[inputScope]; !found {
		hist = make([]*string, 0)
		globalHistory[inputScope] = hist
	}

	ed := editable{
		scope:             inputScope,
		term:              term,
		str:               make([]rune, 0),
		historyLines:      hist,
		modifiedHistory:   make([]*string, len(hist)+1),
		historyLineNumber: len(hist),
	}

	return ed.dispatch()
}

func (ed *editable) dispatch() (string, error) {
	for {
		r, _, err := ed.term.ReadRune()
		if err != nil {
			return "", err
		}

		switch r {
		case keyEnter:
			if len(ed.str) > 0 {
				str := string(ed.str)
				ed.historyLines = append(ed.historyLines, &str)
				globalHistory[ed.scope] = ed.historyLines
			}
			fmt.Println()
			return string(ed.str), nil
		case keyTab:
			continue // ignore tab
		case keyCtrlW:
			ed.deleteLastWord()
		case keyCtrlA:
			ed.toStartOfLine()
		case keyCtrlE:
			ed.toEndOfLine()
		case keyCtrlU:
			ed.deleteToStart()
		case keyCtrlK:
			ed.deleteToEnd()
		case keyCtrlP:
			ed.loadLine(ed.historyLineNumber - 1)
		case keyCtrlN:
			ed.loadLine(ed.historyLineNumber + 1)
		case keyCtrlB:
			ed.moveCursor(-1)
		case keyCtrlF:
			ed.moveCursor(1)
		case keyBackspace:
			ed.deleteSymbol(-1)
		case Escape:
			err := ed.handleControlKeys()
			if err != nil {
				return "", err
			}
		default:
			ed.AppendSymbol(r)
		}
	}
}

func (ed *editable) handleControlKeys() error {
	controlKey, err := readEscapeSequence(ed.term)
	if err != nil {
		return err
	}

	switch controlKey {
	case keyUp:
		ed.loadLine(ed.historyLineNumber - 1)
	case keyDown:
		ed.loadLine(ed.historyLineNumber + 1)
	case keyLeft:
		ed.moveCursor(-1)
	case keyRight:
		ed.moveCursor(1)
	case keyHome:
		ed.toStartOfLine()
	case keyEnd:
		ed.toEndOfLine()
	case keyDelete:
		_, _, _ = ed.term.ReadRune() // ~
		ed.deleteSymbol(0)
	case keyCtrlArrow:
		_, _, _ = ed.term.ReadRune() // 59 ;
		_, _, _ = ed.term.ReadRune() // 53 5
		x, _, _ := ed.term.ReadRune()
		switch x {
		case keyCtrlLeft:
			ed.moveOverWord(-1)
		case keyCtrlRight:
			ed.moveOverWord(1)
		}
	}
	return nil
}

func (ed *editable) deleteLastWord() {
	if ed.pos == 0 {
		return
	}
	prevpos := ed.pos
	l := len(ed.str)
	p := skipSpaces(ed.str, ed.pos, -1)
	ed.pos = skipNonSpaces(ed.str, p, -1)
	ed.str = append(ed.str[:ed.pos], ed.str[prevpos:]...)
	ed.moveLeft(prevpos - ed.pos)
	fmt.Print(strings.Repeat(" ", l-ed.pos)) // till the end of prev line
	ed.moveLeft(l - ed.pos)
	fmt.Print(string(ed.str[ed.pos:]))
	ed.moveLeft(len(ed.str) - ed.pos)
}

func (ed *editable) deleteToStart() {
	if ed.pos == 0 {
		return
	}
	ed.moveLeft(ed.pos)
	fmt.Print(strings.Repeat(" ", len(ed.str)))
	ed.moveLeft(len(ed.str))
	ed.str = ed.str[ed.pos:]
	fmt.Print(string(ed.str))
	ed.moveLeft(len(ed.str))
	ed.pos = 0
}

func (ed *editable) deleteToEnd() {
	if ed.pos == len(ed.str) {
		return
	}
	fmt.Print(strings.Repeat(" ", len(ed.str)-ed.pos))
	ed.moveLeft(len(ed.str) - ed.pos)
	ed.str = ed.str[:ed.pos]
}

// dir: -1 = left, 1 = right
func (ed *editable) moveOverWord(dir int) {
	prevpos := ed.pos
	p := skipSpaces(ed.str, ed.pos, dir)
	ed.pos = skipNonSpaces(ed.str, p, dir)
	ed.moveTo(ed.pos - prevpos)
}

// movePos: 0 = at cursor, -1 = left to cursor
func (ed *editable) deleteSymbol(movePos int) {
	if ed.pos+movePos < 0 {
		return
	}
	if ed.pos+movePos >= len(ed.str) {
		return
	}
	ed.pos += movePos
	ed.str = append(ed.str[:ed.pos], ed.str[ed.pos+1:]...)
	ed.moveTo(movePos)
	fmt.Print(string(ed.str[ed.pos:]) + " ")
	ed.moveLeft(len(ed.str) - ed.pos + 1)
}

func readEscapeSequence(term *Term) (int, error) {
	r, _, err := term.ReadRune()
	if err != nil {
		return 0, err
	}
	if int64(r) != int64('[') {
		return 0, err
	}
	r, _, err = term.ReadRune()
	if err != nil {
		return 0, err
	}
	return int(r), nil
}

// dir: <0 left, >0 right
func (ed *editable) moveCursor(dir int) {
	if ed.pos+dir < 0 || ed.pos+dir > len(ed.str) {
		return
	}
	ed.pos += dir
	ed.moveTo(dir)
}

func (ed *editable) moveLeft(n int)  { ed.moveTo(-n) }
func (ed *editable) moveRight(n int) { ed.moveTo(n) }

func (ed *editable) moveTo(n int) {
	if n > 0 {
		fmt.Printf("\x1B[%dC", n)
	} else if n < 0 {
		fmt.Printf("\x1B[%dD", -n)
	}
}

func (ed *editable) loadLine(n int) {
	if n > len(ed.historyLines) || n < 0 {
		return
	}
	if len(ed.str) > 0 {
		// save current editing to modified history
		if ed.modifiedHistory[ed.historyLineNumber] == nil || *ed.modifiedHistory[ed.historyLineNumber] != string(ed.str) {
			str := string(ed.str)
			ed.modifiedHistory[ed.historyLineNumber] = &str
		}
		// clear current line in terminal
		if ed.pos > 0 {
			ed.moveLeft(ed.pos)
		}
		fmt.Print(strings.Repeat(" ", len(ed.str)))
		ed.moveLeft(len(ed.str))
	}

	if ed.modifiedHistory[n] != nil {
		// look in modified history first
		ed.str = []rune(*ed.modifiedHistory[n])
	} else if n < len(ed.historyLines) {
		// then in history
		ed.str = []rune(*ed.historyLines[n])
	} else {
		// otherwise empty string
		ed.str = []rune{}
	}
	ed.pos = len(ed.str)
	fmt.Print(string(ed.str))

	ed.historyLineNumber = n
}

func (ed *editable) AppendSymbol(r rune) {
	if ed.pos == len(ed.str) {
		ed.str = append(ed.str, r)
	} else {
		ed.str = append(ed.str[:ed.pos+1], ed.str[ed.pos:]...)
		ed.str[ed.pos] = r
	}
	fmt.Print(string(ed.str[ed.pos:]))
	ed.pos++
	if ed.pos < len(ed.str) {
		ed.moveTo(-(len(ed.str) - ed.pos))
	}
}

func (ed *editable) toStartOfLine() {
	ed.moveLeft(ed.pos)
	ed.pos = 0
}

func (ed *editable) toEndOfLine() {
	ed.moveRight(len(ed.str) - ed.pos)
	ed.pos = len(ed.str)
}

func skipSpaces(s []rune, pos int, dir int) int {
	return skipIfSpaces(s, pos, dir, true)
}

func skipNonSpaces(s []rune, pos int, dir int) int {
	return skipIfSpaces(s, pos, dir, false)
}

func skipIfSpaces(s []rune, pos int, dir int, space bool) int {
	lookahead := 0
	if dir < 0 {
		lookahead = -1
	}
	for {
		if pos+lookahead < 0 || pos+lookahead == len(s) || (s[pos+lookahead] != ' ') == space {
			break
		}
		pos += dir
	}
	return pos
}
