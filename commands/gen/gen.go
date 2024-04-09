package gen

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/fengjx/lc/commands/gen/internal/types"
	"github.com/fengjx/lc/pkg/kit"
)

//go:embed template/*
var embedFS embed.FS

var Command = &cli.Command{
	Name:   "gen",
	Usage:  "根据数据库表生成模板代码，模板可以自定义",
	Flags:  flags,
	Action: action,
}

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:     "f",
		Usage:    "配置文件路径",
		Required: true,
	},
}

func action(ctx *cli.Context) error {
	configFile := ctx.String("f")
	bs, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	config := &Config{}
	if err = yaml.Unmarshal(bs, config); err != nil {
		return err
	}
	if config.DS.Type != "mysql" {
		fmt.Println("当前仅支持 mysql")
		return nil
	}
	dsnCfg, err := mysql.ParseDSN(config.DS.Dsn)
	if err != nil {
		log.Fatal(err)
	}
	db := sqlx.MustOpen(config.DS.Type, config.DS.Dsn)
	db.Mapper = reflectx.NewMapperFunc("json", strings.ToTitle)
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	for tableName := range config.Target.Tables {
		table := loadTableMeta(db, dsnCfg.DBName, tableName)
		fmt.Println(table.Name, table.Comment)
		newGen(config, table).genFile()
	}
	return nil
}

func loadTableMeta(db *sqlx.DB, dbName, tableName string) *Table {
	args := []interface{}{dbName, tableName}
	querySQL := "SELECT `TABLE_NAME`, `ENGINE`, `AUTO_INCREMENT`, `TABLE_COMMENT` from" +
		" `INFORMATION_SCHEMA`.`TABLES` WHERE `TABLE_SCHEMA`=? AND TABLE_NAME = ?" +
		" AND (`ENGINE`='MyISAM' OR `ENGINE` = 'InnoDB' OR `ENGINE` = 'TokuDB')"

	rows, err := db.Query(querySQL, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	table := new(Table)
	for rows.Next() {
		var name, engine string
		var comment *string
		var autoIncr *int
		err = rows.Scan(&name, &engine, &autoIncr, &comment)
		if err != nil {
			log.Fatal(err)
		}
		table.Name = name
		table.StructName = kit.GonicCase(name)
		if comment != nil {
			table.Comment = *comment
		}
		table.StoreEngine = engine
		if autoIncr != nil {
			table.AutoIncrement = true
		}
	}
	if rows.Err() != nil {
		log.Fatal(err)
	}
	columns, primaryKey := loadColumnMeta(db, dbName, tableName)
	table.Columns = columns
	table.PrimaryKey = primaryKey
	table.GoImports = goImports(table.Columns)
	return table
}

// loadColumnMeta
// []*Column table column meta
// *Column PrimaryKey column
func loadColumnMeta(db *sqlx.DB, dbName, tableName string) ([]Column, Column) {
	args := []interface{}{dbName, tableName}
	querySQL := "SELECT column_name, column_type, column_comment, column_key FROM information_schema.columns " +
		"WHERE table_schema = ? AND table_name = ? ORDER BY ORDINAL_POSITION"
	rows, err := db.Query(querySQL, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var columns []Column
	var primaryKey Column
	for rows.Next() {
		var columnName string
		var columnType string
		var columnComment string
		var columnKey string
		err = rows.Scan(&columnName, &columnType, &columnComment, &columnKey)
		if err != nil {
			log.Fatal(err)
		}
		col := Column{}
		col.Name = strings.Trim(columnName, "` ")
		col.Comment = columnComment

		fields := strings.Fields(columnType)
		columnType = fields[0]
		cts := strings.Split(columnType, "(")
		colName := cts[0]
		// Remove the /* mariadb-5.3 */ suffix from coltypes
		colName = strings.TrimSuffix(colName, "/* mariadb-5.3 */")
		col.SQLType = strings.ToUpper(colName)

		if columnKey == "PRI" {
			col.IsPrimaryKey = true
			primaryKey = col
		}
		columns = append(columns, col)
	}
	return columns, primaryKey
}

type gen struct {
	config      *Config
	table       *Table
	baseTmplDir string
	outDir      string
	isEmbed     bool
	attr        map[string]any
}

func newGen(config *Config, table *Table) *gen {
	tmplDir := "template"
	isEmbed := true
	if config.Target.Custom.TemplateDir != "" {
		tmplDir = config.Target.Custom.TemplateDir
		isEmbed = false
	}
	attr := map[string]any{
		"Var":      config.Target.Custom.Var,
		"TagName":  config.Target.Custom.TagName,
		"Table":    table,
		"TableVar": config.Target.Tables[table.Name],
	}
	outDir := filepath.Join(config.Target.Custom.OutDir)
	return &gen{
		config:      config,
		table:       table,
		isEmbed:     isEmbed,
		baseTmplDir: tmplDir,
		outDir:      outDir,
		attr:        attr,
	}
}

func (g *gen) genFile() {
	entries, err := readDir(g.baseTmplDir, g.isEmbed)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	g.render("", entries)
}

// render 递归生成文件
func (g *gen) render(parent string, entries []os.DirEntry) {
	if parent == "" {
		parent = g.baseTmplDir
	}
	adminDirs := []string{"static", "endpoint"}
	for _, entry := range entries {
		if !g.config.Target.Custom.UseAdmin && kit.ContainsString(adminDirs, entry.Name()) {
			continue
		}
		path := filepath.Join(parent, entry.Name())
		if entry.IsDir() {
			children, err := readDir(filepath.Join(parent, entry.Name()), g.isEmbed)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			g.render(path, children)
			continue
		}
		targetDirBys, err := parse(strings.ReplaceAll(parent, g.baseTmplDir, ""), g.attr)
		if err != nil {
			log.Fatal(err)
		}
		targetDir := filepath.Join(g.outDir, string(targetDirBys))
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
		filenameBys, err := parse(strings.ReplaceAll(entry.Name(), suffix, ""), g.attr)
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
		bs, err := readFile(path, g.isEmbed)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		newbytes, err := parse(string(bs), g.attr)
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

func parse(text string, attr map[string]interface{}) ([]byte, error) {
	funcMap := template.FuncMap{
		"FirstUpper":           kit.FirstUpper,
		"FirstLower":           kit.FirstLower,
		"SnakeCase":            kit.SnakeCase,
		"TitleCase":            kit.TitleCase,
		"GonicCase":            kit.GonicCase,
		"LineString":           kit.LineString,
		"IsLastIndex":          kit.IsLastIndex,
		"Add":                  kit.Add,
		"Sub":                  kit.Sub,
		"SQLType2GoTypeString": types.SQLType2GoTypeString,
	}
	t, err := template.New("").Funcs(funcMap).Parse(text)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBufferString("")
	if err = t.Execute(buf, attr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func goImports(cols []Column) []string {
	imports := make(map[string]string)
	results := make([]string, 0)
	for _, col := range cols {
		if types.SQLType2GolangType(col.SQLType) == types.TimeType {
			if _, ok := imports["time"]; !ok {
				imports["time"] = "time"
				results = append(results, "time")
			}
		}
	}
	return results
}

func readDir(dir string, isEmbed bool) ([]fs.DirEntry, error) {
	if isEmbed {
		return embedFS.ReadDir(dir)
	}
	return os.ReadDir(dir)
}

func readFile(filepath string, isEmbed bool) ([]byte, error) {
	if isEmbed {
		return embedFS.ReadFile(filepath)
	}
	return os.ReadFile(filepath)
}
