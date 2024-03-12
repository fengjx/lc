package hello

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:   "hello",
	Usage:  "hello command",
	Flags:  flags,
	Action: action,
}

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:    "name, n",
		Aliases: []string{"n"},
		Usage:   "your name",
		Value:   "foo",
	},
}

func action(ctx *cli.Context) error {
	name := ctx.String("name")
	fmt.Printf("hello %s\n", name)
	return nil
}
