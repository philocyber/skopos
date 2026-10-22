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
  openaiGroup: "OpenAI",
  openaiSol: "GPT-6.1 Sol · planner or workers",
  openaiLuna: "GPT-6 Luna · workers, light agents and the interception judge",
  openaiHelp:
    "Uses the Responses API, which Skopos needs here: GPT-6.1 Sol does not support function calling on Chat Completions. Base URL stays empty for the official endpoint. The window is 1.05M tokens with 128K output, so Context window is set to 922 (the input share). Effort is the documented default, medium. Add your OpenAI API key.",
  openaiOfficial: "Model page",
  claudeFable: "Claude Fable 5.1 · planner for the hardest tasks",
  claudeOpus: "Claude Opus 5.5 · planner and main agent (recommended)",
  claudeSonnet: "Claude Sonnet 5.5 · workers",
  claudeHaiku: "Claude Haiku 5.5 · interception judge and light agents",
  claudeHelp:
    "Uses the official Anthropic endpoint, so Base URL stays empty. The model has a 1M-token context and 128K output, so Context window is set to 1000. Thinking is adaptive: the thinking field is left unset and effort is the model's documented default. Raise effort for planning. Add your Anthropic API key.",
  claudeOfficial: "Model page",
  vendor: "Provider",
  chooseVendor: "Choose a provider…",
  vendorAnthropic: "Anthropic (Claude)",
  vendorOpenai: "OpenAI",
  vendorZai: "Z.ai (GLM)",
  vendorVercel: "Vercel AI Gateway (one key, many models)",
  vendorBedrock: "Amazon Bedrock (one key, many models)",
  region: "AWS Region",
  regionHelp:
    "Must match the Region the Bedrock API key was created in. Other Regions: edit the Base URL below.",
  getKey: "Get a {p0} API key",
  keyBedrock: "Bedrock API key",
  keyVercel: "AI Gateway API key",
  gatewayGroup: "Gateways (one key, many models)",
  vercelChat: "Vercel AI Gateway · any model, OpenAI Chat Completions",
  vercelClaude: "Vercel AI Gateway · Claude, Anthropic Messages",
  bedrockClaude: "Amazon Bedrock · Claude, Anthropic Messages",
  bedrockOpen: "Amazon Bedrock · DeepSeek, Qwen, GLM and other models",
  vercelHelp:
    "Use an AI Gateway API key from the Vercel dashboard (AI Gateway, API keys). Model IDs are provider/model, for example anthropic/claude-sonnet-5; use Load models for the exact list of providers and models. For OpenAI models that need function calling, switch Format to OpenAI (Responses API) and keep the same Base URL. Thinking, effort and the 200K window are left conservative: set them for the model you pick.",
  bedrockClaudeHelp:
    "Use an Amazon Bedrock API key (Bedrock console, API keys). The Region in the Base URL must match the Region the key was created in; change us-east-1 if needed. Model IDs have no slash, for example us.anthropic.claude-sonnet-5 (cross-Region inference profile). Load models is unavailable on bedrock-runtime: type the ID from the model's card. Short-term keys expire within 12 hours.",
  bedrockOpenHelp:
    "Use an Amazon Bedrock API key (Bedrock console, API keys). This preset uses the bedrock-mantle endpoint so Load models works and lists DeepSeek, Qwen, GLM, Kimi, Mistral and gpt-oss models. The Region in the Base URL must match the key's Region. If a model answers 'isn't supported on this route', change /v1 to /openai/v1 in the Base URL.",
  vercelDocs: "AI Gateway docs",
  bedrockDocs: "Bedrock API keys",
  warnBedrockRegion:
    "The Base URL still has a region placeholder. Replace it with your AWS Region, for example us-east-1, matching the Region your Bedrock API key was created in.",
  warnVercelAnthropicPath:
    "For Vercel AI Gateway with the Anthropic format, the Base URL must be https://ai-gateway.vercel.sh with no path (a trailing /v1 would become /v1/v1/messages).",
  warnVercelOpenAIPath: "For Vercel AI Gateway with an OpenAI format, the Base URL must be https://ai-gateway.vercel.sh/v1.",
  warnVercelModel:
    "Vercel AI Gateway model IDs use provider/model, for example anthropic/claude-sonnet-5. Use Load models for the exact IDs.",
  warnBedrockAnthropicPath:
    "For Amazon Bedrock with the Anthropic format, the Base URL path must be exactly /anthropic.",
  warnBedrockOpenAIPath:
    "For Amazon Bedrock bedrock-runtime with an OpenAI format, the Base URL path must be exactly /openai/v1.",
  warnMantleOpenAIPath:
    "For Amazon Bedrock bedrock-mantle with an OpenAI format, the Base URL path must be /v1 (or /openai/v1 for the few models that use it).",
  warnBedrockModel:
    "Amazon Bedrock model IDs have no slash, for example us.anthropic.claude-sonnet-5. A provider/model ID belongs to Vercel AI Gateway.",
  cwUnset:
    "Context window is not set, so Skopos assumes 200K and compresses history early. Declare the model's real window, for example 1000 for 1M tokens.",
};
