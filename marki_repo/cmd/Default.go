package cmd

import (
	"fmt"

	"github.com/phillip-england/whip"
)

type Default struct{}

func NewDefault(app *whip.Cli) (whip.Cmd, error) {
	return Default{}, nil
}

func (cmd Default) Execute(app *whip.Cli) error {
	fmt.Print(`marki - a runtime for converting .md into .html

run:
	marki run <SRC> <OUT> <THEME> <FLAGS>
	marki run ./README.md ./README.html dracula --watch
	marki run ./indir ./outdir dracula --watch

*to avoid issues with commas, we use <<EOF to pipe multiple lines of input to marki*
convert:
	marki convert <THEME> <MARKDOWN_STR>
	marki convert dracula <<EOF
		# My Header
		some text
		EOF

Don't forget to give me a star ⭐ at https://github.com/phillip-england/marki
`)
	return nil
}
