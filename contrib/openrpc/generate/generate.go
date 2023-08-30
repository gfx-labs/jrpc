package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"gfx.cafe/open/jrpc/contrib/openrpc/templates"
	"gfx.cafe/open/jrpc/contrib/openrpc/types"
	"gfx.cafe/open/jrpc/contrib/openrpc/util"
)

var funcs = template.FuncMap{
	"list": func(v ...any) []any {
		return v
	},
	"camelCase": util.ToCamel,
	"goType": func(v string) string {
		switch v {
		case "boolean":
			return "bool"
		case "number":
			return "float64"
		case "integer":
			return "int"
		case "string":
			return "string"
		case "null", "object":
			return "struct{}"
		default:
			panic(fmt.Sprintln("unknown go type:", v))
		}
	},
	"refName": filepath.Base,
}

func Generate(rpc *types.OpenRPC, ts string, output string) error {
	var wr bytes.Buffer
	var t *template.Template
	var err error
	if ts == "default" {
		t, err = template.New(path.Base(ts)).Funcs(funcs).Parse(templates.TEMPLATE)
		if err != nil {
			return err
		}
	} else {
		t, err = template.New(path.Base(ts)).Funcs(funcs).ParseFiles(ts)
		if err != nil {
			return err
		}

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

	err = os.WriteFile(output, fmtd, 0600)
	if err != nil {
		return err
	}
	wr.Reset()

	return nil
}
