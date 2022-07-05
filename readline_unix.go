//go:build !windows

package readline

import (
	"errors"
	"syscall"
	"unicode/utf8"

	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
)

type Term struct {
	orgTerm unix.Termios
}

func Open() *Term {
	term := getTerm()
	newTerm := rawMode(term)
	setTerm(newTerm)
	return &Term{orgTerm: term}
}

func (t *Term) Close() {
	setTerm(t.orgTerm)
}

func getTerm() unix.Termios {
	var term unix.Termios
	if err := termios.Tcgetattr(uintptr(syscall.Stdin), &term); err != nil {
		panic(err)
	}
	return term
}

func setTerm(term unix.Termios) {
	if err := termios.Tcsetattr(uintptr(syscall.Stdin), termios.TCSAFLUSH, &term); err != nil {
		panic(err)
	}
}

func rawMode(term unix.Termios) unix.Termios {
	term.Lflag ^= syscall.ICANON // disable canonical mode
	term.Lflag ^= syscall.ECHO   // disable input echo
	term.Lflag ^= syscall.ISIG   // do not send signals on Ctrl+C, Ctrl+\, Ctrl+Z, Ctrl+Y (SIGINT, SIGQUIT, SIGTSTP, SIGTSTP)
	term.Cc[syscall.VMIN] = 1    // minimum number of characters to read
	term.Cc[syscall.VTIME] = 0   // no read timeout
	return term
}

func (t *Term) ReadRune() (rune, int, error) {
	readBuf := make([]byte, 1)
	runeBuf := []byte{}

	n := 0
	for {
		_, err := syscall.Read(syscall.Stdin, readBuf)
		if err != nil {
			return rune(0), n, err
		}
		n++

		// Send char only when runeBuf is valid utf-8 byte sequence
		runeBuf = append(runeBuf, readBuf[0])
		if utf8.FullRune(runeBuf) {
			ch, _ := utf8.DecodeRune(runeBuf)
			return ch, n, nil
		} else if len(runeBuf) > utf8.UTFMax {
			return rune(0), n, errors.New("invalid byte sequence as utf-8")
		}
	}
}
