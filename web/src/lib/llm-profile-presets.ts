// Presets change form fields only; they never contact a provider or save anything.
//
// Z.ai documentation verified 2026-10-03:
//   https://docs.z.ai/guides/llm/glm-5.3
//   https://zcode.z.ai/en/docs/configuration
// Anthropic model pages verified 2026-10-21 (model ID, 1M context, 128K max output, default effort):
//   https://platform.claude.com/docs/en/models/fable-5-1/overview
//   https://platform.claude.com/docs/en/models/opus-5-5/overview
//   https://platform.claude.com/docs/en/models/sonnet-5-5/overview
//   https://platform.claude.com/docs/en/models/haiku-5-5/overview
//
// OpenAI model pages (gpt-6.1-sol, gpt-6-luna) as indexed on 2026-10-21: 1,050,000 context, 128,000 max
// output, so max input is 922,000 and that is the window declared here. GPT-6.1 Sol does not support function
// calling on Chat Completions, so both OpenAI presets use the Responses format.
//   https://developers.openai.com/api/docs/models/gpt-6.1-sol
//   https://developers.openai.com/api/docs/models/gpt-6-luna
//
// Claude presets leave thinking.type empty on purpose: these models use adaptive thinking, Opus 5.5 and
// Fable 5.1 cannot turn it off, and effort is the control. Effort is each model's documented default.
export const LLM_PROFILE_PRESET_IDS = [
  "zai-general",
  "zai-coding-reference",
  "claude-fable-5-1",
  "claude-opus-5-5",
  "claude-sonnet-5-5",
  "claude-haiku-5-5",
  "openai-gpt-6-1-sol",
  "openai-gpt-6-luna",
] as const;
export type LLMProfilePresetId = (typeof LLM_PROFILE_PRESET_IDS)[number];

// Suggested profile name, used only when the name field is still empty.
export const LLM_PROFILE_PRESET_NAMES: Record<LLMProfilePresetId, string> = {
  "zai-general": "GLM-5.3",
  "zai-coding-reference": "GLM-5.3 Coding Plan",
  "claude-fable-5-1": "Claude Fable 5.1",
  "claude-opus-5-5": "Claude Opus 5.5",
  "claude-sonnet-5-5": "Claude Sonnet 5.5",
  "claude-haiku-5-5": "Claude Haiku 5.5",
  "openai-gpt-6-1-sol": "GPT-6.1 Sol",
  "openai-gpt-6-luna": "GPT-6 Luna",
};

export interface LLMProfilePresetFields {
  format: "openai" | "anthropic" | "openai-responses";
  model: string;
  base_url: string;
  thinking_type: "enabled" | "";
  reasoning_effort: string;
  context_window_k: number;
  max_tokens: number;
  max_tokens_field: "";
}

const ZAI_GENERAL = "https://api.z.ai/api/paas/v4";
const ZAI_CODING = "https://api.z.ai/api/coding/paas/v4";

const PRESETS: Record<LLMProfilePresetId, LLMProfilePresetFields> = {
  "zai-general": zai(ZAI_GENERAL),
  "zai-coding-reference": zai(ZAI_CODING),
  "claude-fable-5-1": claude("claude-fable-5-1", "high"),
  "claude-opus-5-5": claude("claude-opus-5-5", "medium"),
  "claude-sonnet-5-5": claude("claude-sonnet-5-5", "high"),
  "claude-haiku-5-5": claude("claude-haiku-5-5", "medium"),
  "openai-gpt-6-1-sol": openaiResponses("gpt-6.1-sol", "medium"),
  "openai-gpt-6-luna": openaiResponses("gpt-6-luna", "medium"),
};

function zai(baseUrl: string): LLMProfilePresetFields {
  return {
    format: "openai",
    model: "glm-5.3",
    base_url: baseUrl,
    thinking_type: "enabled",
    reasoning_effort: "max",
    context_window_k: 1000,
    // Preserve the existing default: omit the output limit and use provider defaults.
    max_tokens: 0,
    max_tokens_field: "",
  };
}

function claude(model: string, effort: string): LLMProfilePresetFields {
  return {
    format: "anthropic",
    model,
    // Empty uses the official Anthropic endpoint.
    base_url: "",
    thinking_type: "",
    reasoning_effort: effort,
    context_window_k: 1000,
    max_tokens: 0,
    max_tokens_field: "",
  };
}

function openaiResponses(model: string, effort: string): LLMProfilePresetFields {
  return {
    format: "openai-responses",
    model,
    // Empty uses the official OpenAI endpoint.
    base_url: "",
    thinking_type: "",
    reasoning_effort: effort,
    // 1,050,000 total minus 128,000 max output.
    context_window_k: 922,
    max_tokens: 0,
    max_tokens_field: "",
  };
}

export function getLLMProfilePreset(id: string): Readonly<LLMProfilePresetFields> {
  if (!(LLM_PROFILE_PRESET_IDS as readonly string[]).includes(id)) {
    throw new RangeError(`Unknown LLM profile preset: ${id}`);
  }
  return { ...PRESETS[id as LLMProfilePresetId] };
}

// Unknown fields, including credentials and profile names, survive unchanged.
// The original draft is not mutated, and no request or persistence occurs here.
export function applyLLMProfilePreset<T extends object>(draft: T, id: string): T & LLMProfilePresetFields {
  return { ...draft, ...getLLMProfilePreset(id) };
}
