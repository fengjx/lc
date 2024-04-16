package main

import (
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/fengjx/lc/commands"
	"github.com/fengjx/lc/common"
)

// build info
var (
	version = strings.Join([]string{"v1.0.0", common.GitInfo.Branch, common.GitInfo.Hash}, "-")
	commit  = ""
	date    = ""
)

const appDescription = "lca 命令行工具, 源码: https://github.com/fengjx/lc"

var Metadata = map[string]interface{}{
	"Commit":      commit,
	"Date":        date,
	"Description": appDescription,
}

func main() {
	app := cli.NewApp()
	app.Name = "lc"
	app.Version = version
	app.Description = appDescription
	app.EnableBashCompletion = true
	app.Commands = commands.Commands
	app.Metadata = Metadata
	app.Suggest = true
	app.Authors = []*cli.Author{
		{
			Name:  "fengjx",
			Email: "fengjianxin2012@gmail.com",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
