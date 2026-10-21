package server

import (
	"encoding/json"
	"github.com/Philocyber/skopos/agent"
	"github.com/Philocyber/skopos/db"
	"github.com/Philocyber/skopos/locale"
	"github.com/Philocyber/skopos/traffic"
	actool "github.com/Autumn-27/norma/tool"
)

// knownToolLanguages constructs metadata-only tool shells. No handlers, database
// operations, provider calls, or filesystem actions are executed by constructors.
func (s *Server) knownToolLanguages() map[string][2]actool.CoreTool {
	pairs := map[string][2]actool.CoreTool{}
	for index, lang := range []locale.Lang{locale.En, locale.Ko} {
		tools := agent.NewToolSet(nil, "", lang).AllDomainTools()
		tools = append(tools, (&traffic.Traffic{}).Tools(lang)...)
		tools = append(tools, s.orchestrationTools(lang)...)
		tools = append(tools, s.platformTools(lang)...)
		tools = append(tools, s.findingRetestTools(lang)...)
		for _, tool := range tools {
			pair := pairs[tool.Name()]
			pair[index] = tool
			pairs[tool.Name()] = pair
		}
	}
	return pairs
}

func translatedToolMetadata(tool actool.CoreTool, pair [2]actool.CoreTool, lang locale.Lang) actool.CoreTool {
	if pair[0] == nil || pair[1] == nil {
		return tool
	}
	if lang == locale.Ko {
		return agent.TranslateToolMetadata(tool, pair[0], pair[1])
	}
	return agent.TranslateToolMetadata(tool, pair[1], pair[0])
}

// translatedStoredTool returns a display/save copy of a system tool. Exact stock
// descriptions can change language; custom fields and all executable data remain.
func translatedStoredTool(row *db.Tool, pairs map[string][2]actool.CoreTool, lang locale.Lang) *db.Tool {
	copy := *row
	pair, known := pairs[row.Key]
	if !row.System || !known || pair[0] == nil || pair[1] == nil {
		return &copy
	}
	var schema map[string]any
	if json.Unmarshal(row.Schema, &schema) != nil {
		return &copy
	}
	current := agent.DecorateTool(pair[0], row.Description, schema)
	rendered := translatedToolMetadata(current, pair, lang)
	copy.Description = rendered.Description()
	if raw, err := json.Marshal(rendered.InputSchema()); err == nil {
		copy.Schema = raw
	}
	return &copy
}
