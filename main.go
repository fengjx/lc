package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/fengjx/lc/commands"
	"github.com/fengjx/lc/common"
)

const appName = "lc"
const appAbout = "源码: https://github.com/fengjx/lc"
const appDescription = "luchen cli tools"
const appCopyright = "(c) 2024 by xd-fjx@qq.com All rights reserved."

var Metadata = map[string]interface{}{
	"Name":        appName,
	"About":       appAbout,
	"Description": appDescription,
}

func main() {
	app := cli.NewApp()
	app.Version = common.Version
	app.Usage = appAbout
	app.Description = appDescription
	app.Copyright = appCopyright
	app.EnableBashCompletion = true
	app.Commands = commands.Commands
	app.Metadata = Metadata
	app.Suggest = true

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
