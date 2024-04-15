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
	builtBy = ""
)

const appAbout = "源码: https://github.com/fengjx/lc"
const appDescription = "luchen cli tools"
const appCopyright = "(c) 2024 by xd-fjx@qq.com All rights reserved."

var Metadata = map[string]interface{}{
	"About":       appAbout,
	"Description": appDescription,
	"Commit":      commit,
	"Date":        date,
	"builtBy":     builtBy,
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

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
