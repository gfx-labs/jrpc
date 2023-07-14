package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gfx.cafe/open/jrpc/contrib/openrpc/templates"
	"gfx.cafe/open/jrpc/contrib/openrpc/types"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Compile  CompileCommand  `cmd:"" help:"Compile a folder into a single openrpc spec"`
	Generate GenerateCommand `cmd:"" help:"Compile a folder into a single openrpc spec"`
	Template TemplateCommand `cmd:"" help:"print template to stdout"`
}

type TemplateCommand struct {
}

func (c *TemplateCommand) Run() error {
	fmt.Print(templates.TEMPLATE)
	return nil
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
	Spec     string `name:"spec" short:"s" help:"path to jopenrpc spec"`
	Output   string `name:"output" short:"o" help:"output directory and package"`
	Template string `name:"template" short:"t" help:"template to generate with"`
	Package  string `name:"package" short:"p" default:"api" help:"package name"`
}

func (c *GenerateCommand) Run() error {
	if c.Spec == "" {
		return fmt.Errorf("spec file is required")
	}
	openrpc, err := readSpec(c.Spec)
	if err != nil {
		return err
	}
	openrpc.Package = c.Package

	//if err = generate.Generate(openrpc, c.Template, c.Output); err != nil {
	//	return err
	//}

	return nil
}

func readSpec(file string) (out *types.OpenRPC, err error) {
	var data []byte
	data, err = os.ReadFile(file)
	if err != nil {
		return
	}

	out = new(types.OpenRPC)
	err = json.Unmarshal(data, out)
	return
}

func NewCLI() *kong.Context {
	ctx := kong.Parse(&CLI)
	return ctx
}
