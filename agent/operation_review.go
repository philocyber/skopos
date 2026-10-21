package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Philocyber/skopos/intercept"
	"github.com/Philocyber/skopos/locale"
	"github.com/Autumn-27/norma/agentcore"
	"github.com/Autumn-27/norma/harness"
	"github.com/Autumn-27/norma/llm"
	actool "github.com/Autumn-27/norma/tool"
)

// OperationReviewTool identifies code-owned executors that review their resolved
// operation immediately before execution instead of reviewing abstract params.
type OperationReviewTool interface{ RequiresOperationReview() bool }
type toolUnwrapper interface{ UnwrapTool() actool.CoreTool }
type reviewBindingKey struct{}
type requiredReviewKey struct{}
type operationReviewBinding struct {
	name  string
	hooks harness.HookRunner
}

func RequireOperationReview(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, requiredReviewKey{}, name)
}

// ReviewOperation invokes the existing guard once with the final operation. It
// cannot rewrite an already resolved plan: a changed input requires a new review.
func ReviewOperation(ctx context.Context, executor string, operation any, tc *actool.ToolContext) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	binding, ok := ctx.Value(reviewBindingKey{}).(operationReviewBinding)
	if !ok {
		if name, _ := ctx.Value(requiredReviewKey{}).(string); name != "" {
			return locale.Errorf("Custom operation %s has no runtime review binding", name)
		}
		return nil // Standalone internal executor tests do not constitute agent calls.
	}
	if required, _ := ctx.Value(requiredReviewKey{}).(string); required != "" && required != binding.name {
		return locale.Errorf("Custom operation review binding does not match %s", required)
	}
	if binding.hooks == nil {
		return nil
	} // An explicitly unguarded session.
	input, err := json.Marshal(operation)
	if err != nil {
		return err
	}
	id := ""
	if tc != nil {
		id = tc.ToolUseID
	}
	ctx = intercept.WithResolvedOperation(ctx, id, binding.name, executor, input)
	block, message, updated := binding.hooks.PreToolUse(ctx, binding.name, input)
	if block {
		return fmt.Errorf("%s", message)
	}
	if len(updated) != 0 && string(updated) != string(input) {
		return locale.Errorf("Resolved operation was changed by a hook; execution requires a new review")
	}
	return ctx.Err()
}

// ConfigureOperationReview is applied before each production SDK session is
// created. Custom executors inherit the same hooks, including worker steering;
// the SDK's earlier raw-param hook is skipped only for those executors.
func ConfigureOperationReview(opts *agentcore.Options) {
	opts.BashEnv = ExecutorEnvironmentOverrides(opts.BashEnv)
	if _, ok := opts.Hooks.(*operationHooks); ok {
		return
	}
	if opts.Tools != nil {
		opts.Tools = append([]actool.CoreTool{}, opts.Tools...)
	}
	byName := make(map[string]actool.CoreTool)
	custom := make(map[string]bool)
	for _, tool := range opts.Tools {
		byName[tool.Name()] = tool
	}
	// NewRegistry uses the last handler for each name. SDK-added handlers also
	// replace host tools, so their raw hooks must never inherit a custom exemption.
	for name := range sdkOperationOverrides(opts) {
		delete(byName, name)
	}
	for name, tool := range byName {
		custom[name] = requiresOperationReview(tool)
	}
	for i, tool := range opts.Tools {
		if requiresOperationReview(tool) {
			opts.Tools[i] = &operationBoundTool{CoreTool: tool, hooks: opts.Hooks}
		}
	}
	if opts.Hooks != nil {
		opts.Hooks = &operationHooks{inner: opts.Hooks, tools: byName, custom: custom}
	}
}

func requiresOperationReview(tool actool.CoreTool) bool {
	for inner := tool; inner != nil; {
		if marker, ok := inner.(OperationReviewTool); ok && marker.RequiresOperationReview() {
			return true
		}
		unwrap, ok := inner.(toolUnwrapper)
		if !ok {
			break
		}
		inner = unwrap.UnwrapTool()
	}
	return false
}

// These registrations follow norma v0.4.3 NewSession. Keep this list aligned
// when upgrading the SDK; a replaced handler must retain its ordinary hook.
func sdkOperationOverrides(opts *agentcore.Options) map[string]bool {
	names := make(map[string]bool)
	add := func(name string) { names[name] = true }
	if opts.Plan != nil {
		add("EnterPlanMode")
		add("ExitPlanMode")
	}
	if opts.Memory != nil && opts.Memory.Store != nil {
		add("RecallMemory")
		add("RecordMemory")
	}
	if opts.Skills != nil {
		add(opts.Skills.Tool().Name())
	}
	if opts.EnableWebFetch {
		add("WebFetch")
	}
	if opts.EnableWebSearch {
		if tool, err := actool.NewWebSearch(actool.WebSearchConfig{Backend: opts.WebSearchBackend, BraveAPIKey: opts.BraveSearchAPIKey, TavilyAPIKey: opts.TavilySearchAPIKey, DeepSeekBaseURL: opts.DeepSeekSearchBaseURL, DeepSeekAPIKey: opts.DeepSeekSearchAPIKey, DeepSeekModel: opts.DeepSeekSearchModel, Proxy: opts.WebSearchProxy, CACert: opts.WebSearchCACert, InsecureTLS: opts.WebSearchInsecureTLS}); err == nil {
			add(tool.Name())
		}
	}
	if opts.Todos != nil {
		add(opts.Todos.Tool().Name())
	}
	if opts.AskUser != nil {
		add("AskUserQuestion")
	}
	if !actool.BackgroundTasksDisabled() && !opts.DisableBackgroundTasks {
		// Manager creation can fail. In that case a same-name host tool keeps
		// both reviews, rather than risking an unreviewed SDK management tool.
		add("TaskOutput")
		add("TaskStop")
		add("TaskList")
		add("Monitor")
	}
	if len(opts.DeferredTools) > 0 {
		add(actool.ExecuteExtraToolName)
		add("SearchExtraTools")
	}
	return names
}

type operationBoundTool struct {
	actool.CoreTool
	hooks harness.HookRunner
}

func (t *operationBoundTool) Call(ctx context.Context, input json.RawMessage, tc *actool.ToolContext) (actool.Result, error) {
	if _, ok := ctx.Value(reviewBindingKey{}).(operationReviewBinding); !ok {
		ctx = context.WithValue(ctx, reviewBindingKey{}, operationReviewBinding{name: t.Name(), hooks: t.hooks})
	}
	return t.CoreTool.Call(ctx, input, tc)
}
func (t *operationBoundTool) UnwrapTool() actool.CoreTool { return t.CoreTool }

type operationHooks struct {
	inner  harness.HookRunner
	tools  map[string]actool.CoreTool
	custom map[string]bool
}

func (h *operationHooks) PreToolUse(ctx context.Context, name string, input []byte) (bool, string, []byte) {
	if h.custom[name] {
		return false, "", nil
	}
	if name == actool.ExecuteExtraToolName {
		var invocation struct {
			Name string `json:"tool_name"`
		}
		if json.Unmarshal(input, &invocation) == nil && h.custom[invocation.Name] {
			return false, "", nil
		}
	}
	effective := json.RawMessage(input)
	for tool := h.tools[name]; tool != nil; {
		if decorator, ok := tool.(*overriddenTool); ok {
			effective = injectDefaults(effective, decorator.schema)
		}
		unwrap, ok := tool.(toolUnwrapper)
		if !ok {
			break
		}
		tool = unwrap.UnwrapTool()
	}
	if string(effective) != string(input) {
		ctx = intercept.WithEffectiveInput(ctx, name, input, effective)
	}
	block, message, updated := h.inner.PreToolUse(ctx, name, effective)
	if !block && len(updated) == 0 && string(effective) != string(input) {
		updated = effective
	}
	return block, message, updated
}
func (h *operationHooks) PostToolUse(ctx context.Context, name string, input, result []byte, isErr bool) {
	h.inner.PostToolUse(ctx, name, input, result, isErr)
}
func (h *operationHooks) Stop(ctx context.Context, messages []llm.Message) (bool, []string, string) {
	return h.inner.Stop(ctx, messages)
}

// Preserve decorators' discovery of the underlying custom executor/defaults.
func (g *guardedTool) UnwrapTool() actool.CoreTool    { return g.CoreTool }
func (t *overriddenTool) UnwrapTool() actool.CoreTool { return t.CoreTool }
