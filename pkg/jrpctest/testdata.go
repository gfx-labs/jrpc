package jrpctest

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
)

//go:embed testdata
var originalTestDataFS embed.FS

var OriginalTestData = &TestData{}

func init() {
	err := fs.WalkDir(originalTestDataFS, "testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		file, err := originalTestDataFS.Open(path)
		if err != nil {
			return err
		}
		OriginalTestData.AddTestData(d.Name(), file)
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (td *TestData) AddTestData(name string, rd io.Reader) *TestData {
	s := bufio.NewScanner(rd)
	t := &TestFile{
		Name: name,
	}
	td.Files = append(td.Files, t)
	currentPair := &TestPair{
		Request: nil,
	}
	var (
		arrowRight = []byte("-->")
		arrowLeft  = []byte("<--")
		comment    = []byte("//")
	)
	for s.Scan() {
		txt := bytes.TrimSpace(s.Bytes())
		if string(txt) == "" {
			continue
		}
		// ignore comments
		if bytes.HasPrefix(txt, comment) {
			continue
		}
		if bytes.HasPrefix(txt, arrowRight) {
			if currentPair.Request != nil {
				t.Pairs = append(t.Pairs, currentPair)
			}
			currentPair = &TestPair{
				Request: nil,
			}
			currentPair.Request = bytes.TrimSpace(bytes.TrimPrefix(txt, arrowRight))
			continue
		}
		if bytes.HasPrefix(txt, arrowLeft) {
			xs := bytes.TrimSpace(bytes.TrimPrefix(txt, arrowLeft))
			currentPair.Responses = append(currentPair.Responses, xs)
			continue
		}
	}
	if currentPair.Request != nil {
		t.Pairs = append(t.Pairs, currentPair)
	}
	return nil
}

type TestData struct {
	Files []*TestFile
}

type TestFile struct {
	Name  string
	Pairs []*TestPair
}

type TestPair struct {
	Request   json.RawMessage
	Responses []json.RawMessage
}
