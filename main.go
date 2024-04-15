package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/fengjx/lc/commands"
)

// build info
var (
	version = "dev"
	commit  = ""
	date    = ""
)

const appDescription = "cli tools for lca, repo: https://github.com/fengjx/lc"
const appCopyright = "(c) 2024 by fengjianxin2012@gmail.com All rights reserved."

var Metadata = map[string]interface{}{
	"Commit":      commit,
	"Date":        date,
	"Description": appDescription,
}

func main() {
	app := cli.NewApp()
	app.Name = "lc"
	app.Copyright = appCopyright
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
