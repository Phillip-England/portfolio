package main

import (
	"fmt"

	"github.com/phillip-england/marki/cmd"
	"github.com/phillip-england/whip"
)

func main() {

	cli, err := whip.New(cmd.NewDefault)
	if err != nil {
		fmt.Println(err.Error())
	}

	cli.At("convert", cmd.NewConvert)
	cli.At("run", cmd.NewRun)

	err = cli.Run()
	if err != nil {
		fmt.Println(err.Error())
	}

}
