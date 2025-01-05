package gen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"
)

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
		Value:   "./pb",
	},
}

const handlerTmpl = `package {{ .Package }}

import (
	context "context"

	"github.com/fengjx/luchen"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// New{{ .ServiceName }}Service 返回一个 {{ .ServiceName }}Client
func New{{ .ServiceName }}Service(serverName string) {{ .ServiceName }}Client {
	cli := luchen.GetGRPCClient(
		serverName,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	return New{{ .ServiceName }}Client(cli)
}

type {{ .ServiceName }}Handler interface {
	{{ range .Methods }}
	{{ .Name }}(ctx context.Context, in *{{ .InputType }}) (*{{ .OutputType }}, error)
	{{- end }}
}

type {{ .ServiceName }}Endpoint interface {
	{{ .ServiceName }}Handler
	{{- range .Methods }}
	{{ .Name }}EdnpointDefine() *luchen.EdnpointDefine
	{{- end }}
}

type {{ .ServiceName }}ServiceImpl struct {
	Unimplemented{{ .ServiceName }}Server
	{{- range .Methods }}
	{{ .FieldName }} grpctransport.Handler
	{{- end }}
}

{{ range .Methods }}
func (s *{{ $.ServiceName }}ServiceImpl) {{ .Name }}(ctx context.Context, req *{{ .InputType }}) (*{{ .OutputType }}, error) {
	_, resp, err := s.{{ .FieldName }}.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*{{ .OutputType }}), nil
}
{{ end }}

func Register{{ .ServiceName }}GRPCHandler(gs *luchen.GRPCServer, e {{ .ServiceName }}Endpoint) {
	impl := &{{ .ServiceName }}ServiceImpl{
		{{- range .Methods }}
		{{ .FieldName }}: luchen.NewGRPCTransportServer(
			e.{{ .Name }}EdnpointDefine(),
		),
		{{- end }}
	}
	Register{{ .ServiceName }}Server(gs, impl)
}

func Register{{ .ServiceName }}HTTPHandler(hs *luchen.HTTPServer, e {{ .ServiceName }}Endpoint) {
	{{- range .Methods }}
	{{ .Name }}Def := e.{{ .Name }}EdnpointDefine()
	hs.Mux().Handle({{ .Name }}Def.Path, luchen.NewHTTPTransportServer({{ .Name }}Def))
	{{- end }}
}
`

type Method struct {
	Name       string
	FieldName  string
	InputType  string
	OutputType string
}

type ServiceData struct {
	Package     string
	ServiceName string
	Methods     []Method
}

func action(ctx *cli.Context) error {
	protoFile := ctx.String("file")
	outDir := ctx.String("out")

	// 1. 创建输出目录
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 2. 运行 protoc 命令生成基础代码
	cmd := exec.Command("protoc",
		"--go_out="+outDir,
		"--go_opt=paths=source_relative",
		"--go-grpc_out="+outDir,
		"--go-grpc_opt=paths=source_relative",
		protoFile,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("执行 protoc 命令失败: %v, output: %s", err, string(output))
	}

	// 3. 解析 proto 文件获取服务信息
	protoContent, err := os.ReadFile(protoFile)
	if err != nil {
		return fmt.Errorf("读取 proto 文件失败: %v", err)
	}

	// 简单解析 proto 文件获取服务信息
	serviceData, err := parseProtoFile(string(protoContent))
	if err != nil {
		return fmt.Errorf("解析 proto 文件失败: %v", err)
	}

	// 4. 生成自定义 handler 文件
	tmpl, err := template.New("handler").Parse(handlerTmpl)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, serviceData); err != nil {
		return fmt.Errorf("生成代码失败: %v", err)
	}

	// 5. 写入文件
	fileName := strings.TrimSuffix(filepath.Base(protoFile), ".proto")
	handlerFile := filepath.Join(outDir, fileName+".handler.luchen.go")
	if err := os.WriteFile(handlerFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fmt.Printf("生成代码成功，文件位置: %s\n", handlerFile)
	return nil
}

func parseProtoFile(content string) (*ServiceData, error) {
	lines := strings.Split(content, "\n")
	var data ServiceData

	// 解析 package
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package") {
			data.Package = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "package"))
			data.Package = strings.TrimSuffix(data.Package, ";")
			break
		}
	}

	// 解析 service
	var inService bool
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "service") {
			inService = true
			parts := strings.Fields(line)
			data.ServiceName = strings.TrimSpace(parts[1])
			continue
		}

		if inService {
			if line == "}" {
				break
			}
			if strings.HasPrefix(line, "rpc") {
				method := parseMethod(line)
				data.Methods = append(data.Methods, method)
			}
		}
	}

	return &data, nil
}

func parseMethod(line string) Method {
	// 格式: rpc SayHello(HelloReq) returns (HelloResp) {}
	parts := strings.Fields(line)
	name := parts[1]
	input := strings.Trim(strings.Split(parts[2], "(")[1], ")")
	output := strings.Trim(strings.Split(parts[4], "(")[1], ")")

	return Method{
		Name:       name,
		FieldName:  strings.ToLower(name),
		InputType:  input,
		OutputType: output,
	}
}
