// Documentation content (English only). Statements here are grounded in README.md, docs/llm-providers.md and
// the LLM/Agents/Settings pages. Model guidance is selection criteria, not benchmark results.
import {
  BookOpen,
  Brain,
  Compass,
  Cpu,
  LifeBuoy,
  type LucideIcon,
  Rocket,
  ShieldAlert,
  Target,
  Wrench,
} from "lucide-react";

export type DocBlock =
  | { type: "p"; text: string }
  | { type: "list"; items: string[] }
  | { type: "steps"; items: { title: string; body: string; href?: string }[] }
  | { type: "table"; head: string[]; rows: string[][] }
  | { type: "callout"; tone: "info" | "warn"; title: string; text: string }
  | { type: "features"; items: { title: string; href: string; body: string }[] }
  | { type: "faq"; items: { q: string; a: string }[] }
  | { type: "code"; text: string };

export interface DocSection {
  id: string;
  title: string;
  icon: LucideIcon;
  summary: string;
  blocks: DocBlock[];
}

export const DOC_SECTIONS: DocSection[] = [
  {
    id: "overview",
    title: "Overview",
    icon: BookOpen,
    summary: "What Skopos does and how the pieces fit together.",
    blocks: [
      {
        type: "p",
        text: "Skopos coordinates LLM agents that plan and run security research against targets you are authorized to test. A single Go backend (with this UI embedded) stores everything in PostgreSQL; the model provider is configured separately.",
      },
      {
        type: "table",
        head: ["Role", "What it does"],
        rows: [
          [
            "Planner",
            "Reads the exploration graph and the asset scope. It is the only component that creates intents (directions to explore), and most wake-ups send none, which means there is nothing new to try.",
          ],
          [
            "Worker",
            "Claims one intent, runs real tools (Bash, HTTP, MCP, skills), writes facts, assets and findings back, then stops. The number of workers per task is a system setting (default 3).",
          ],
          [
            "Main agent",
            "The human-in-the-loop agent behind Chat. Use it to create tasks, ask about progress and steer work.",
          ],
          ["Goals", "Decomposes a goal and extracts scope from it when a task is created."],
        ],
      },
      {
        type: "p",
        text: "Two graphs hold state. The asset graph is global and shared: root domains, subdomains, IPs, services, apps and endpoints, with relationships computed by the program. The exploration graph is per task: goals, intents, facts, findings and hints linked by spawns / derived_from / yields / proves edges. Anchors connect the two, which is what powers asset coverage.",
      },
      {
        type: "callout",
        tone: "warn",
        title: "Authorization",
        text: "Use Skopos only against systems you own or are explicitly authorized to test. The usage restrictions in the Terms of Use and README apply and are unchanged from upstream.",
      },
    ],
  },
  {
    id: "quick-start",
    title: "Quick start",
    icon: Rocket,
    summary: "From a fresh install to your first reviewed finding.",
    blocks: [
      {
        type: "steps",
        items: [
          {
            title: "Configure a model profile",
            body: "Open LLM, create a profile (optionally from a template), enter your API key, run Test connection, save, and activate it. A new profile is not used until it is activated.",
            href: "/system/llm",
          },
          {
            title: "Check concurrency and defaults",
            body: "In System settings, review the number of worker agents per task. The value applies to tasks started afterwards.",
            href: "/system/settings",
          },
          {
            title: "Turn on approvals before you run anything",
            body: "In Interception rules, enable interception for Bash and the other execution tools so unmatched commands wait for a human. Review the Approval records page later to audit decisions.",
            href: "/system/intercept",
          },
          {
            title: "Create a task",
            body: "Use Tasks (or ask in Chat). Give it a clear goal and the targets in scope. Coverage proposals from agents appear in the task Overview for you to approve.",
            href: "/function/tasks",
          },
          {
            title: "Watch it work",
            body: "The task detail shows the exploration graph, sessions, and per-step activity. Dashboard summarises everything running.",
            href: "/dashboard",
          },
          {
            title: "Triage findings",
            body: "Open Findings, read the evidence and captured traffic, then retest or change workflow status as appropriate.",
            href: "/function/findings",
          },
        ],
      },
    ],
  },
  {
    id: "models",
    title: "Choosing models",
    icon: Brain,
    summary: "What to look for in a model, and how to assign models to roles.",
    blocks: [
      {
        type: "callout",
        tone: "info",
        title: "How to read this",
        text: "Recommendations follow from how Skopos works (long tool-using loops, a rarely-waking planner, many parallel workers). They are starting points rather than benchmark results: confirm availability, pricing and limits with your provider, and compare models on your own tasks using the LLM recordings page.",
      },
      {
        type: "p",
        text: "Skopos is tool-heavy and long-running, so a model is judged less by chat quality than by how reliably it keeps calling tools correctly across many turns.",
      },
      {
        type: "table",
        head: ["Requirement", "Why it matters", "How to check"],
        rows: [
          [
            "Reliable tool calling",
            "Workers and the planner act only through tools (graph tools, Bash, HTTP, MCP). Malformed arguments waste turns and trigger retries.",
            "Run a small task and review LLM recordings for tool-call errors and empty turns.",
          ],
          [
            "Long context window",
            "Scans and traffic produce large outputs. Skopos compresses history using the context window you declare, so the declared value must match the model.",
            "Set Context window (K) on the profile to the model's real limit; profiles with an insufficient window are skipped during rotation.",
          ],
          [
            "Controllable reasoning",
            "Planning benefits from deeper reasoning; routine execution does not. Effort levels are low, medium, high, xhigh and max where the provider supports them.",
            "Set thinking type and reasoning effort per profile; some APIs have no thinking field and enable reasoning through effort alone.",
          ],
          [
            "Sufficient output budget",
            "Too small an output limit can cut reasoning off before an answer is produced.",
            "Leave Output limit at 0 (provider default) unless you have a reason; use max_completion_tokens for OpenAI reasoning models.",
          ],
          [
            "Predictable cost and rate limits",
            "Several workers run in parallel and every step is a model call.",
            "Check provider rate limits against your worker count; use rotation or retry overrides for bursts.",
          ],
        ],
      },
      {
        type: "p",
        text: "Recommended assignment by role. Bind a profile to an Agent when you want a role to use a specific model; otherwise the active profile is used.",
      },
      {
        type: "table",
        head: ["Role", "Prioritise", "Suggested profile"],
        rows: [
          [
            "Planner",
            "Reasoning depth and judgment about what is still uncovered. Runs rarely, so cost per call matters least.",
            "claude-fable-5-1 or claude-opus-5-5 at effort high or above.",
          ],
          [
            "Worker",
            "Tool-calling reliability and throughput. Runs most often and in parallel.",
            "claude-sonnet-5-5 at medium effort, or glm-5.3 as a lower-cost alternative. Raise effort only if workers stall.",
          ],
          [
            "Main agent (Chat)",
            "Instruction following and clear explanations for you.",
            "The planner's model, or claude-sonnet-5-5.",
          ],
          [
            "Interception fallback judge",
            "Low latency and conservative decisions. It cannot authorize the unmatched-command fallback; humans decide that.",
            "claude-haiku-5-5 at low effort.",
          ],
        ],
      },
      {
        type: "p",
        text: "Recommended models and how to configure them.",
      },
      {
        type: "table",
        head: ["Model", "Format", "Best for", "Configuration notes"],
        rows: [
          [
            "claude-fable-5-1",
            "anthropic",
            "Planner, hardest tasks",
            "1M context, 128K output. Thinking is adaptive and cannot be turned off; default effort is high. Highest cost per call, which is acceptable because the planner wakes rarely. Template: LLM, then New.",
          ],
          [
            "claude-opus-5-5",
            "anthropic",
            "Planner, main agent",
            "1M context, 128K output. Thinking cannot be turned off; default effort is medium, so raise it to high for planning. A good default for both the planner and Chat. Template available.",
          ],
          [
            "claude-sonnet-5-5",
            "anthropic",
            "Workers, main agent",
            "1M context, 128K output. Adaptive thinking, default effort high. Best balance of tool-calling reliability, speed and cost for the many parallel workers. Template available.",
          ],
          [
            "claude-haiku-5-5",
            "anthropic",
            "Interception judge, light agents",
            "1M context, 128K output. Adaptive thinking, default effort medium; lower it for short, conservative decisions. Avoid it for planning. Template available.",
          ],
          [
            "glm-5.3",
            "openai",
            "Workers, cost-sensitive runs",
            "Built-in template (LLM, then New). Text-only, 1M context, 128K output, reasoning must stay on; effort low, high or max. Use the General API endpoint.",
          ],
          [
            "GPT-5 class",
            "openai-responses or openai",
            "Planner or workers",
            "OpenAI lists GPT-5 at 400K total context but accepts at most 272K input tokens, so declare 272 (not 400) as Context window. Set the output-limit field to max_completion_tokens when using the openai format. Check your account's model access first. No template.",
          ],
        ],
      },
      {
        type: "callout",
        tone: "info",
        title: "Fastest way to set a model up",
        text: "Open LLM, then New, and pick a template under Anthropic Claude or Z.ai GLM. It fills the format, model ID, base URL, context window (1000 for a 1M window), effort, and a profile name. Add your API key, run Test connection, save, then use Set active. If Context window shows 0, Skopos assumes 200K and compresses history early.",
      },
      {
        type: "table",
        head: ["Provider format", "Use when", "Notes"],
        rows: [
          [
            "anthropic",
            "Anthropic models, or gateways exposing the Anthropic Messages API.",
            "Output-limit field name is fixed by the format.",
          ],
          [
            "openai",
            "OpenAI Chat Completions and the many compatible gateways, including the Z.ai GLM-5.3 templates.",
            "max_tokens is the default; official OpenAI reasoning models (o series / GPT-5) need max_completion_tokens and reject max_tokens.",
          ],
          ["openai-responses", "OpenAI Responses API.", "Output-limit field name is fixed by the format."],
        ],
      },
      {
        type: "p",
        text: "Built-in template: Z.ai GLM-5.3 (LLM, then New, then Configuration template). General API is the recommended entry for API billing. The Coding Plan template is reference only: Z.ai limits that plan to officially supported tools and Skopos is not listed, so obtain Z.ai authorization first. GLM-5.3 is text-only with a 1M-token context, requires reasoning to stay enabled, and supports effort low, high and max (the template uses max).",
      },
      {
        type: "callout",
        tone: "warn",
        title: "Reliability",
        text: "Enable profile rotation if a single provider is a risk. The active profile is tried first, then others by priority. A failed profile cools down (60 s, then 5 min, then 30 min) and rejoins automatically. Agents and tasks bound to a specific profile do not rotate unless you enable fallback for them.",
      },
    ],
  },
  {
    id: "features",
    title: "Feature reference",
    icon: Compass,
    summary: "Every page in the sidebar and what it is for.",
    blocks: [
      {
        type: "features",
        items: [
          {
            title: "Dashboard",
            href: "/dashboard",
            body: "Overview of running work, token usage and recent activity.",
          },
          {
            title: "Chat",
            href: "/chat",
            body: "Talk to the main agent. Create tasks, ask for progress, and steer exploration in natural language.",
          },
          {
            title: "Tasks",
            href: "/function/tasks",
            body: "Create and manage tasks. The detail view has Overview (scope approval), sessions, the exploration graph and per-step activity.",
          },
          {
            title: "Findings",
            href: "/function/findings",
            body: "Recorded findings with evidence, grouping, workflow status and manual retest.",
          },
          {
            title: "Traffic",
            href: "/function/traffic",
            body: "Requests recorded by the capture proxy for Bash and HTTP tool traffic, linkable to findings as evidence.",
          },
          {
            title: "Tool executions",
            href: "/function/commands",
            body: "Searchable history of tool runs with parameters, task IDs and statistics.",
          },
          {
            title: "LLM recordings",
            href: "/function/llm-records",
            body: "Raw model requests and responses. The first place to look when a model misbehaves.",
          },
          {
            title: "Assets",
            href: "/function/assets",
            body: "The shared asset graph and coverage view of what has been tested.",
          },
          {
            title: "Asset sync",
            href: "/function/sync",
            body: "Import assets from ScopeSentry by project or task and merge them into the company scope.",
          },
          {
            title: "Workspace",
            href: "/function/workspace",
            body: "File manager for the server-side workspace: browse, view and edit text files, upload, create folders, download and delete.",
          },
          {
            title: "LLM",
            href: "/system/llm",
            body: "Model profiles, templates, connection test, model loading, rotation and retry policy.",
          },
          { title: "Agents", href: "/system/agents", body: "Edit agent prompts and bind agents to model profiles." },
          { title: "MCP", href: "/system/mcp", body: "Register MCP servers to extend what agents can call." },
          {
            title: "Skills",
            href: "/system/skills",
            body: "Bundled and uploaded skill guidance that agents load on demand.",
          },
          {
            title: "Tools",
            href: "/system/tools",
            body: "Declare tools available in Bash, such as nmap, sqlmap or ffuf, or define custom tools.",
          },
          { title: "Notifications", href: "/system/notify", body: "Send task events to external channels." },
          {
            title: "Interception rules",
            href: "/system/intercept",
            body: "Approval gate for tool execution: allow and deny rules, with human approval for unmatched commands.",
          },
          {
            title: "Asset blocking",
            href: "/system/intercept/assets",
            body: "Keep specific assets out of reach of agents.",
          },
          {
            title: "Approval records",
            href: "/system/intercept/approvals",
            body: "Audit trail of approval decisions. The sidebar badge counts pending requests.",
          },
          { title: "Logs", href: "/system/logs", body: "Live backend log stream with level filter and search." },
          {
            title: "System settings",
            href: "/system/settings",
            body: "Worker concurrency, Python interpreter and automatic traffic binding.",
          },
        ],
      },
    ],
  },
  {
    id: "manual",
    title: "Operating manual",
    icon: Target,
    summary: "Step-by-step procedures for day-to-day work.",
    blocks: [
      {
        type: "p",
        text: "Writing a good task goal",
      },
      {
        type: "list",
        items: [
          "State one outcome you can verify, for example confirming whether a specific class of weakness exists on named hosts.",
          "List targets explicitly. Anything outside the approved scope is a coverage proposal, not permission.",
          "Say what is off limits (destructive actions, third-party services, rate limits).",
        ],
      },
      { type: "p", text: "Running a task" },
      {
        type: "steps",
        items: [
          {
            title: "Create it",
            body: "Tasks, then New, or ask the main agent in Chat. The Goals component extracts scope from your wording.",
          },
          {
            title: "Approve scope",
            body: "Open the task Overview and review agent-proposed and automatic entries. Approved scope from directly linked source tasks also grants reads.",
          },
          {
            title: "Answer approvals",
            body: "Pending approvals show a badge on Approval records. Cancelling or letting the five-minute timeout expire denies the command.",
          },
          {
            title: "Intervene when needed",
            body: "Pause or resume from the task, or message the main agent to redirect. Paused tasks drain in-flight work before stopping.",
          },
          {
            title: "Review results",
            body: "Use the exploration graph to see which facts led to each finding, and the Assets coverage view to see what was never tested.",
          },
        ],
      },
      { type: "p", text: "Handling a finding" },
      {
        type: "steps",
        items: [
          { title: "Read the evidence", body: "Open the finding and inspect the linked traffic." },
          {
            title: "Retest manually",
            body: "Use the retest action to re-run validation and record the outcome on the finding.",
          },
          {
            title: "Move it through the workflow",
            body: "Update the status as you triage so the list reflects reality, then export a report when done.",
          },
        ],
      },
      { type: "p", text: "Writing approval rules" },
      {
        type: "list",
        items: [
          "Matching allow and deny rules take priority over the fallback.",
          "With a restrictive rule enabled, every unmatched Bash command needs a human, including scripts, native programs and empty interactive input.",
          "Commands whose executable comes from a variable or command substitution, generated shell programs, eval and source of generated content all need approval.",
          "For curl and wget, request methods that depend on runtime arguments or configuration files need approval. Use curl -q / --disable or wget --no-config as the first option to avoid default-config approval.",
          "Allow rules should be narrow. Prefer one rule per repeated workflow over broad patterns.",
        ],
      },
      {
        type: "callout",
        tone: "warn",
        title: "What approvals are not",
        text: "The approval gate is not an OS sandbox. It does not inspect every generated program or external script, and it does not enforce runtime network limits. Run Skopos in an isolated environment.",
      },
    ],
  },
  {
    id: "safety",
    title: "Safety and isolation",
    icon: ShieldAlert,
    summary: "Controls that keep a run inside its boundaries.",
    blocks: [
      {
        type: "list",
        items: [
          "Scope: coverage proposals need operator approval; asset blocking removes assets from agent reach.",
          "Interception: human approval for unmatched execution, with the model judge unable to authorize the strict fallback.",
          "Traffic: Bash and HTTP tool traffic goes through the recording proxy so you can audit what was sent.",
          "Isolation: see docs/EXECUTION-ISOLATION.md in the repository for the execution isolation model.",
          "Credentials: API keys are shown only as hints after saving. Never paste secrets into task goals or chat.",
        ],
      },
    ],
  },
  {
    id: "system",
    title: "Configuration",
    icon: Wrench,
    summary: "Deployment and runtime configuration.",
    blocks: [
      {
        type: "table",
        head: ["Setting", "Where", "Notes"],
        rows: [
          [
            "Database",
            "config.json or SKOPOS_PG_DSN",
            "PostgreSQL is required. The schema is created idempotently on every start.",
          ],
          [
            "Model provider",
            "LLM page, or ANTHROPIC_API_KEY / OPENAI_API_KEY",
            "Optional: SKOPOS_LLM_PROVIDER, SKOPOS_LLM_MODEL, SKOPOS_LLM_BASE_URL, SKOPOS_LLM_PROXY.",
          ],
          ["Worker concurrency", "System settings", "Default 3 per task; applies to tasks started afterwards."],
          [
            "Listen addresses",
            "-addr and -proxy flags",
            "-addr serves frontend and API, -proxy is the traffic-capture proxy.",
          ],
          ["Language", "English only", "The interface and built-in agent guidance are English."],
        ],
      },
      { type: "code", text: "./start.sh -addr 127.0.0.1:8787 -proxy 127.0.0.1:8788" },
    ],
  },
  {
    id: "troubleshooting",
    title: "Troubleshooting",
    icon: LifeBuoy,
    summary: "Common problems and where to look.",
    blocks: [
      {
        type: "faq",
        items: [
          {
            q: "A task starts but nothing happens.",
            a: "Check that a profile is activated (saving is not enough) and that Test connection succeeds. Then open LLM recordings to see the actual request and error.",
          },
          {
            q: "The model keeps returning empty turns or malformed tool calls.",
            a: "Try a model with stronger tool calling, lower the reasoning effort, or raise the output limit. Compare recordings between models before changing prompts.",
          },
          {
            q: "Requests fail with unsupported_parameter.",
            a: "OpenAI reasoning models reject max_tokens. Set the output-limit field to max_completion_tokens on the profile.",
          },
          {
            q: "Context limit exceeded messages appear.",
            a: "Skopos reduces history and retries. If it recurs, make sure Context window (K) matches the model, or switch to a longer-context model.",
          },
          {
            q: "Commands hang with no output.",
            a: "Check Approval records for a pending request. Unmatched commands wait for a human and are denied after five minutes.",
          },
          {
            q: "A provider keeps failing.",
            a: "Failed profiles enter cooldown and are skipped, then rejoin. Enable rotation and add a second profile for resilience.",
          },
          {
            q: "I cannot find an asset in coverage.",
            a: "Scope is approved per task. Check the task Overview for pending proposals, and Asset blocking for exclusions.",
          },
        ],
      },
    ],
  },
  {
    id: "glossary",
    title: "Glossary",
    icon: Cpu,
    summary: "Terms used across the interface.",
    blocks: [
      {
        type: "table",
        head: ["Term", "Meaning"],
        rows: [
          ["Intent", "One direction for a worker to explore. Created only by the planner."],
          ["Frontier", "The queue of intents waiting to be claimed."],
          ["Fact", "An observation a worker wrote back to the exploration graph."],
          ["Anchor", "A link between an exploration node and an asset."],
          ["Coverage", "Which in-scope assets have been tested."],
          ["Profile", "A saved model connection: format, base URL, model, key, reasoning and limits."],
          ["Skill", "Reusable guidance an agent can load for a specific kind of work."],
          ["MCP", "Model Context Protocol servers that add tools for agents."],
          ["Interception", "The approval gate in front of tool execution."],
        ],
      },
    ],
  },
];
