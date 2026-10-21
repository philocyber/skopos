import type { Activity } from "./types";

// Keep existing task URLs stable, including bookmarks and approval source links.
export const TASK_WORKSPACES = [
  { value: "sessions", label: "Session", description: "Conversation and agent execution" },
  { value: "overview", label: "Overview", description: "Objectives, scope and constraints" },
  { value: "checks", label: "Results ledger", description: "Recorded conclusions and unfinished exploration" },
  { value: "graph", label: "Exploration paths", description: "Intent and evidence relationships" },
  { value: "broadcast", label: "Broadcast board", description: "The complete exploration record" },
  { value: "findings", label: "Findings", description: "Evidence, severity and review" },
  { value: "retests", label: "Retest", description: "Validation after remediation" },
  { value: "assets", label: "Test assets", description: "Assets associated with this task" },
  { value: "coverage", label: "Asset coverage graph", description: "Assets with recorded results" },
  { value: "intercept", label: "Interception approvals", description: "Pending actions and decisions" },
  { value: "report", label: "Report", description: "Task report and export" },
] as const;

export function validTaskTab(value: string | null): string {
  return TASK_WORKSPACES.some((tab) => tab.value === value) ? value! : "sessions";
}

export function taskWorkspaceHref(id: string, tab = "sessions"): string {
  return `/function/tasks/detail/?${new URLSearchParams({ id, tab: validTaskTab(tab) })}`;
}

// Only a successful tool response can suggest a task. Conversation prose and
// user-supplied identifiers must never masquerade as a created-task event.
export function latestSpawnedTask(activities: Activity[]): string | null {
  const calls = new Map(
    activities.filter((a) => a.kind === "tool_use" && a.tool_use_id).map((a) => [a.tool_use_id, a.tool]),
  );
  for (let i = activities.length - 1; i >= 0; i--) {
    const activity = activities[i];
    if (activity.kind !== "tool_result" || activity.is_error) continue;
    const tool = activity.tool || (activity.tool_use_id ? calls.get(activity.tool_use_id) : undefined);
    if (tool !== "spawn_task") continue;
    for (const text of [activity.detail, activity.summary]) {
      const match = text?.trim().match(/^task created:\s*([1-9]\d*)$/);
      if (match) return match[1];
    }
  }
  return null;
}

export function nodeSummary(payload?: string): string {
  if (!payload) return "No recorded summary";
  try {
    const value = JSON.parse(payload);
    if (value && typeof value === "object") {
      for (const key of ["summary", "text", "name", "description", "conclusion"]) {
        if (typeof value[key] === "string" && value[key].trim()) return value[key];
      }
    }
  } catch {
    /* Legacy nodes may contain plain text. */
  }
  return payload;
}
