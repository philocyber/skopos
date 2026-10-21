import { latestSpawnedTask, nodeSummary, TASK_WORKSPACES, taskWorkspaceHref, validTaskTab } from "./task-workspace.ts";
import assert from "node:assert/strict";
import test from "node:test";

test("task navigation preserves all original workspaces and encodes untrusted IDs", () => {
  for (const tab of [
    "sessions",
    "overview",
    "graph",
    "broadcast",
    "findings",
    "retests",
    "assets",
    "coverage",
    "intercept",
    "report",
  ]) {
    assert.ok(TASK_WORKSPACES.some((item) => item.value === tab));
    assert.equal(validTaskTab(tab), tab);
  }
  assert.equal(validTaskTab("javascript:alert(1)"), "sessions");
  assert.equal(
    taskWorkspaceHref("1&tab=report", "findings"),
    "/function/tasks/detail/?id=1%26tab%3Dreport&tab=findings",
  );
});

test("only a successful spawn_task result can suggest a created task", () => {
  const result = { seq: 2, kind: "tool_result", tool: "spawn_task", summary: "task created: 42" };
  assert.equal(latestSpawnedTask([result]), "42");
  for (const override of [
    { kind: "user" },
    { kind: "text" },
    { is_error: true },
    { tool: "Bash" },
    { summary: "task created: 0" },
    { summary: "task created: 42 and then 43" },
  ]) {
    assert.equal(latestSpawnedTask([{ ...result, ...override }]), null);
  }
  assert.equal(
    latestSpawnedTask([
      { seq: 1, kind: "tool_use", tool: "spawn_task", tool_use_id: "call-1" },
      { seq: 2, kind: "tool_result", tool_use_id: "call-1", summary: "task created: 7" },
    ]),
    "7",
  );
});

test("recorded conclusions are displayed without inventing a security verdict", () => {
  assert.equal(nodeSummary('{"summary":"Request timed out"}'), "Request timed out");
  assert.equal(nodeSummary('{"text":"No issue reproduced in this attempt"}'), "No issue reproduced in this attempt");
  assert.equal(nodeSummary("Legacy raw evidence"), "Legacy raw evidence");
  assert.equal(nodeSummary(), "No recorded summary");
});
