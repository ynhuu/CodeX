package types

type OpenAIResponse struct {
	Model           string           `json:"model"`
	Input           []map[string]any `json:"input"`
	MaxOutputTokens uint64           `json:"max_output_tokens"`
	Text            map[string]any   `json:"text"`
	Store           bool             `json:"store"`
	Include         []string         `json:"include"`
	Reasoning       map[string]any   `json:"reasoning"`
	Tools           []map[string]any `json:"tools"`
	ToolChoice      string           `json:"tool_choice"`
	Stream          bool             `json:"stream"`
}

func (d *OpenAIResponse) ToCodex(promptCacheKey string) CodexResponse {
	var instructions string
	newInput := make([]map[string]any, 0, len(d.Input))
	for _, v := range d.Input {
		if v["role"] == "developer" {
			if s, ok := v["content"].(string); ok {
				instructions = s
			}
			continue
		}
		newInput = append(newInput, v)
	}

	return CodexResponse{
		Model:          d.Model,
		Input:          newInput,
		Instructions:   instructions,
		Text:           d.Text,
		Store:          d.Store,
		Include:        d.Include,
		PromptCacheKey: promptCacheKey,
		Reasoning:      d.Reasoning,
		Tools:          d.Tools,
		ToolChoice:     d.ToolChoice,
		Stream:         d.Stream,
	}
}
