package main

import (
	"fmt"

	"github.com/bhmj/readline"
)

func main() {

	msg := `Readline test.

Type some text, press enter. Use the following keys to test readline finctionality:

Editing
    Left/Right arrow    to move cursor
    CTRL + L/R arrow    to move cursor to previous/next word
    Home, CTRL + A      to move cursor to the beginning of line
    End, CTRL + E       to move cursor to the end of line
    Backspace, Delete   to delete character before/after cursor
    CTRL + W            to delete a word before cursor
History
    Up/Down arrow       to get previous/next line from history

    q to quit
`

	fmt.Println(msg)
	for {
		fmt.Print("> ")
		input, err := readline.Read()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}
		if input == "q" {
			break
		}
	}
}
