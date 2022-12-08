package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// TODO:
	fmt.Println("generate proto done")
}

const defaultSize = 16

type MethodDeclare struct {
	Method      string
	InputParam  string
	OutputParam string
}

func genNewProtoFile(srcPath, dstPath, content string) error {
	b, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	b = append(b, []byte(content)...)
	return os.WriteFile(dstPath, b, 0644)
}

func genRpcMethodsDeclare(service string, mds []MethodDeclare) string {
	retLines := make([]string, 0, defaultSize)
	startLine := fmt.Sprintf("\n\nservice %s {", service)
	retLines = append(retLines, startLine)

	const prefixTab = "    "
	for i := range mds {
		rpcMethodLine := genRpcMethod(mds[i])
		retLines = append(retLines, prefixTab+rpcMethodLine)
	}

	retLines = append(retLines, "}")
	return strings.Join(retLines, "\n")
}

func genRpcMethod(md MethodDeclare) string {
	tmpl := `rpc %s(%s) returns (%s);`
	return fmt.Sprintf(tmpl, md.Method, md.InputParam, md.OutputParam)
}

func parserCommandLines(lines []string) (string, []MethodDeclare, error) {
	serviceName := ""
	mds := make([]MethodDeclare, 0, defaultSize)
	md := MethodDeclare{}
	for _, line := range lines {
		line := line[2:] // exclude prefix "//"
		startIdx := strings.Index(line, "(")
		if startIdx > 0 {
			_, service, method := parseFullMethod(line[:startIdx])
			if len(serviceName) == 0 {
				serviceName = service
			}
			md.Method = method
		}

		endIdx := strings.Index(line, ")")
		if endIdx == -1 {
			continue
		}

		if startIdx == -1 {
			startIdx = 0
		} else {
			startIdx += 1
		}
		params := strings.Split(line[startIdx:endIdx], ",")
		md.InputParam = params[0]
		md.OutputParam = params[1]
		mds = append(mds, md)
	}

	return serviceName, mds, nil
}

// parseFullMethod returns apinamespace, service and method.
func parseFullMethod(fullMethod string) (string, string, string) {
	items := strings.Split(fullMethod, ".")
	return items[0], items[1], items[2]
}

func readCmdLinesFromProto(path string) ([]string, error) {
	const tag = "readCmdLinesFromProto"
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cmdLines := make([]string, 0, defaultSize)
	isInclude := false
	for _, line := range strings.Split(string(b), "\n") {
		line := strings.Replace(line, " ", "", -1)
		if strings.HasPrefix(line, "//commands") {
			isInclude = true
			continue
		}
		if isInclude {
			if strings.HasSuffix(line, "}") {
				isInclude = false
				break
			}
			cmdLines = append(cmdLines, line)
		}
	}

	if isInclude {
		return nil, fmt.Errorf("%s: no command block end tag found", tag)
	}
	if len(cmdLines) == 0 {
		return nil, fmt.Errorf("%s: no command lines found", tag)
	}
	return cmdLines, nil
}
