package jrpctest_test

import (
	"fmt"
	"log"
	"testing"

	"gfx.cafe/open/jrpc/pkg/jrpctest"
)

func TestLoadTestData(t *testing.T) {
	log.Println(jrpctest.OriginalTestData)
	for _, file := range jrpctest.OriginalTestData.Files {
		fmt.Printf("file %s:\n", file.Name)
		for idx, pair := range file.Action {
			fmt.Printf(" %d %s %s\n", idx, pair.Direction, string(pair.Data))
		}
	}
}
