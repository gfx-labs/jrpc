package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"gfx.cafe/open/jrpc/openrpc/generate"
	"gfx.cafe/open/jrpc/openrpc/parse"
	"gfx.cafe/open/jrpc/openrpc/types"
	"github.com/alecthomas/kong"
	"github.com/gobuffalo/packr/v2"
)

var CLI struct {
	Compile  CompileCommand  `cmd:"" help:"Compile a folder into a single openrpc spec"`
	Generate GenerateCommand `cmd:"" help:"Compile a folder into a single openrpc spec"`
}

type CompileCommand struct {
	Methods []string `name:"methods" short:"m" help:"root of method dirs" type:"path"`
	Schemas []string `name:"schemas" short:"s" help:"root schema dirs" type:"path"`
	Output  string   `name:"output" short:"o" help:"path to output file"`
}

func (c *CompileCommand) Run() error {
	openrpc := types.NewOpenRPCSpec1()
	var err error
	for _, v := range c.Methods {
		err = openrpc.AddMethods(v)
		if err != nil {
			return err
		}
	}
	for _, v := range c.Schemas {
		err = openrpc.AddSchemas(v)
		if err != nil {
			return err
		}
	}
	jzn, err := json.MarshalIndent(openrpc, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(c.Output, jzn, 0644)
	if err != nil {
		return err
	}
	return nil
}

type GenerateCommand struct {
	Spec      string   `name:"spec" short:"s" help:"path to jopenrpc spec"`
	Output    string   `name:"output" short:"o" help:"output directory and package"`
	Templates []string `name:"templates" short:"t" help:"list of template types to generate for"`
}

func (c *GenerateCommand) Run() error {
	if c.Spec == "" {
		return fmt.Errorf("spec file is required")
	}
	openrpc, err := readSpec(c.Spec)
	if err != nil {
		return err
	}
	parse.GetTypes(openrpc, openrpc.Objects)
	box := packr.New("template", "./templates")

	if err = generate.WriteFile(box, "server", c.Output, openrpc); err != nil {
		return err
	}
	if err = generate.WriteFile(box, "types", c.Output, openrpc); err != nil {
		return err
	}
	return nil
}

func readSpec(file string) (*types.OpenRPCSpec1, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	spec := types.NewOpenRPCSpec1()
	err = json.Unmarshal(data, spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func NewCLI() *kong.Context {
	ctx := kong.Parse(&CLI)
	return ctx
}
