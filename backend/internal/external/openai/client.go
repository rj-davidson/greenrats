package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type Client struct {
	client openai.Client
	model  string
}

func New(apiKey, model string) *Client {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &Client{client: client, model: model}
}

func (c *Client) SearchTournamentEarnings(ctx context.Context, tournamentName string, year int, golfers []GolferInput) ([]EarningsResult, error) {
	golfersJSON, err := json.Marshal(golfers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal golfers: %w", err)
	}

	prompt := earningsAgentPrompt(tournamentName, year, string(golfersJSON))

	resp, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
		Model: c.model,
		Tools: []responses.ToolUnionParam{
			{OfWebSearch: &responses.WebSearchToolParam{Type: "web_search_preview"}},
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "earnings_response",
					Strict: openai.Bool(true),
					Type:   "json_schema",
					Schema: earningsResponseSchema(),
				},
			},
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search tournament earnings: %w", err)
	}

	var result EarningsResponse
	if err := json.Unmarshal([]byte(resp.OutputText()), &result); err != nil {
		return nil, fmt.Errorf("failed to parse earnings response: %w", err)
	}

	return result.Results, nil
}

func earningsResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"results": map[string]any{
				"type":        "array",
				"description": "List of tournament results matched to golfer_id values from the input",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"golfer_id": map[string]any{
							"type":        "string",
							"description": "The golfer_id from the input list (database identifier, not placement)",
						},
						"position": map[string]any{
							"type":        "integer",
							"description": "Final finishing position in the tournament (1 = winner, 2 = second, etc)",
						},
						"earnings": map[string]any{
							"type":        "integer",
							"description": "Prize money earned in USD",
						},
					},
					"required":             []string{"golfer_id", "position", "earnings"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"results"},
		"additionalProperties": false,
	}
}
