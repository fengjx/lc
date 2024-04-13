package migrate

// Config config file
type Config struct {
	DS     *DS            `yaml:"ds"`
	Target *ReverseTarget `yaml:"target"`
}

type DS struct {
	Type string `yaml:"type"`
	Dsn  string `yaml:"dsn"`
}

type Var map[string]string

type TableConfig struct {
	Module     string `yaml:"module"`
	SimpleName string `yaml:"simple-name"`
	IsTime     bool   `yaml:"is-time"`
	InputType  string `yaml:"input-type"`
	Var        Var    `yaml:"var"`
}

type ReverseTarget struct {
	Custom *Custom                `yaml:"custom"` // 自定义参数
	Tables map[string]TableConfig `yaml:"tables"` // 生成的数据库表名和对应的自定义变量
}

type Custom struct {
	OutDir      string `yaml:"out-dir"`      // 必须，生成文件路径
	TemplateDir string `yaml:"template-dir"` // 可选，模板路径
	TagName     string `yaml:"tag-name"`     // 可选，生成 model 的 tag，默认 json
	UseAdmin    bool   `yaml:"use-admin"`    // 可选，是否生成管理后台接口和页面，对所有表生效
	Var         Var    `yaml:"exy"`          // 自定义变量，对所有表生效
}

// Table represents a database table
type Table struct {
	Name          string
	StructName    string
	Columns       []Column
	PrimaryKey    Column
	AutoIncrement bool
	Comment       string
	StoreEngine   string
	GoImports     []string
}

type Column struct {
	TableName    string
	Name         string
	SQLType      string
	Comment      string
	IsPrimaryKey bool
}