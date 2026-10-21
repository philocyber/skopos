export const en = {
  template: "Configuration template (new profile only)",
  chooseTemplate: "Choose a template…",
  general: "Z.ai GLM-5.3 · General API (recommended)",
  coding: "Z.ai GLM-5.3 · Coding Plan (reference)",
  introduction:
    "Optional: prefill connection fields, then enter a profile name and your API key. Your key, proxy, and name are kept. Selection does not save or activate a profile. Save when ready, then activate separately. The model remains editable.",
  generalHelp:
    "Uses the General API endpoint and API billing, separate from Coding Plan subscription quota. Confirm model access and billing for your Z.ai account.",
  codingHelp:
    "Reference only: Coding Plan uses a separate endpoint and subscription quota. Z.ai restricts it to officially supported tools; Skopos is not listed. Obtain Z.ai authorization before using this endpoint with Skopos.",
  modelHelp:
    "GLM-5.3 is text-only with a 1M-token context. Reasoning must stay enabled; supported effort levels are low, high, and max. This template selects max and leaves the output limit at 0 (provider default). Review these settings if you change the model.",
  officialModel: "GLM-5.3 model guide",
  officialPolicy: "Coding Plan usage policy",
  claudeGroup: "Anthropic Claude",
  zaiGroup: "Z.ai GLM",
  claudeFable: "Claude Fable 5.1 · planner for the hardest tasks",
  claudeOpus: "Claude Opus 5.5 · planner and main agent (recommended)",
  claudeSonnet: "Claude Sonnet 5.5 · workers",
  claudeHaiku: "Claude Haiku 5.5 · interception judge and light agents",
  claudeHelp:
    "Uses the official Anthropic endpoint, so Base URL stays empty. The model has a 1M-token context and 128K output, so Context window is set to 1000. Thinking is adaptive: the thinking field is left unset and effort is the model's documented default. Raise effort for planning. Add your Anthropic API key.",
  claudeOfficial: "Model page",
  cwUnset:
    "Context window is not set, so Skopos assumes 200K and compresses history early. Declare the model's real window, for example 1000 for 1M tokens.",
};
