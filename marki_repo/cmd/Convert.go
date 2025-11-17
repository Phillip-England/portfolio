package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/phillip-england/marki/mdfile"
	"github.com/phillip-england/wherr"
	"github.com/phillip-england/whip"
)

type Convert struct {
	Theme string
	MdStr string
}

func NewConvert(cli *whip.Cli) (whip.Cmd, error) {
	theme, err := cli.ArgGetByPositionForce(2, "<THEME> is required in 'marki convert <THEME> <MARKDOWN_STR>'")
	if err != nil {
		return Convert{}, wherr.Consume(wherr.Here(), err, "")
	}
	err = isValidTheme(theme)
	if err != nil {
		return Convert{}, wherr.Consume(wherr.Here(), err, "")
	}
	mdBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return Convert{}, wherr.Consume(wherr.Here(), err, "")
	}
	return Convert{
		Theme: theme,
		MdStr: string(mdBytes),
	}, nil
}

func (cmd Convert) Execute(cli *whip.Cli) error {
	mdFile, err := mdfile.NewMarkdownFileFromStr(cmd.MdStr, cmd.Theme)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	_, err = fmt.Print(mdFile.Html)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing HTML to stdout:", err)
		return wherr.Consume(wherr.Here(), err, "")
	}
	return nil
}
