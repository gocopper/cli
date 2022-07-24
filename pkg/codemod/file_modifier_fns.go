package codemod

import (
	"encoding/json"
	"regexp"
	"strings"
	"text/template"

	"github.com/gocopper/copper/cerrors"
)

type FileModifierFn func(contents string, data map[string]string) (string, error)

func ModReplaceRegex(expr, text string) FileModifierFn {
	return func(contents string, data map[string]string) (string, error) {
		m, err := regexp.Compile(expr)
		if err != nil {
			return "", cerrors.New(err, "failed to compile regex", nil)
		}

		return m.ReplaceAllString(contents, text), nil
	}
}

func ModInsertText(text string, offset int) FileModifierFn {
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

func ModInsertLineAfter(find, insert string) FileModifierFn {
	return func(contents string, data map[string]string) (string, error) {
		return strings.Replace(contents, find, find+"\n"+insert+"\n", 1), nil
	}
}

func ModAppendText(text string) FileModifierFn {
	return func(contents string, data map[string]string) (string, error) {
		return contents + text, nil
	}
}

func ModAddProviderToWireSet(provider string) FileModifierFn {
	return func(contents string, data map[string]string) (string, error) {
		return ModReplaceRegex("wire\\.NewSet\\(", "wire.NewSet(\n"+provider+",\n")(contents, data)
	}
}

func ModAddGoImports(imports []string) FileModifierFn {
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

		return ModInsertText(strings.Join(importStmts, "\n")+"\n", pos+len(pkgImportStmt)+1)(contents, data)
	}
}

func ModAddJSONSection(section string, sectionContents interface{}) FileModifierFn {
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

type ModInsertCHTTPRouteParams struct {
	Path        string
	Method      string
	HandlerName string
	HandlerBody string
}

func ModInsertCHTTPRoute(p ModInsertCHTTPRouteParams) FileModifierFn {
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

		contents, err := ModInsertLineAfter("[]chttp.Route{", routeOut.String())(contents, data)
		if err != nil {
			return "", err
		}

		contents, err = ModAppendText(handlerOut.String())(contents, data)
		if err != nil {
			return "", err
		}

		return contents, nil
	}
}
