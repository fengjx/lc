package pbgen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	_ "embed"

	"github.com/emicklei/proto"
	"github.com/fatih/color"
	"github.com/fengjx/lc/pkg/execx"
	"github.com/fengjx/lc/pkg/formater"
	"github.com/fengjx/lc/pkg/kit"
	"github.com/urfave/cli/v2"
)

//go:embed template/handler.go.tmpl
var handlerTmpl string

//go:embed template/endpoint.go.tmpl
var endpointTmpl string

//go:embed template/service.go.tmpl
var serviceTmpl string

//go:embed template/curl.tmpl
var curlTmpl string

// Command 定义了 pbgen 子命令，用于根据 proto 文件生成代码
var Command = &cli.Command{
	Name:   "pbgen",
	Usage:  "根据 proto 文件生成代码",
	Flags:  flags,
	Action: action,
}

// flags 定义了命令行参数
var flags = []cli.Flag{
	&cli.StringFlag{
		Name:     "file",
		Aliases:  []string{"f"},
		Usage:    "指定 proto 文件或路径",
		Required: true,
	},
	&cli.StringFlag{
		Name:    "out",
		Aliases: []string{"o"},
		Usage:   "输出目录",
		Value:   "./",
	},
	&cli.StringFlag{
		Name:  "go_opt",
		Usage: "protoc 的 go_opt 参数",
		Value: "",
	},
	&cli.StringFlag{
		Name:  "grpc_opt",
		Usage: "protoc 的 go-grpc_opt 参数",
		Value: "",
	},
}

// funcMap 定义了模板中可用的函数
var funcMap = template.FuncMap{
	"lower":           strings.ToLower,
	"kebab":           kit.KebabCase,
	"split":           strings.Split,
	"genDefaultValue": genDefaultValue,
}

// 生成基本类型的默认值
func genBasicTypeValue(fieldType string) string {
	switch fieldType {
	case "string":
		return `""`
	case "int32", "int64", "int", "uint32", "uint64":
		return "0"
	case "bool":
		return "false"
	case "float32", "float64", "float", "double":
		return "0.0"
	default:
		return "null"
	}
}

// 生成消息类型的默认值
func genMessageValue(msg *Message, messages map[string]*Message, enums map[string]*Enum, indent string) string {
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, f := range msg.Fields {
		if i > 0 {
			buf.WriteString(",\n")
		}
		buf.WriteString(fmt.Sprintf("%s\"%s\": %s", indent, f.Name, genDefaultValue(f, messages, enums)))
	}
	buf.WriteString(fmt.Sprintf("\n%s}", indent[:len(indent)-2]))
	return buf.String()
}

// genDefaultValue 生成字段的默认值
func genDefaultValue(field Field, messages map[string]*Message, enums map[string]*Enum) string {
	// 处理基本类型
	basicValue := genBasicTypeValue(field.Type)
	if basicValue != "null" {
		if field.IsArray {
			return fmt.Sprintf("[%s]", basicValue)
		}
		return basicValue
	}

	// 处理枚举类型
	if field.IsEnum {
		enumValue := "0"
		if enum, ok := enums[field.Type]; ok && len(enum.Values) > 0 {
			enumValue = fmt.Sprintf("%d", enum.Values[0].Number)
		}
		if field.IsArray {
			return fmt.Sprintf("[%s]", enumValue)
		}
		return enumValue
	}

	// 处理引用类型
	typeName := getTypeName(field.Type)
	if msg, ok := messages[typeName]; ok {
		value := genMessageValue(msg, messages, enums, "      ")
		if field.IsArray {
			return fmt.Sprintf("[%s]", value)
		}
		return value
	}

	// 处理未知类型
	if field.IsArray {
		return "[]"
	}
	return "null"
}

// Field 表示 protobuf 消息中的字段定义
type Field struct {
	Name    string // 字段名称
	Type    string // 字段类型
	IsArray bool   // 是否是数组类型
	Comment string // 字段注释
	IsEnum  bool   // 是否是枚举类型
}

// Message 表示 protobuf 消息定义
type Message struct {
	Name    string  // 消息名称
	Fields  []Field // 消息包含的字段列表
	Comment string  // 消息注释
}

// EnumValue 表示枚举值
type EnumValue struct {
	Name    string // 枚举值名称
	Number  int    // 枚举值
	Comment string // 注释
}

// Enum 表示枚举类型定义
type Enum struct {
	Name    string      // 枚举名称
	Values  []EnumValue // 枚举值列表
	Comment string      // 注释
}

// Method 表示 protobuf 服务中的方法定义
type Method struct {
	Name       string   // 方法名称
	FieldName  string   // 字段名称（小写）
	InputType  string   // 输入参数类型
	OutputType string   // 输出参数类型
	HTTPPath   string   // HTTP 路径
	Comment    string   // 方法注释
	Request    *Message // 请求消息类型
	Response   *Message // 响应消息类型
}

// PbInfo 包含解析 proto 文件后的所有信息
type PbInfo struct {
	ProtoFile       string              // proto 文件路径
	Package         string              // proto 包名
	ServiceName     string              // 服务名称
	Methods         []Method            // 服务方法列表
	GoPackage       string              // go 包路径
	GoModPath       string              // go mod 路径
	EndpointPath    string              // endpoint 代码生成路径
	PkgName         string              // go_package 的最后一个路径
	EndpointPkgName string              // endpoint 包名，使用 epath 的最后一个路径
	Messages        map[string]*Message // 消息类型映射表
	Enums           map[string]*Enum    // 枚举类型映射表
}

// action 是命令的主要执行函数
func action(ctx *cli.Context) error {
	protoFile := ctx.String("file")
	outDir := ctx.String("out")

	// 创建输出目录
	if err := os.MkdirAll(outDir, 0755); err != nil {
		color.Red("创建输出目录失败: %v", err)
		return err
	}

	// 解析 proto 文件获取服务信息
	pbiInfo, err := parseProtoFile(protoFile)
	if err != nil {
		color.Red("解析 proto 文件失败: %v", err)
		return err
	}

	goOpt := ctx.String("go_opt")
	grpcOpt := ctx.String("grpc_opt")
	// 如果存在 GoModPath，将其添加到 go_opt 中
	if pbiInfo.GoModPath != "" {
		if goOpt != "" {
			goOpt = goOpt + ",module=" + pbiInfo.GoModPath
		} else {
			goOpt = "module=" + pbiInfo.GoModPath
		}

		if grpcOpt != "" {
			grpcOpt = grpcOpt + ",module=" + pbiInfo.GoModPath
		} else {
			grpcOpt = "module=" + pbiInfo.GoModPath
		}
	}

	args := []string{
		"--go_out=" + outDir,
		"--go-grpc_out=" + outDir,
	}
	if goOpt != "" {
		args = append(args, "--go_opt="+goOpt)
	}
	if grpcOpt != "" {
		args = append(args, "--go-grpc_opt="+grpcOpt)
	}
	args = append(args, protoFile)

	cmd := execx.WrapCmd("protoc", args)
	if _, err := execx.Run(cmd, ""); err != nil {
		color.Red("执行 protoc 命令失败: %v", err)
		return err
	}

	if pbiInfo.ServiceName == "" {
		return nil
	}

	// 生成 handler 文件
	if err := genHandlerFile(pbiInfo, outDir); err != nil {
		return err
	}

	// 生成 endpoint 相关文件
	if err := genEndpointFiles(pbiInfo, outDir); err != nil {
		return err
	}

	// 生成 curl 命令脚本文件
	if err := genCurlCmdFiles(pbiInfo, outDir); err != nil {
		return err
	}

	return nil
}

// 生成 handler 文件
func genHandlerFile(pbiInfo *PbInfo, outDir string) error {
	protoFile := pbiInfo.ProtoFile
	fileName := strings.TrimSuffix(filepath.Base(protoFile), ".proto")
	var handlerFile string

	if pbiInfo.GoPackage != "" {
		pkgPath := strings.TrimPrefix(pbiInfo.GoPackage, "./")
		// 去掉 gomodpath 的路径
		pkgPath = strings.ReplaceAll(pkgPath, pbiInfo.GoModPath, "")
		handlerFile = filepath.Join(outDir, pkgPath, fileName+".handler.luchen.go")
	} else {
		handlerFile = filepath.Join(outDir, fileName+".handler.luchen.go")
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(handlerFile), 0755); err != nil {
		color.Red("创建目录失败: %v", err)
		return err
	}

	// 生成 handler 文件，强制覆盖
	if err := genFileFromTemplate(handlerFile, handlerTmpl, pbiInfo, true); err != nil {
		return err
	}

	return nil
}

// 生成 endpoint 相关文件
func genEndpointFiles(pbiInfo *PbInfo, outDir string) error {
	// 使用 epath 或默认路径
	var endpointDir string
	if pbiInfo.EndpointPath == "" {
		color.Red("没有设置 spath 参数: %s", pbiInfo.ProtoFile)
		return fmt.Errorf("没有设置 spath 参数")
	}
	// 如果 EndpointPath 是相对路径，则使用 out 参数作为根路径
	if !filepath.IsAbs(pbiInfo.EndpointPath) {
		endpointDir = filepath.Join(outDir, pbiInfo.EndpointPath)
	} else {
		endpointDir = pbiInfo.EndpointPath
	}

	if err := os.MkdirAll(endpointDir, 0755); err != nil {
		color.Red("创建 endpoint 目录失败: %v", err)
		return err
	}

	// 生成 endpoint 文件，如果存在则跳过
	endpointFile := filepath.Join(endpointDir, strings.ToLower(pbiInfo.ServiceName+"endpoint.go"))
	if err := genFileFromTemplate(endpointFile, endpointTmpl, pbiInfo, true); err != nil {
		return err
	}

	// 为每个方法生成单独的服务文件，如果存在则跳过
	for _, method := range pbiInfo.Methods {
		data := struct {
			*PbInfo
			Method Method
		}{
			PbInfo: pbiInfo,
			Method: method,
		}
		filename := filepath.Join(endpointDir, strings.ToLower(method.Name)+".go")
		if err := genFileFromTemplate(filename, serviceTmpl, data, false); err != nil {
			return err
		}
	}

	return nil
}

// 从模板生成文件的通用方法
func genFileFromTemplate(filename, tmpl string, data any, forceOverwrite bool) error {
	// 获取文件的绝对路径
	absPath, err := filepath.Abs(filename)
	if err != nil {
		color.Red("获取文件绝对路径失败: %v", err)
		return err
	}

	// 如果文件存在且不强制覆盖，则跳过
	if !forceOverwrite && fileExists(filename) {
		return nil
	}

	t, err := template.New(filepath.Base(filename)).Funcs(funcMap).Parse(tmpl)
	if err != nil {
		color.Red("解析模板失败: %v", err)
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		color.Red("生成代码失败: %v", err)
		return err
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		color.Red("创建目录失败: %v", err)
		return err
	}

	// 写入文件
	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		color.Red("写入文件失败: %v", err)
		return err
	}

	// 格式化文件
	if err := formater.FormatFile(filename); err != nil {
		color.Red("格式化文件失败: %v", err)
		return err
	}

	color.Green("生成文件: %s", absPath)
	return nil
}

// 添加一个辅助函数来检查文件是否存在
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// 添加一个辅助函数来解析注释参数
func parseCommentParam(line, key string) (string, bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", false
	}
	if strings.TrimSpace(parts[0]) != key {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

// 获取类型的最后一部分（去掉包名）
func getTypeName(fullType string) string {
	parts := strings.Split(fullType, ".")
	return parts[len(parts)-1]
}

// 获取完整的类型名（包含包名）
func getFullTypeName(field *proto.NormalField) string {
	if strings.Contains(field.Type, ".") {
		return strings.TrimPrefix(field.Type, ".")
	}
	return field.Type
}

// 递归加载引用的消息类型
func loadReferencedMessages(definition *proto.Proto, data *PbInfo) {
	proto.Walk(definition,
		proto.WithMessage(func(m *proto.Message) {
			msg := &Message{
				Name: m.Name,
			}
			if m.Comment != nil {
				msg.Comment = strings.Join(m.Comment.Lines, "\n")
			}
			// 解析消息字段
			for _, element := range m.Elements {
				if field, ok := element.(*proto.NormalField); ok {
					fullType := getFullTypeName(field)
					typeName := getTypeName(field.Type)
					f := Field{
						Name:    field.Name,
						Type:    fullType, // 使用完整类型名
						IsArray: field.Repeated,
					}
					if field.Comment != nil {
						f.Comment = strings.Join(field.Comment.Lines, "\n")
					}
					// 检查是否是枚举类型
					if _, ok := data.Enums[typeName]; ok {
						f.IsEnum = true
						f.Type = typeName // 枚举类型使用短名称
					}
					msg.Fields = append(msg.Fields, f)
				}
			}
			data.Messages[m.Name] = msg
		}),
	)
}

func parseProtoFile(protoFile string) (*PbInfo, error) {
	protoContent, err := os.ReadFile(protoFile)
	if err != nil {
		color.Red("读取 proto 文件失败: %v", err)
		return nil, err
	}
	content := string(protoContent)
	reader := strings.NewReader(content)
	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		color.Red("解析proto文件失败: %v", err)
		return nil, err
	}

	data := &PbInfo{
		ProtoFile: protoFile,
		Messages:  make(map[string]*Message),
		Enums:     make(map[string]*Enum),
	}

	// 先处理枚举类型
	proto.Walk(definition,
		proto.WithEnum(func(e *proto.Enum) {
			enum := &Enum{
				Name: e.Name,
			}
			if e.Comment != nil {
				enum.Comment = strings.Join(e.Comment.Lines, "\n")
			}
			for _, element := range e.Elements {
				if field, ok := element.(*proto.EnumField); ok {
					value := EnumValue{
						Name:   field.Name,
						Number: field.Integer,
					}
					if field.Comment != nil {
						value.Comment = strings.Join(field.Comment.Lines, "\n")
					}
					enum.Values = append(enum.Values, value)
				}
			}
			data.Enums[e.Name] = enum
		}),
	)

	// 加载当前文件中的所有消息类型
	loadReferencedMessages(definition, data)

	// 加载引用的文件中的消息类型
	for _, element := range definition.Elements {
		if imp, ok := element.(*proto.Import); ok {
			importPath := strings.Trim(imp.Filename, "\"")
			// 如果是相对路径，则基于工作目录解析
			if !filepath.IsAbs(importPath) {
				// 使用当前工作目录作为基准
				workDir, err := os.Getwd()
				if err != nil {
					color.Red("获取工作目录失败: %v", err)
					continue
				}
				importPath = filepath.Join(workDir, importPath)
			}
			importContent, err := os.ReadFile(importPath)
			if err != nil {
				color.Red("读取引用的 proto 文件失败: %v", err)
				continue
			}
			importReader := strings.NewReader(string(importContent))
			importParser := proto.NewParser(importReader)
			importDefinition, err := importParser.Parse()
			if err != nil {
				color.Red("解析引用的 proto 文件失败: %v", err)
				continue
			}
			// 加载引用文件中的消息类型
			loadReferencedMessages(importDefinition, data)
		}
	}

	// 处理 Service 定义
	proto.Walk(definition,
		proto.WithPackage(func(p *proto.Package) {
			data.Package = p.Name
		}),
		proto.WithOption(func(o *proto.Option) {
			if o.Name == "go_package" {
				// 处理 go_package 选项
				goPackage := strings.Trim(o.Constant.Source, "\"")
				// 如果包含分号，取分号前的路径部分
				if idx := strings.Index(goPackage, ";"); idx != -1 {
					goPackage = goPackage[:idx]
				}
				data.GoPackage = goPackage
				// 使用最后一个路径部分作为包名
				data.PkgName = filepath.Base(goPackage)
			}
		}),
		proto.WithService(func(s *proto.Service) {
			data.ServiceName = s.Name
			for _, element := range s.Elements {
				if rpc, ok := element.(*proto.RPC); ok {
					method := Method{
						Name:       rpc.Name,
						FieldName:  strings.ToLower(rpc.Name),
						InputType:  getTypeName(rpc.RequestType),
						OutputType: getTypeName(rpc.ReturnsType),
						Request:    data.Messages[getTypeName(rpc.RequestType)], // 关联请求消息
						Response:   data.Messages[getTypeName(rpc.ReturnsType)], // 关联响应消息
					}
					// 保存 RPC 方法的注释
					if rpc.Comment != nil {
						var comments []string
						for _, line := range rpc.Comment.Lines {
							// 只去除每行开头的空格，保留换行
							line = strings.TrimLeft(line, " \t")
							if line != "" {
								comments = append(comments, line)
							}
						}
						// 使用原始的换行符连接
						method.Comment = strings.Join(comments, "\n")
					}
					// 解析 RPC 注释
					if rpc.Comment != nil {
						for _, line := range rpc.Comment.Lines {
							if value, ok := parseCommentParam(line, "http.path"); ok {
								method.HTTPPath = value
							}
						}
					}
					data.Methods = append(data.Methods, method)
				}
			}
		}),
	)

	// 解析文件级别的注释
	for _, c := range definition.Elements {
		if comment, ok := c.(*proto.Comment); ok {
			for _, line := range comment.Lines {
				if value, ok := parseCommentParam(line, "gomodpath"); ok {
					data.GoModPath = value
				} else if value, ok := parseCommentParam(line, "epath"); ok {
					data.EndpointPath = value
					// 使用 epath 的最后一个路径作为 endpoint 包名
					data.EndpointPkgName = filepath.Base(value)
				}
			}
		}
	}

	// 如果没有设置 EndpointPkgName，则使用 PkgName
	if data.EndpointPkgName == "" {
		data.EndpointPkgName = data.PkgName
	}

	return data, nil
}

// 生成 curl 命令脚本文件
func genCurlCmdFiles(pbiInfo *PbInfo, outDir string) error {
	protoFile := pbiInfo.ProtoFile
	fileName := strings.TrimSuffix(filepath.Base(protoFile), ".proto")
	var curlFile string

	if pbiInfo.EndpointPath != "" {
		// 使用 endpoint 路径作为基础路径
		curlFile = filepath.Join(outDir, pbiInfo.EndpointPath, strings.ToLower(pbiInfo.ServiceName)+"endpoint.curl")
	} else if pbiInfo.GoPackage != "" {
		pkgPath := strings.TrimPrefix(pbiInfo.GoPackage, "./")
		// 去掉 gomodpath 的路径
		pkgPath = strings.ReplaceAll(pkgPath, pbiInfo.GoModPath, "")
		curlFile = filepath.Join(outDir, pkgPath, fileName+".curl")
	} else {
		curlFile = filepath.Join(outDir, fileName+".curl")
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(curlFile), 0755); err != nil {
		color.Red("创建目录失败: %v", err)
		return err
	}

	// 生成 curl 命令文件
	if err := genFileFromTemplate(curlFile, curlTmpl, pbiInfo, true); err != nil {
		return err
	}

	// 设置文件权限为可执行
	if err := os.Chmod(curlFile, 0755); err != nil {
		color.Red("设置文件权限失败: %v", err)
		return err
	}

	return nil
}
