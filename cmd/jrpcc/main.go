package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"gfx.cafe/open/jrpc"
	"sigs.k8s.io/yaml"
)

func main() {
	scan := bufio.NewReader(os.Stdin)
	ctx := context.Background()
	c := &cli{}
	c.silent(`commands:
  [m]ethod, [d]ial, [s]end, [p]arams, [c]lient, [q]uit
	`)
	for {
		if err := c.tick(ctx, scan); err != nil {
			if !errors.Is(err, context.Canceled) {
				c.silent("%s", err)
			}
			break
		}
	}
}

type cli struct {
	remote  string
	method  string
	params  json.RawMessage
	verbose bool

	conn jrpc.Conn
}

func (c *cli) silent(s string, args ...any) {
	if c.verbose {
		fmt.Printf(s+"\n", args...)
	}
}

func (c *cli) tick(ctx context.Context, scan *bufio.Reader) error {
	cmd, err := scan.ReadString('\n')
	if err != nil {
		return err
	}
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}
	splt := strings.SplitN(cmd, " ", 2)
	if len(splt) == 0 {
		return nil
	}
	switch splt[0] {
	case "d", "dial":
		if len(splt) == 1 {
			c.silent("need to specify url to dial to")
			return nil
		}
		c.conn, err = jrpc.DialContext(ctx, splt[1])
		if err != nil {
			return fmt.Errorf("failed to dial: %w", err)
		}
		c.remote = splt[1]
		c.silent("connected to %s", splt[1])
	case "m", "method":
		if len(splt) == 1 {
			fmt.Println("need to specify method to use")
			return nil
		}
		c.method = splt[1]
		c.silent("method to %s", splt[1])
	case "p", "params":
		nextExit := false
		bts := []byte(splt[1])
		for {
			tmp, err := scan.ReadBytes('\n')
			if err != nil {
				return err
			}
			if len(tmp) == 1 {
				if nextExit {
					break
				}
				nextExit = true
				continue
			}
			bts = append(bts, tmp...)
		}
		var m any
		err = yaml.Unmarshal(bts, &m)
		if err != nil {
			return err
		}
		c.params, _ = json.Marshal(m)
	case "c", "client":
		fmt.Printf(
			`
remote: %s
method: %s
params: %s
`, c.remote, c.method, c.params)
	case "s", "send":
		var res json.RawMessage
		c.conn.Do(ctx, &res, c.method, c.params)
		fmt.Println(string(res))
	case "q", "quit", "exit":
		os.Exit(0)
	}
	return nil
}
