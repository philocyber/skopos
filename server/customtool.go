package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Philocyber/skopos/locale"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Philocyber/skopos/agent"
	"github.com/Philocyber/skopos/db"
	"github.com/Autumn-27/norma/permission"
	actool "github.com/Autumn-27/norma/tool"
)

// Custom tool executors. Rows with system=false are dispatched by kind:
// command renders a command and reuses Bash execution; script supports Python only, writing a temporary file,
// passing JSON on stdin and TOOL_* variables in the environment with the configured interpreter; http uses native requests and optional proxying.
// Like traffic/orchestration tools, these enter through hostTools and binding filters; no seeding is needed because they already exist in tools.

// Custom tool CRUD.

type customToolReq struct {
	Key         string          `json:"key"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
	Agents      []string        `json:"agents"`
	Enabled     bool            `json:"enabled"`
	Kind        string          `json:"kind"` // command | script | http
	Exec        json.RawMessage `json:"exec"`
	Deferred    bool            `json:"deferred"`
}

var reToolKey = reAgentKey // Same rules as agent keys: lowercase initial letter, followed by lowercase letters, digits, or underscores.

func (s *Server) pgCreateCustomTool(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	var req customToolReq
	if err := decode(r, &req); err != nil {
		writeError(w, 400, err)
		return
	}
	req.Key = strings.TrimSpace(req.Key)
	if !reToolKey.MatchString(req.Key) {
		writeErr(w, 400, "key must start with a lowercase letter and contain only lowercase letters, digits, or underscores")
		return
	}
	if req.Kind != "command" && req.Kind != "script" && req.Kind != "http" && req.Kind != "shell" {
		writeErr(w, 400, "kind must be command / script / http / shell")
		return
	}
	if req.Kind == "http" && !hasSchemaProps(req.Schema) {
		writeErr(w, 400, "HTTP tools require a nonempty parameter JSON Schema")
		return
	}
	if exist, _ := pg.GetTool(req.Key); exist != nil {
		writeErr(w, 409, "This key already exists as a built-in or custom tool")
		return
	}
	if err := pg.CreateCustomTool(&db.Tool{
		Key: req.Key, Description: req.Description, Schema: req.Schema, Agents: req.Agents,
		Enabled: req.Enabled, Kind: req.Kind, Exec: req.Exec, Deferred: req.Deferred,
	}); err != nil {
		writeError(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"key": req.Key})
}

func (s *Server) pgUpdateCustomTool(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	key := r.PathValue("key")
	existing, err := pg.GetTool(key)
	if err != nil {
		writeError(w, 500, err)
		return
	}
	if existing == nil || existing.System {
		writeErr(w, 400, "Only custom tools can be edited")
		return
	}
	var req customToolReq
	if err := decode(r, &req); err != nil {
		writeError(w, 400, err)
		return
	}
	if req.Kind != "command" && req.Kind != "script" && req.Kind != "http" && req.Kind != "shell" {
		writeErr(w, 400, "kind must be command / script / http / shell")
		return
	}
	if req.Kind == "http" && !hasSchemaProps(req.Schema) {
		writeErr(w, 400, "HTTP tools require a nonempty parameter JSON Schema")
		return
	}
	if err := pg.UpdateCustomTool(&db.Tool{
		Key: key, Description: req.Description, Schema: req.Schema, Agents: req.Agents,
		Enabled: req.Enabled, Kind: req.Kind, Exec: req.Exec, Deferred: req.Deferred,
	}); err != nil {
		writeError(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) pgDeleteCustomTool(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	key := r.PathValue("key")
	if err := pg.DeleteCustomTool(key); err != nil {
		writeError(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"deleted": key})
}

// testToolReq is a dry-run request from the editor: run the given (possibly unsaved)
// exec spec with sample params, without persisting the tool. Same executor path as a
// real tool call — it runs arbitrary command/script/http on the server, which the
// custom-tool feature already allows, so no new capability is granted.
type testToolReq struct {
	Kind   string          `json:"kind"` // command | script | http
	Exec   json.RawMessage `json:"exec"`
	Params map[string]any  `json:"params"`
}

// pgTestCustomTool executes an exec spec once and returns its raw output + error
// flag, so the editor can debug a tool before saving it. Per-kind timeouts still
// apply from the exec spec (with defaults); the outer ceiling is a hard backstop.
func (s *Server) pgTestCustomTool(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	var req testToolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "Invalid request body")
		return
	}
	params := req.Params
	if params == nil {
		params = map[string]any{}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()
	tc := &actool.ToolContext{WorkingDir: s.m.dir} // run in the project dir, like a real call
	var res actool.Result
	switch req.Kind {
	case "command":
		res, _ = s.runCommandTool(ctx, req.Exec, params, tc)
	case "script":
		res, _ = s.runScriptTool(ctx, "test", req.Exec, params, tc)
	case "http":
		res, _ = s.runHTTPTool(ctx, req.Exec, params, tc)
	case "shell":
		writeErr(w, 400, "Shell tools declare the Bash environment and have no executable content")
		return
	default:
		writeError(w, 400, locale.Errorf("Unknown tool kind: %s", req.Kind))
		return
	}
	writeJSON(w, 200, map[string]any{"output": res.Flatten(), "is_error": res.IsError})
}

// Python interpreter detection, persistence, and overrides.

const settingPythonInterp = "python_interpreter"

// detectPython finds a python interpreter absolute path (python3 preferred).
func detectPython() string {
	for _, c := range []string{"python3", "python"} {
		if p, err := exec.LookPath(c); err == nil {
			return p
		}
	}
	return ""
}

// pythonInterpreter resolves the interpreter: user-set > stored auto-detect > live
// detect. "" only when truly none found.
func (s *Server) pythonInterpreter() string {
	if v, ok, _ := s.m.pg.GetSetting(settingPythonInterp); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return detectPython()
}

// seedPythonInterpreter stores the auto-detected interpreter on startup if unset
// (never clobbers a user-set value).
func (s *Server) seedPythonInterpreter() {
	if v, ok, _ := s.m.pg.GetSetting(settingPythonInterp); ok && strings.TrimSpace(v) != "" {
		return
	}
	if p := detectPython(); p != "" {
		_ = s.m.pg.SetSetting(settingPythonInterp, p)
		log.Printf(locale.Text(locale.ServerDefault(), "[custom-tool] Automatically detected Python interpreter: %s"), p)
	}
}

// Execution specifications.

type commandExec struct {
	Command   string `json:"command"`
	TimeoutMs int    `json:"timeout_ms"`
}
type scriptExec struct {
	Code      string `json:"code"`
	TimeoutMs int    `json:"timeout_ms"`
}
type httpExec struct {
	Method            string            `json:"method"`
	URL               string            `json:"url"`
	Headers           map[string]string `json:"headers"`
	Body              string            `json:"body"`
	TimeoutMs         int               `json:"timeout_ms"`
	Proxy             string            `json:"proxy"`
	UseRecordingProxy bool              `json:"use_recording_proxy"`
}

func timeoutOr(ms, def int) time.Duration {
	if ms <= 0 {
		return time.Duration(def) * time.Millisecond
	}
	return time.Duration(ms) * time.Millisecond
}

// Shared tool construction.

// customTools builds CoreTools for every user-defined (system=false) tool row.
// shell-kind tools are environment hints only — they surface in the Bash tool
// description via ToolResolve and do NOT create callable tool entries here.
func (s *Server) customTools() ([]actool.CoreTool, error) {
	return s.customToolsForLanguage(locale.ServerDefault())
}

func (s *Server) customToolsForLanguage(lang locale.Lang) ([]actool.CoreTool, error) {
	rows, err := s.m.pg.ListCustomTools()
	if err != nil {
		return nil, err
	}
	out := make([]actool.CoreTool, 0, len(rows))
	for _, t := range rows {
		if t.Kind == "shell" {
			continue // shell hints are handled by ToolResolve → Bash description
		}
		out = append(out, s.buildCustomToolForLanguage(t, lang))
	}
	return out, nil
}

// buildCustomTool turns one custom-tool row into a CoreTool. Empty schema → a thin
// {args:string} schema, so command/http templates can use {args}.
func (s *Server) buildCustomTool(t *db.Tool) actool.CoreTool {
	return s.buildCustomToolForLanguage(t, locale.ServerDefault())
}

func (s *Server) buildCustomToolForLanguage(t *db.Tool, lang locale.Lang) actool.CoreTool {
	schema := ensureSchemaForLanguage(t.Schema, lang)
	key, kind, execRaw := t.Key, t.Kind, append(json.RawMessage(nil), t.Exec...)
	declared := map[string]bool{}
	if kind == "command" {
		properties, _ := schema["properties"].(map[string]any)
		for name := range properties {
			declared[name] = true
		}
		// Runtime validation uses this private snapshot rather than mutable
		// model-facing metadata. Extra keys cannot create new substitutions.
		schema["additionalProperties"] = false
	}
	run := func(ctx context.Context, in json.RawMessage, tc *actool.ToolContext) (actool.Result, error) {
		ctx = agent.RequireOperationReview(ctx, key)
		if tc != nil {
			copied := *tc
			copied.Env = append([]string(nil), tc.Env...)
			tc = &copied
		}
		var params map[string]any
		if len(in) > 0 {
			if err := json.Unmarshal(in, &params); err != nil {
				return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), err)), nil
			}
		}
		if params == nil {
			params = map[string]any{}
		}
		if kind == "command" {
			for name := range params {
				if !declared[name] {
					return actool.Errorf(locale.Text(locale.FromContext(ctx), "Command tool parameters must be declared in its schema")), nil
				}
			}
		}
		switch kind {
		case "command":
			return s.runCommandTool(ctx, execRaw, params, tc)
		case "script":
			return s.runScriptTool(ctx, key, execRaw, params, tc)
		case "http":
			return s.runHTTPTool(ctx, execRaw, params, tc)
		default:
			return actool.Errorf(locale.Text(locale.FromContext(ctx), "Unknown custom tool kind: %s", kind)), nil
		}
	}
	return &customOperationTool{CoreTool: actool.Build(actool.Spec{
		Name: key, Description: t.Description, Schema: schema,
		Permissions: func(context.Context, json.RawMessage, permission.Context) permission.Decision {
			return permission.Allowed()
		},
		Run: run,
	})}
}

type customOperationTool struct{ actool.CoreTool }

func (*customOperationTool) RequiresOperationReview() bool { return true }

// hasSchemaProps reports whether raw is a JSON-Schema object with ≥1 property.
// http tools require an explicit schema (the auto {args} shell can't name the
// {param} placeholders in URL/headers/body), so an empty schema is rejected.
func hasSchemaProps(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var m map[string]any
	if json.Unmarshal(raw, &m) != nil {
		return false
	}
	props, _ := m["properties"].(map[string]any)
	return len(props) > 0
}

// ensureSchema returns the tool's schema, or a thin {args:string} when none given.
func ensureSchema(raw json.RawMessage) map[string]any {
	return ensureSchemaForLanguage(raw, locale.ServerDefault())
}

// User-authored schema is returned unchanged; only the generated fallback is localized.
func ensureSchemaForLanguage(raw json.RawMessage, lang locale.Lang) map[string]any {
	var m map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	}
	props, _ := m["properties"].(map[string]any)
	if len(props) > 0 {
		return m
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"args": map[string]any{"type": "string", "description": locale.Text(lang, "Command/arguments (free text)")},
		},
	}
}

// command: render the command and reuse Bash execution.

func (s *Server) runCommandTool(ctx context.Context, execRaw json.RawMessage, params map[string]any, tc *actool.ToolContext) (actool.Result, error) {
	childContext := actool.ToolContext{}
	if tc != nil {
		childContext = *tc
	}
	childContext.Env = agent.ExecutorEnvironmentOverrides(childContext.Env)
	tc = &childContext
	var spec commandExec
	_ = json.Unmarshal(execRaw, &spec)
	if strings.TrimSpace(spec.Command) == "" {
		return actool.Errorf(locale.Text(locale.FromContext(ctx), "command is empty")), nil
	}
	var environment templateExecutionEnvironment
	if tc != nil {
		environment.dir, environment.env = tc.WorkingDir, tc.Env
	}
	cmd, err := renderCommandTemplateContext(ctx, spec.Command, params, environment)
	if err != nil {
		return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), err)), nil
	}
	operation := map[string]any{"kind": "command", "command": cmd, "params": params, "timeout_ms": timeoutOr(spec.TimeoutMs, 120000).Milliseconds()}
	if tc != nil {
		operation["working_directory"] = tc.WorkingDir
		operation["session_env"] = tc.Env
	}
	if err := agent.ReviewOperation(ctx, "Bash", operation, tc); err != nil {
		return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), err)), nil
	}
	// Reuse Bash's executor through Bash CoreTool.Call to inherit safety checks, timeouts,
	// proxy environment, and output overflow handling. This is a peer tool sharing execution, not a model-invoked Bash wrapper.
	bashIn, _ := json.Marshal(map[string]any{"command": cmd})
	if spec.TimeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeoutOr(spec.TimeoutMs, 120000))
		defer cancel()
	}
	return actool.NewBash().Call(ctx, bashIn, tc)
}

// script (Python only): temporary file, JSON stdin, and environment variables.

func (s *Server) runScriptTool(ctx context.Context, key string, execRaw json.RawMessage, params map[string]any, tc *actool.ToolContext) (actool.Result, error) {
	var spec scriptExec
	_ = json.Unmarshal(execRaw, &spec)
	if strings.TrimSpace(spec.Code) == "" {
		return actool.Errorf(locale.Text(locale.FromContext(ctx), "script code is empty")), nil
	}
	interp := s.pythonInterpreter()
	if interp == "" {
		return actool.Errorf(locale.Text(locale.FromContext(ctx), "No Python interpreter is configured or detected; set one in system settings")), nil
	}
	workDir := s.m.dir
	var sessionEnv []string
	if tc != nil {
		if tc.WorkingDir != "" {
			workDir = tc.WorkingDir
		}
		sessionEnv = tc.Env
	}
	plan, err := preparePython(interp, key, spec.Code, params, workDir, sessionEnv, timeoutOr(spec.TimeoutMs, 120000))
	if err != nil {
		return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), err)), nil
	}
	if err := agent.ReviewOperation(ctx, "Bash", plan.review, tc); err != nil {
		return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), err)), nil
	}
	body, err := plan.execute(ctx)
	if err != nil {
		return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), err)), nil
	}
	return actool.Text(actool.Capture(tc, body)), nil
}

// execPython writes the code to a temp .py under workDir/.tools, runs it via interp
// with the params JSON on stdin + scalar params mirrored to TOOL_<NAME> env, and
// returns combined stdout+stderr (with a timeout/exit note). Standalone + testable.
func execPython(ctx context.Context, interp, key, code string, params map[string]any, workDir string, sessionEnv []string, timeout time.Duration) (string, error) {
	plan, err := preparePython(interp, key, code, params, workDir, sessionEnv, timeout)
	if err != nil {
		return "", err
	}
	return plan.execute(ctx)
}

type preparedPython struct {
	interpreter, path, code, dir string
	env                          []string
	stdin                        []byte
	timeout                      time.Duration
	review                       map[string]any
}

func preparePython(interp, key, code string, params map[string]any, workDir string, sessionEnv []string, timeout time.Duration) (*preparedPython, error) {
	resolved, err := exec.LookPath(interp)
	if err != nil {
		return nil, err
	}
	interp, err = filepath.Abs(resolved)
	if err != nil {
		return nil, err
	}
	stdin, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	paramsEnv := []string{}
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if value, ok := scalarStr(params[key]); ok {
			paramsEnv = append(paramsEnv, "TOOL_"+strings.ToUpper(key)+"="+value)
		}
	}
	sessionEnv = agent.ExecutorEnvironmentOverrides(sessionEnv)
	env := agent.ExecutorEnvironment(sessionEnv)
	env = append(env, paramsEnv...)
	envJSON, _ := json.Marshal(env)
	digest := sha256.Sum256(envJSON)
	plan := &preparedPython{interpreter: interp, path: filepath.Join(workDir, ".tools", key+"-"+rand.Text()+".py"), code: code, dir: workDir, env: env, stdin: stdin, timeout: timeout}
	plan.review = map[string]any{"kind": "script", "interpreter": interp, "argv": []string{plan.path}, "code": code, "stdin": string(stdin), "params": params, "working_directory": workDir, "session_env": append([]string{}, sessionEnv...), "parameter_env": paramsEnv, "environment_sha256": fmt.Sprintf("%x", digest), "inherits_host_environment": true, "timeout_ms": timeout.Milliseconds()}
	return plan, nil
}
func (plan *preparedPython) execute(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	toolsDir := filepath.Dir(plan.path)
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return "", err
	}
	f, err := os.OpenFile(plan.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err := f.WriteString(plan.code); err != nil {
		f.Close()
		return "", err
	}
	f.Close()

	runCtx, cancel := context.WithTimeout(ctx, plan.timeout)
	defer cancel()
	c := exec.CommandContext(runCtx, plan.interpreter, tmp)
	c.Dir = plan.dir
	c.Env = plan.env
	c.Stdin = bytes.NewReader(plan.stdin)
	out, err := c.CombinedOutput()
	body := string(out)
	if runCtx.Err() == context.DeadlineExceeded {
		body += locale.Text(locale.FromContext(ctx), "\n... [terminated after timeout] ...")
	} else if err != nil {
		body += locale.Text(locale.FromContext(ctx), "\n[exit: %s]", err.Error())
	}
	return body, nil
}

// http: native requests and proxy support.

func (s *Server) runHTTPTool(ctx context.Context, execRaw json.RawMessage, params map[string]any, tc *actool.ToolContext) (actool.Result, error) {
	var spec httpExec
	_ = json.Unmarshal(execRaw, &spec)
	method := strings.ToUpper(strings.TrimSpace(spec.Method))
	if method == "" {
		method = "GET"
	}
	rawURL := renderTemplate(spec.URL, params, identity)
	if strings.TrimSpace(rawURL) == "" {
		return actool.Errorf(locale.Text(locale.FromContext(ctx), "HTTP URL is empty")), nil
	}
	body := ""
	var bodyReader io.Reader
	if spec.Body != "" {
		body = renderTemplate(spec.Body, params, identity)
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), err)), nil
	}
	for k, v := range spec.Headers {
		req.Header.Set(k, renderTemplate(v, params, identity))
	}
	client := &http.Client{Timeout: timeoutOr(spec.TimeoutMs, 30000)}
	transport := s.httpProxyTransport(spec)
	if transport == nil {
		standard, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return actool.Errorf(locale.Text(locale.FromContext(ctx), "Cannot resolve the default HTTP transport for operation review")), nil
		}
		transport = standard.Clone()
	}
	defer transport.CloseIdleConnections()
	client.Transport = transport
	checkDestination := func(next *http.Request) error {
		// Native HTTP owns the actual destination, including each redirect. Asset
		// policy applies independently of optional intent metadata and tool rules.
		taskID := agent.RunInfoFrom(next.Context()).TaskID
		if s.m != nil && s.m.Assets() != nil {
			if err := s.m.Assets().CheckHTTPDestination(locale.FromContext(next.Context()), taskID, next.URL); err != nil {
				return err
			}
		} else if taskID > 0 {
			return locale.NewError("HTTP execution destination policy is unavailable")
		}
		return nil
	}
	review := func(next *http.Request, hop int) error {
		if err := checkDestination(next); err != nil {
			return err
		}
		proxy := ""
		if transport.Proxy != nil {
			selected, err := transport.Proxy(next)
			if err != nil {
				return err
			}
			if selected != nil {
				proxy = selected.String()
			}
		}
		nextBody := body
		if next.Body == nil {
			nextBody = ""
		}
		if err := agent.ReviewOperation(next.Context(), "WebFetch", map[string]any{"kind": "http", "method": next.Method, "url": next.URL.String(), "headers": next.Header.Clone(), "body": nextBody, "proxy": proxy, "use_recording_proxy": spec.UseRecordingProxy, "timeout_ms": timeoutOr(spec.TimeoutMs, 30000).Milliseconds(), "redirect_hop": hop, "params": params}, tc); err != nil {
			return err
		}
		// A review can await human approval. Re-read asset policy after it returns
		// so approval of an operation cannot override a newly revoked target grant.
		return checkDestination(next)
	}
	if err := review(req, 0); err != nil {
		return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), err)), nil
	}
	if s.m != nil && s.m.Assets() != nil {
		var recordingProxy *url.URL
		if s.m.traffic != nil {
			recordingProxy, _ = url.Parse(s.m.traffic.ProxyAddr())
		}
		peerTransport := &httpPeerTransport{base: transport, recordingProxy: recordingProxy, destination: checkDestination, proxyDestination: func(next *http.Request) error {
			return s.m.Assets().CheckHTTPProxyDestination(locale.FromContext(next.Context()), agent.RunInfoFrom(next.Context()).TaskID, next.URL)
		}, check: func(ctx context.Context, ip net.IP) error {
			return s.m.Assets().CheckHTTPPeer(locale.FromContext(ctx), agent.RunInfoFrom(ctx).TaskID, ip)
		}}
		if s.m.traffic != nil {
			peerTransport.bindRecording = s.m.traffic.BindNativeHTTPConnection
			peerTransport.recordingPeer = s.m.traffic.OwnsRecordingPeer
		}
		client.Transport = peerTransport
	}
	runCtx, cancel := context.WithTimeout(ctx, timeoutOr(spec.TimeoutMs, 30000))
	defer cancel()
	req = req.WithContext(runCtx)
	var redirectReviewErr error
	client.CheckRedirect = func(next *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
		redirectReviewErr = review(next, len(via))
		return redirectReviewErr
	}
	resp, err := client.Do(req)
	if err != nil {
		var peerErr *httpPeerPolicyError
		if errors.As(err, &peerErr) {
			// Strip net/http's URL wrapper, preserving localized policy errors
			// without putting userinfo/query credentials in tool diagnostics.
			return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), locale.Errorf("Request failed: %w", peerErr.cause))), nil
		}
		if redirectReviewErr != nil {
			// net/http wraps redirect failures in url.Error, which loses the
			// localizable cause and includes the destination in diagnostics.
			return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), locale.Errorf("Request failed: %w", redirectReviewErr))), nil
		}
		return actool.Errorf(locale.ErrorMessage(locale.FromContext(ctx), locale.Errorf("Request failed: %w", err))), nil
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	out := map[string]any{"status": resp.StatusCode, "body": string(respBody)}
	b, _ := json.Marshal(out)
	return actool.Text(actool.Capture(tc, string(b))), nil
}

// httpProxyTransport builds a Transport for the http tool's proxy config, or nil
// (direct). use_recording_proxy routes through the recording proxy + trusts its CA.
func (s *Server) httpProxyTransport(spec httpExec) *http.Transport {
	proxyStr := strings.TrimSpace(spec.Proxy)
	var caFile string
	if spec.UseRecordingProxy {
		if addr := s.m.ProxyAddr(); addr != "" {
			proxyStr = addr // ProxyAddr already returns a complete proxy URL.
			caFile = s.m.ProxyCACert()
		}
	}
	if proxyStr == "" {
		return nil
	}
	pu, err := url.Parse(proxyStr)
	if err != nil {
		return nil
	}
	tr := &http.Transport{Proxy: http.ProxyURL(pu)}
	if caFile != "" {
		if pem, err := os.ReadFile(caFile); err == nil {
			pool := x509.NewCertPool()
			if pool.AppendCertsFromPEM(pem) {
				tr.TLSClientConfig = &tls.Config{RootCAs: pool}
			}
		}
	}
	return tr
}

// ---------- helpers ----------

func identity(s string) string { return s }

// shellQuote single-quotes a value for safe shell interpolation.
func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// renderTemplate substitutes only placeholders in the original template. Values
// may contain placeholder-like text, but that text must never be reinterpreted.
func renderTemplate(tmpl string, params map[string]any, quote func(string) string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, 2*len(keys))
	for _, k := range keys {
		pairs = append(pairs, "{"+k+"}", quote(valToStr(params[k])))
	}
	return strings.NewReplacer(pairs...).Replace(tmpl)
}

// valToStr renders a param value: scalars as-is, arrays/objects as compact JSON.
func valToStr(v any) string {
	if sv, ok := scalarStr(v); ok {
		return sv
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// scalarStr returns (string, true) for scalar values, ("", false) for arrays/objects.
func scalarStr(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case bool:
		return fmt.Sprintf("%t", x), true
	case float64:
		if x == float64(int64(x)) {
			return fmt.Sprintf("%d", int64(x)), true
		}
		return fmt.Sprintf("%g", x), true
	case nil:
		return "", true
	default:
		return "", false
	}
}
