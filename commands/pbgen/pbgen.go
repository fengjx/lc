package pbgen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	_ "embed"

	"github.com/emicklei/proto"
	"github.com/fatih/color"
	"github.com/fengjx/lc/pkg/execx"
	"github.com/urfave/cli/v2"
)

//go:embed template/handler.go.tmpl
var handlerTmpl string

//go:embed template/endpoint.go.tmpl
var endpointTmpl string

//go:embed template/service.go.tmpl
var serviceTmpl string

// 添加模板函数
var funcMap = template.FuncMap{
	"lower": strings.ToLower,
	"kebab": toKebabCase,
	"split": strings.Split,
}

// 添加 kebab case 转换函数
func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

var Command = &cli.Command{
	Name:   "pbgen",
	Usage:  "根据 proto 文件生成代码",
	Flags:  flags,
	Action: action,
}

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
}

type Method struct {
	Name       string
	FieldName  string
	InputType  string
	OutputType string
	HTTPPath   string
	Comment    string
}

type PbInfo struct {
	Package      string
	ServiceName  string
	Methods      []Method
	GoPackage    string
	GoModPath    string
	EndpointPath string
}

func action(ctx *cli.Context) error {
	protoFile := ctx.String("file")
	outDir := ctx.String("out")

	// 1. 创建输出目录
	if err := os.MkdirAll(outDir, 0755); err != nil {
		color.Red("创建输出目录失败: %v", err)
		return err
	}

	// 2. 运行 protoc 命令生成基础代码
	args := []string{
		"--go_out=" + outDir,
		"--go-grpc_out=" + outDir,
		protoFile,
	}
	cmd := execx.WrapCmd("protoc", args)
	if _, err := execx.Run(cmd, ""); err != nil {
		color.Red("执行 protoc 命令失败: %v", err)
		return err
	}

	// 3. 解析 proto 文件获取服务信息
	protoContent, err := os.ReadFile(protoFile)
	if err != nil {
		color.Red("读取 proto 文件失败: %v", err)
		return err
	}

	pbiInfo, err := parseProtoFile(string(protoContent))
	if err != nil {
		color.Red("解析 proto 文件失败: %v", err)
		return err
	}

	if pbiInfo.ServiceName == "" {
		return nil
	}

	// 生成 handler 文件
	if err := genHandlerFile(pbiInfo, outDir, protoFile); err != nil {
		return err
	}

	// 生成 endpoint 相关文件
	if err := genEndpointFiles(pbiInfo, outDir); err != nil {
		return err
	}

	return nil
}

// 生成 handler 文件
func genHandlerFile(pbiInfo *PbInfo, outDir, protoFile string) error {
	fileName := strings.TrimSuffix(filepath.Base(protoFile), ".proto")
	var handlerFile string

	if pbiInfo.GoPackage != "" {
		pkgPath := strings.TrimPrefix(pbiInfo.GoPackage, "./")
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
	if pbiInfo.EndpointPath != "" {
		// 如果 EndpointPath 是相对路径，则使用 out 参数作为根路径
		if !filepath.IsAbs(pbiInfo.EndpointPath) {
			endpointDir = filepath.Join(outDir, pbiInfo.EndpointPath)
		} else {
			endpointDir = pbiInfo.EndpointPath
		}
	} else {
		endpointDir = filepath.Join(outDir, "endpoint/api", strings.ToLower(pbiInfo.ServiceName))
	}

	if err := os.MkdirAll(endpointDir, 0755); err != nil {
		color.Red("创建 endpoint 目录失败: %v", err)
		return err
	}

	// 生成 endpoint 文件，如果存在则跳过
	endpointFile := filepath.Join(endpointDir, strings.ToLower(pbiInfo.ServiceName+"endpoint.go"))
	if err := genFileFromTemplate(endpointFile, endpointTmpl, pbiInfo, false); err != nil {
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
		color.Yellow("文件已存在，跳过生成: %s", absPath)
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

	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		color.Red("写入文件失败: %v", err)
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

func parseProtoFile(content string) (*PbInfo, error) {
	reader := strings.NewReader(content)
	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		color.Red("解析proto文件失败: %v", err)
		return nil, err
	}

	data := &PbInfo{}

	// 遍历proto定义
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
			}
		}),
		proto.WithService(func(s *proto.Service) {
			data.ServiceName = s.Name
			for _, element := range s.Elements {
				if rpc, ok := element.(*proto.RPC); ok {
					method := Method{
						Name:       rpc.Name,
						FieldName:  strings.ToLower(rpc.Name),
						InputType:  rpc.RequestType,
						OutputType: rpc.ReturnsType,
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
				}
			}
		}
	}

	if data.ServiceName == "" {
		err := fmt.Errorf("未找到service定义")
		color.Red(err.Error())
		return nil, err
	}

	return data, nil
}
