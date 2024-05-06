package migrate

import (
	"context"
	"embed"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/fengjx/lc/commands/migrate/internal/types"
	"github.com/fengjx/lc/common"
	"github.com/fengjx/lc/pkg/filegen"
	"github.com/fengjx/lc/pkg/kit"
)

//go:embed template/*
var embedFS embed.FS

var Command = &cli.Command{
	Name:   "migrate",
	Usage:  "根据数据库表生成模板代码，模板可以自定义",
	Flags:  flags,
	Action: action,
}

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:     "c",
		Usage:    "配置文件路径",
		Required: true,
	},
}

func action(ctx *cli.Context) error {
	configFile := ctx.String("c")
	bs, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	config := &Config{}
	if err = yaml.Unmarshal(bs, config); err != nil {
		return err
	}
	if config.DS.Type != "mysql" {
		color.Red("当前仅支持 mysql")
		return nil
	}
	if config.Target.Custom.Gomod == "" {
		color.Red("缺少配置[target.custom.gomod]")
		return nil
	}
	if config.Target.Custom.OutDir == "" {
		config.Target.Custom.OutDir = "./"
	}

	dsnCfg, err := mysql.ParseDSN(config.DS.Dsn)
	if err != nil {
		color.Red("数据库dsn配置错误：%s, %s", config.DS.Dsn, err.Error())
		return err
	}
	db := sqlx.MustOpen(config.DS.Type, config.DS.Dsn)
	db.Mapper = reflectx.NewMapperFunc("json", strings.ToTitle)
	err = db.Ping()
	if err != nil {
		color.Red("数据库连接失败，请检查配置，%s", err.Error())
		return err
	}

	debug := common.IsDebug()
	rctx := context.Background()

	var eFS *embed.FS
	tmplDir := "template"
	if debug {
		eFS = &embedFS
	} else if config.Target.Custom.TemplateDir != "" {
		tmplDir = config.Target.Custom.TemplateDir
	} else {
		unzipDir, err := common.SyncTemplate(rctx)
		if err != nil {
			color.Red("同步模板文件失败，%s", err.Error())
			panic(err)
		}
		tmplDir = filepath.Join(unzipDir, "template", "migrate")
	}

	for tableName := range config.Target.Tables {
		table := loadTableMeta(db, dsnCfg.DBName, tableName)
		color.Green("[%s %s]", table.Name, table.Comment)
		newGen(tmplDir, eFS, config, table).Gen()
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
		color.Red("读取表信息失败：%s, %s", tableName, err.Error())
		panic(err)
	}
	defer rows.Close()
	table := new(Table)
	for rows.Next() {
		var name, engine string
		var comment *string
		var autoIncr *int
		err = rows.Scan(&name, &engine, &autoIncr, &comment)
		if err != nil {
			color.Red("读取表信息失败：%s, %s", tableName, err.Error())
			panic(err)
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
		color.Red("读取表信息失败：%s, %s", tableName, err.Error())
		panic(err)
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
	querySQL := `SELECT
			column_name,
			column_type,
			column_comment, 
			column_key,
			ifnull(column_default, ''),
			ifnull(extra, '')
		FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ? ORDER BY ORDINAL_POSITION`
	rows, err := db.Query(querySQL, args...)
	if err != nil {
		color.Red("读取表[%s]字段信息失败", tableName)
		panic(err)
	}
	defer rows.Close()
	var columns []Column
	var primaryKey Column
	for rows.Next() {
		var columnName string
		var columnType string
		var columnComment string
		var columnKey string
		var columnDefault string
		var extra string
		err = rows.Scan(
			&columnName,
			&columnType,
			&columnComment,
			&columnKey,
			&columnDefault,
			&extra,
		)
		if err != nil {
			color.Red("读取表[%s]字段信息失败", tableName)
			panic(err)
		}
		col := Column{
			TableName: tableName,
		}
		col.Name = strings.Trim(columnName, "` ")
		col.Comment = columnComment
		col.DefaultValue = columnDefault
		col.Extra = extra

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
	*filegen.FileGen
}

func newGen(tmplDir string, eFS *embed.FS, config *Config, table *Table) *gen {
	tableOpt := config.Target.Tables[table.Name]
	if tableOpt.UseAdmin == nil {
		tableOpt.UseAdmin = &config.Target.Custom.UseAdmin
	}
	attr := map[string]any{
		"Var":      config.Target.Custom.Var,
		"TagName":  config.Target.Custom.TagName,
		"Gomod":    config.Target.Custom.Gomod,
		"Table":    table,
		"TableOpt": tableOpt,
	}
	outDir := filepath.Join(config.Target.Custom.OutDir)
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
	fg := &filegen.FileGen{
		EmbedFS:     eFS,
		BaseTmplDir: tmplDir,
		OutDir:      outDir,
		Attr:        attr,
		FuncMap:     funcMap,
	}
	if !*tableOpt.UseAdmin {
		// 排除admin目录
		fg.EntryFilter = func(ctx context.Context, entry os.DirEntry) bool {
			adminDirs := []string{"static", "endpoint"}
			return !kit.ContainsString(adminDirs, entry.Name())
		}
	}
	g := &gen{
		FileGen: fg,
	}
	return g
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
