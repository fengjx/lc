package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/fengjx/lc/commands/hello"
	"github.com/fengjx/lc/commands/migrate"
)

var Commands = []*cli.Command{
	hello.Command,
	migrate.Command,
}
