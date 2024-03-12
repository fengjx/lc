package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/fengjx/lc/commands/gen"
	"github.com/fengjx/lc/commands/hello"
)

var Commands = []*cli.Command{
	hello.Command,
	gen.Command,
}
