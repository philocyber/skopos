import assert from "node:assert/strict";
import test from "node:test";
import { applyLLMProfilePreset, getLLMProfilePreset } from "./llm-profile-presets.ts";

test("GLM templates distinguish billing endpoints and use documented model settings", () => {
  const general = getLLMProfilePreset("zai-general");
  const coding = getLLMProfilePreset("zai-coding-reference");
  assert.equal(general.base_url, "https://api.z.ai/api/paas/v4");
  assert.equal(coding.base_url, "https://api.z.ai/api/coding/paas/v4");
  for (const preset of [general, coding]) {
    assert.equal(preset.format, "openai");
    assert.equal(preset.model, "glm-5.3");
    assert.equal(preset.thinking_type, "enabled");
    assert.equal(preset.reasoning_effort, "max");
    assert.equal(preset.context_window_k, 1000);
    assert.equal(preset.max_tokens, 0);
    assert.equal(preset.max_tokens_field, "");
  }
});

test("applying a template preserves entered secrets and unrelated draft preferences", () => {
  const draft = Object.freeze({ name: "My profile", api_key: "user-entered-key", proxy: "http://user:pass@proxy:8080", username: "operator", session_header_key: "x-session", streaming: false, rate_per_minute: 7, model: "custom-model", base_url: "https://previous.example" });
  const next = applyLLMProfilePreset(draft, "zai-general");
  for (const key of ["name", "api_key", "proxy", "username", "session_header_key", "streaming", "rate_per_minute"]) assert.equal(next[key], draft[key]);
  assert.equal(draft.model, "custom-model");
  assert.notEqual(next, draft);
  next.model = "user-selected-model";
  assert.equal(next.model, "user-selected-model", "the preset does not lock model editing");
});

test("unsupported template IDs fail explicitly without changing the draft", () => {
  const draft = { api_key: "keep", model: "keep-model" };
  for (const id of ["", "glm-unknown", "__proto__", "zai-general-extra"]) {
    assert.throws(() => applyLLMProfilePreset(draft, id), RangeError);
    assert.deepEqual(draft, { api_key: "keep", model: "keep-model" });
  }
});
