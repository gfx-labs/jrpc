package generate

import (
	"bytes"
	"fmt"
	"gfx.cafe/open/jrpc/openrpc/types"
	"github.com/iancoleman/strcase"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

var defaultTemplates = []string{
	"./templates/types.gotmpl",
}

var funcs = template.FuncMap{
	"list": func(v ...any) []any {
		return v
	},
	"camelCase": func(v string) string {
		return strcase.ToCamel(v)
	},
	"goType": func(v string) string {
		switch v {
		case "boolean":
			return "bool"
		case "number":
			return "float64"
		case "string":
			return "string"
		case "null":
			return "struct{}"
		default:
			panic(fmt.Sprintln("unknown go type:", v))
		}
	},
	"refName": func(v string) string {
		return filepath.Base(v)
	},
}

func Generate(rpc *types.OpenRPC, templates []string, output string) error {
	if len(templates) == 0 {
		templates = defaultTemplates
	}

	var wr bytes.Buffer
	for _, tmpl := range templates {
		name := filepath.Base(tmpl)

		t, err := template.New(name).Funcs(funcs).ParseFiles(tmpl)
		if err != nil {
			return err
		}

		err = t.Execute(&wr, rpc)
		if err != nil {
			return err
		}

		var fmtd []byte
		fmtd, err = format.Source(wr.Bytes())
		if err != nil {
			return err
		}
		wr.Reset()

		name = name[:len(name)-len(filepath.Ext(name))]
		err = os.WriteFile(filepath.Join(output, name+".go"), fmtd, 0777)
		if err != nil {
			return err
		}
	}

	return nil
}
