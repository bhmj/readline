# Readline library

## What is it?

This is a simple readline implementation for CLI. It supports a limited set of keyboard shortcuts and currently works in Linux environment only.

## Supported keys

Keys | Function
---|---
Left, Right | Move the cursor left or right
Ctrl+Left, Ctrl+Right | Move the cursor one word left or right
Home, Ctrl+A | Move the cursor to the beginning of line
End, Ctrl+E |  Move the cursor to the end of line
Backspace | Delete symbol before the cursor
Delete | Delete symbol after the cursor
Ctrl+W | Delete a word before the cursor (words are delimited by spaces)
Up, Down | Navigate through history

## Usage

```Go
import (
    "fmt"

    "github.com/bhmj/readline"
)

func main() {
    fmt.Println("Type anything or q to quit")
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
```

## Test coverage

TODO

## Benchmarks

TODO

## Changelog

**0.1.0** (2022-07-05) -- MVP.

## Roadmap

- [ ] switch to symbolic escape sequences instead of current dumb state machine
- [ ] handle `Ctrl+C`
- [ ] `Ctrl+K` to cut text to the end of line
- [ ] `Ctrl+U` to cut text to the beginning of line
- [ ] `Ctrl+N`, `Ctrl+P` == `Up`, `Down`
- [ ] `Ctrl+B`, `Ctrl+F` == `Left`, `Right`
- [ ] `Ctrl+R`, `Ctrl+S` to search in history

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :)

## Licence

[MIT](http://opensource.org/licenses/MIT)

## Author

Michael Gurov aka BHMJ
