package jrpctest

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"strings"
)

//go:embed testdata
var originalTestDataFS embed.FS

var OriginalTestData = &TestData{}

func init() {
	err := fs.WalkDir(originalTestDataFS, "testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || strings.HasSuffix(d.Name(), ".old") {
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
			currentPair := &TestAction{
				Direction: DirectionSend,
			}
			currentPair.Data = bytes.TrimSpace(bytes.TrimPrefix(txt, arrowRight))
			t.Action = append(t.Action, currentPair)
			continue
		}
		if bytes.HasPrefix(txt, arrowLeft) {
			currentPair := &TestAction{
				Direction: DirectionRecv,
			}
			currentPair.Data = bytes.TrimSpace(bytes.TrimPrefix(txt, arrowLeft))
			t.Action = append(t.Action, currentPair)
			continue
		}
	}
	return nil
}

type TestData struct {
	Files []*TestFile
}

type TestFile struct {
	Name   string
	Action []*TestAction
}

type TestDirection string

const (
	DirectionSend = TestDirection("send")
	DirectionRecv = TestDirection("recv")
)

type TestAction struct {
	Direction TestDirection
	Data      json.RawMessage
}
