package start

import (
	"context"
	"embed"
	"path/filepath"
	"text/template"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"

	"github.com/fengjx/lc/common"
	"github.com/fengjx/lc/pkg/filegen"
	"github.com/fengjx/lc/pkg/kit"
)

//go:embed template/*
var embedFS embed.FS

const tipsLckit = `
项目创建完成，执行一下步骤启动服务
1. cd %s
2. go mod tidy
3. 修改数据库配置 conf/app.yml
4. 初始化数据库 go run tools/init/main.go
5. 启动服务 go run main.go
`

const tipsSimple = `
项目创建完成，执行一下步骤启动服务
1. cd %s
2. go mod tidy
5. 启动服务 go run main.go
`

var tmplTips = map[string]string{
	"lckit":  tipsLckit,
	"simple": tipsSimple,
}

var Command = &cli.Command{
	Name:   "start",
	Usage:  "开始一个新项目",
	Flags:  flags,
	Action: action,
}

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:     "gomod",
		Aliases:  []string{"m"},
		Usage:    "指定 go.mod module",
		Required: true,
	},
	&cli.StringFlag{
		Name:    "out",
		Aliases: []string{"o"},
		Usage:   "文件生成目录，默认从 gomod 读取",
	},
	&cli.StringFlag{
		Name:    "template",
		Aliases: []string{"t"},
		Usage:   "使用模板，可选参数：lckit, simple，默认 simple",
		Value:   "simple",
	},
}

func action(ctx *cli.Context) error {
	debug := common.IsDebug()
	rctx := context.Background()
	mod := ctx.String("m")
	out := ctx.String("o")
	tmplType := ctx.String("t")
	proj := filepath.Base(mod)
	if out == "" {
		out = proj
	}
	attr := map[string]any{
		"gomod": mod,
		"proj":  proj,
	}
	funcMap := template.FuncMap{
		"FirstUpper":  kit.FirstUpper,
		"FirstLower":  kit.FirstLower,
		"SnakeCase":   kit.SnakeCase,
		"TitleCase":   kit.TitleCase,
		"GonicCase":   kit.GonicCase,
		"LineString":  kit.LineString,
		"IsLastIndex": kit.IsLastIndex,
		"Add":         kit.Add,
		"Sub":         kit.Sub,
	}
	var eFS *embed.FS
	tmplDir := filepath.Join("template", tmplType)
	if debug {
		eFS = &embedFS
	} else {
		unzipDir, err := common.SyncTemplate(rctx)
		if err != nil {
			color.Red("同步模板文件失败，%s", err.Error())
			return err
		}
		tmplDir = filepath.Join(unzipDir, "template", "start")
	}
	fg := &filegen.FileGen{
		EmbedFS:     eFS,
		BaseTmplDir: tmplDir,
		OutDir:      out,
		Attr:        attr,
		FuncMap:     funcMap,
	}
	fg.Gen()
	color.Green(tmplTips[tmplType], out)
	return nil
}
