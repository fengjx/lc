package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/fengjx/lc/commands/hello"
	"github.com/fengjx/lc/commands/migrate"
	"github.com/fengjx/lc/commands/start"
)

var Commands = []*cli.Command{
	hello.Command,
	start.Command,
	migrate.Command,
}
