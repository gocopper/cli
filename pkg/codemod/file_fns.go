package codemod

import (
	"encoding/json"
	"regexp"
	"strings"
	"text/template"

	"github.com/gocopper/copper/cerrors"
)

type FileFn func(contents string, data map[string]string) (string, error)

func ReplaceRegex(expr, text string) FileFn {
	return func(contents string, data map[string]string) (string, error) {
		m, err := regexp.Compile(expr)
		if err != nil {
			return "", cerrors.New(err, "failed to compile regex", nil)
		}

		return m.ReplaceAllString(contents, text), nil
	}
}

func InsertTextAfter(find, insert string) FileFn {
	return func(contents string, data map[string]string) (string, error) {
		return strings.Replace(contents, find, find+insert, 1), nil
	}
}

func InsertText(text string, offset int) FileFn {
	return func(contents string, data map[string]string) (string, error) {
		var (
			newContents = make([]byte, 0)

			beg = contents[:offset]
			end = contents[offset:]
		)

		newContents = append(newContents, beg...)
		newContents = append(newContents, []byte(text)...)
		newContents = append(newContents, end...)

		return string(newContents), nil
	}
}

func InsertLineAtLineNum(lineNum int, text string) FileFn {
	return func(contents string, data map[string]string) (string, error) {
		var (
			newContents = make([]byte, 0)

			lines = strings.Split(contents, "\n")
		)

		for i, line := range lines {
			if i+1 == lineNum {
				newContents = append(newContents, []byte(text+"\n")...)
			}
			newContents = append(newContents, []byte(line+"\n")...)
		}

		return string(newContents), nil
	}
}

func InsertLineAfter(find, insert string) FileFn {
	return func(contents string, data map[string]string) (string, error) {
		return strings.Replace(contents, find, find+"\n"+insert+"\n", 1), nil
	}
}

func AppendText(text string) FileFn {
	return func(contents string, data map[string]string) (string, error) {
		var out strings.Builder

		t, err := template.New(text).Parse(text)
		if err != nil {
			return "", err
		}

		err = t.Execute(&out, data)
		if err != nil {
			return "", err
		}

		return contents + out.String(), nil
	}
}

func AddProviderToWireSet(provider string) FileFn {
	return func(contents string, data map[string]string) (string, error) {
		return ReplaceRegex("wire\\.NewSet\\(", "wire.NewSet(\n"+provider+",\n")(contents, data)
	}
}

func AddGoImports(imports []string) FileFn {
	const pkgImportStmt = "import ("

	return func(contents string, data map[string]string) (string, error) {
		pos := strings.Index(contents, pkgImportStmt)
		if pos == -1 {
			return "", cerrors.New(nil, "failed to find existing 'import' statement in file", nil)
		}

		importStmts := make([]string, 0, len(imports))
		for i := range imports {
			importTmpl, err := template.New(imports[i]).Parse(imports[i])
			if err != nil {
				return "", err
			}

			var out strings.Builder
			err = importTmpl.Execute(&out, data)
			if err != nil {
				return "", err
			}

			d := "\"" + out.String() + "\""
			if strings.HasPrefix(out.String(), "_") {
				d = out.String()
			}

			if strings.Contains(contents, d) {
				continue
			}

			importStmts = append(importStmts, d)
		}

		return InsertText(strings.Join(importStmts, "\n")+"\n", pos+len(pkgImportStmt)+1)(contents, data)
	}
}

func AddJSONSection(section string, sectionContents interface{}) FileFn {
	return func(contents string, data map[string]string) (string, error) {
		var (
			contentsJ map[string]interface{}

			updatedJSONOut strings.Builder
		)

		err := json.Unmarshal([]byte(contents), &contentsJ)
		if err != nil {
			return "", err
		}

		contentsJ[section] = sectionContents

		enc := json.NewEncoder(&updatedJSONOut)
		enc.SetIndent("", "  ")

		err = enc.Encode(&contentsJ)
		if err != nil {
			return "", err
		}

		return updatedJSONOut.String(), nil
	}
}

type InsertCHTTPRouteParams struct {
	Path        string
	Method      string
	HandlerName string
	HandlerBody string
}

func InsertCHTTPRoute(p InsertCHTTPRouteParams) FileFn {
	return func(contents string, data map[string]string) (string, error) {
		var (
			routeOut   strings.Builder
			handlerOut strings.Builder

			routeT = template.Must(template.New("Router#ScaffoldRoute.route").Parse(`{
	Path:        "{{.Path}}",
	Methods:     []string{http.Method{{.Method}}},
	Handler:     ro.{{.HandlerName}},
},
`))
			handlerT = template.Must(template.New("Router#ScaffoldRoute.handler").Parse(`
func (ro *Router) {{.HandlerName}}(w http.ResponseWriter, r *http.Request) {
{{ .HandlerBody }}
}
`))
		)

		if err := routeT.Execute(&routeOut, p); err != nil {
			return "", err
		}

		if err := handlerT.Execute(&handlerOut, p); err != nil {
			return "", err
		}

		contents, err := InsertLineAfter("[]chttp.Route{", routeOut.String())(contents, data)
		if err != nil {
			return "", err
		}

		contents, err = AppendText(handlerOut.String())(contents, data)
		if err != nil {
			return "", err
		}

		return contents, nil
	}
}
