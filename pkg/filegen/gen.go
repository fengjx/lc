package filegen

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/fengjx/lc/pkg/kit"
)

// EntryFilter 模板文件过滤
// 返回true表示需要生成，false表示不需要生成
type EntryFilter func(ctx context.Context, entry os.DirEntry) bool

type FileGen struct {
	ctx         context.Context
	BaseTmplDir string
	OutDir      string
	EmbedFS     *embed.FS
	IsEmbed     bool
	Attr        map[string]any
	FuncMap     template.FuncMap
	EntryFilter EntryFilter
}

func (g *FileGen) Gen() {
	entries, err := g.readDir(g.BaseTmplDir)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	g.render("", entries)
}

func (g *FileGen) With(ctx context.Context, key any, value any) *FileGen {
	if g.ctx == nil {
		g.ctx = context.Background()
	}
	g.ctx = context.WithValue(ctx, key, value)
	return g
}

// render 递归生成文件
func (g *FileGen) render(parent string, entries []os.DirEntry) {
	if parent == "" {
		parent = g.BaseTmplDir
	}
	for _, entry := range entries {
		if g.EntryFilter != nil && !g.EntryFilter(g.ctx, entry) {
			continue
		}
		path := filepath.Join(parent, entry.Name())
		if entry.IsDir() {
			children, err := g.readDir(filepath.Join(parent, entry.Name()))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			g.render(path, children)
			continue
		}
		targetDirBys, err := g.parse(strings.ReplaceAll(parent, g.BaseTmplDir, ""), g.Attr)
		if err != nil {
			log.Fatal(err)
		}
		targetDir := filepath.Join(g.OutDir, string(targetDirBys))
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		suffix := ""
		override := false
		re := false
		if strings.HasSuffix(entry.Name(), ".override.tmpl") {
			// 覆盖之前的文件
			suffix = ".override.tmpl"
			override = true
		} else if strings.HasSuffix(entry.Name(), ".re.tmpl") {
			// 重新生成一个文件，加上时间戳
			suffix = ".re.tmpl"
			re = true
		} else if strings.HasSuffix(entry.Name(), ".tmpl") {
			// 保留原来的文件
			suffix = ".tmpl"
		}
		if suffix == "" {
			targetFile := filepath.Join(targetDir, entry.Name())
			fmt.Println(targetFile)
			// 其他不需要渲染的文件直接复制
			err = kit.CopyFile(path, targetFile)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			continue
		}
		filenameBys, err := g.parse(strings.ReplaceAll(entry.Name(), suffix, ""), g.Attr)
		if err != nil {
			log.Fatal(err)
		}
		targetFile := filepath.Join(targetDir, string(filenameBys))
		if _, err = os.Stat(targetFile); !override && err == nil {
			if !re {
				continue
			}
			targetFile = fmt.Sprintf("%s.%d", targetFile, time.Now().Unix())
		}
		fmt.Println(targetFile)
		bs, err := g.readFile(path)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		newbytes, err := g.parse(string(bs), g.Attr)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = os.WriteFile(targetFile, newbytes, 0600)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func (g *FileGen) parse(text string, attr map[string]interface{}) ([]byte, error) {
	t, err := template.New("").Funcs(g.FuncMap).Parse(text)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBufferString("")
	if err = t.Execute(buf, attr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *FileGen) readDir(dir string) ([]fs.DirEntry, error) {
	if g.EmbedFS != nil {
		return g.EmbedFS.ReadDir(dir)
	}
	return os.ReadDir(dir)
}

func (g *FileGen) readFile(filepath string) ([]byte, error) {
	if g.EmbedFS != nil {
		return g.EmbedFS.ReadFile(filepath)
	}
	return os.ReadFile(filepath)
}
