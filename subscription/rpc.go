package subscription

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type packet struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
type rpc struct {
	cmd      *exec.Cmd
	in       io.WriteCloser
	messages chan packet
	done     chan struct{}
	scanErr  error
	queue    []packet
	seq      int
	once     sync.Once
	cancel   context.CancelFunc
}

func runtimeEnv(dir string) []string {
	env := []string{}
	for _, e := range os.Environ() {
		key, _, _ := strings.Cut(e, "=")
		if strings.HasPrefix(key, "CODEX_") || strings.HasPrefix(key, "OPENAI_") || strings.HasPrefix(key, "CHATGPT_") {
			continue
		}
		env = append(env, e)
	}
	return append(env, "CODEX_HOME="+dir)
}
func (m *Manager) launch(ctx context.Context, dir string) (*rpc, error) {
	ctx, cancel := context.WithCancel(ctx)
	binary := os.Getenv("SKOPOS_CODEX_BINARY")
	if binary == "" {
		binary = "codex"
	}
	cmd := exec.CommandContext(ctx, binary, "app-server", "--listen", "stdio://")
	if m.command != nil {
		cmd = m.command(ctx)
	}
	cmd.Env = append(runtimeEnv(dir), cmd.Env...)
	cmd.Dir = dir
	cmd.Stderr = io.Discard
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		cancel()
		return nil, errors.New("Could not start Codex. Install the Codex CLI on the backend host, or set SKOPOS_CODEX_BINARY")
	}
	c := &rpc{cmd: cmd, in: in, messages: make(chan packet, 256), done: make(chan struct{}), cancel: cancel}
	go func() {
		defer close(c.done)
		defer close(c.messages)
		scanner := bufio.NewScanner(out)
		scanner.Buffer(make([]byte, 4096), 8<<20)
		for scanner.Scan() {
			var p packet
			if json.Unmarshal(scanner.Bytes(), &p) != nil {
				continue
			}
			select {
			case c.messages <- p:
			case <-ctx.Done():
				return
			}
		}
		c.scanErr = scanner.Err()
	}()
	if _, err = c.call(ctx, "initialize", map[string]any{"clientInfo": map[string]any{"name": "skopos", "title": "Skopos", "version": "1"}, "capabilities": map[string]any{"experimentalApi": true}}); err != nil {
		c.close()
		return nil, err
	}
	if err = c.send(map[string]any{"method": "initialized", "params": map[string]any{}}); err != nil {
		c.close()
		return nil, err
	}
	return c, nil
}
func (c *rpc) send(v any) error { return json.NewEncoder(c.in).Encode(v) }
func (c *rpc) read(ctx context.Context) (packet, error) {
	select {
	case <-ctx.Done():
		return packet{}, ctx.Err()
	case p, ok := <-c.messages:
		if !ok {
			if c.scanErr != nil {
				return packet{}, c.scanErr
			}
			return packet{}, errors.New("Codex runtime stopped before completing the request")
		}
		return p, nil
	}
}
func (c *rpc) next(ctx context.Context) (packet, error) {
	if len(c.queue) > 0 {
		p := c.queue[0]
		c.queue = c.queue[1:]
		return p, nil
	}
	return c.read(ctx)
}
func (c *rpc) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.seq++
	id := c.seq
	if err := c.send(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		return nil, err
	}
	for {
		p, err := c.read(ctx)
		if err != nil {
			return nil, err
		}
		if p.Method == "" && string(p.ID) == fmt.Sprint(id) {
			if p.Error != nil {
				return nil, fmt.Errorf("Codex %s: %s", method, p.Error.Message)
			}
			return p.Result, nil
		}
		if p.Method != "" && len(p.ID) > 0 && p.Method != "item/tool/call" {
			c.deny(p)
			continue
		}
		c.queue = append(c.queue, p)
		if len(c.queue) > 4096 {
			return nil, errors.New("Codex emitted too many events before its RPC response")
		}
	}
}
func (c *rpc) deny(p packet) {
	_ = c.send(map[string]any{"id": p.ID, "error": map[string]any{"code": -32601, "message": "This action must be executed through Skopos tools"}})
}
func (c *rpc) close() {
	c.once.Do(func() { c.cancel(); _ = c.in.Close(); _ = c.cmd.Process.Kill(); _ = c.cmd.Wait() })
}
