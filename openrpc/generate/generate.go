package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"gfx.cafe/open/jrpc/openrpc/types"
	"github.com/iancoleman/strcase"
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

func Generate(rpc *types.OpenRPC, ts string, output string) error {
	var wr bytes.Buffer
	t, err := template.New(path.Base(ts)).Funcs(funcs).ParseFiles(ts)
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

	err = os.WriteFile(output, fmtd, 0777)
	if err != nil {
		return err
	}

	return nil
}
