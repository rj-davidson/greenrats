package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type Client struct {
	client openai.Client
	model  string
	logger *slog.Logger
}

func New(apiKey, model string, logger *slog.Logger) *Client {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &Client{client: client, model: model, logger: logger}
}

func (c *Client) SearchTournamentEarnings(ctx context.Context, tournamentName string, year int, golfers []GolferInput) ([]EarningsResult, error) {
	c.logger.Info("searching tournament earnings", "tournament", tournamentName, "year", year, "golfers", len(golfers))

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

	c.logger.Debug("tournament earnings search complete", "results", len(result.Results))
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
						"earnings": map[string]any{
							"type":        "integer",
							"description": "Prize money earned in USD",
						},
					},
					"required":             []string{"golfer_id", "earnings"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"results"},
		"additionalProperties": false,
	}
}

func (c *Client) SearchTournamentLeaderboard(ctx context.Context, tournamentName string, year int) (*LeaderboardResponse, error) {
	c.logger.Info("searching tournament leaderboard", "tournament", tournamentName, "year", year)

	prompt := leaderboardSearchPrompt(tournamentName, year)

	resp, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
		Model: c.model,
		Tools: []responses.ToolUnionParam{
			{OfWebSearch: &responses.WebSearchToolParam{Type: "web_search_preview"}},
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "leaderboard_response",
					Strict: openai.Bool(true),
					Type:   "json_schema",
					Schema: leaderboardResponseSchema(),
				},
			},
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search tournament leaderboard: %w", err)
	}

	var result LeaderboardResponse
	if err := json.Unmarshal([]byte(resp.OutputText()), &result); err != nil {
		return nil, fmt.Errorf("failed to parse leaderboard response: %w", err)
	}

	c.logger.Debug("tournament leaderboard search complete", "entries", len(result.Entries))
	return &result, nil
}

func leaderboardResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"tournament_name": map[string]any{
				"type":        "string",
				"description": "The name of the tournament",
			},
			"entries": map[string]any{
				"type":        "array",
				"description": "List of all players who earned prize money",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type":        "string",
							"description": "The player's full name",
						},
						"earnings": map[string]any{
							"type":        "integer",
							"description": "Prize money earned in USD",
						},
					},
					"required":             []string{"name", "earnings"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"tournament_name", "entries"},
		"additionalProperties": false,
	}
}

func (c *Client) MatchPlayersToLeaderboard(ctx context.Context, leaderboard *LeaderboardResponse, golfers []GolferInput) ([]EarningsResult, error) {
	c.logger.Info("matching players to leaderboard", "golfers", len(golfers), "leaderboard_entries", len(leaderboard.Entries))

	leaderboardJSON, err := json.Marshal(leaderboard.Entries)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal leaderboard: %w", err)
	}

	golfersJSON, err := json.Marshal(golfers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal golfers: %w", err)
	}

	prompt := matchPlayersPrompt(string(leaderboardJSON), string(golfersJSON))

	resp, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
		Model: c.model,
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
		return nil, fmt.Errorf("failed to match players to leaderboard: %w", err)
	}

	var result EarningsResponse
	if err := json.Unmarshal([]byte(resp.OutputText()), &result); err != nil {
		return nil, fmt.Errorf("failed to parse match response: %w", err)
	}

	c.logger.Debug("player matching complete", "results", len(result.Results))
	for _, r := range result.Results {
		c.logger.Debug("matched golfer", "golfer_id", r.GolferID, "earnings", r.Earnings)
	}
	return result.Results, nil
}

func (c *Client) ParseLeaderboardContent(ctx context.Context, content, tournamentName string) (*LeaderboardResponse, error) {
	c.logger.Info("parsing leaderboard content", "tournament", tournamentName)

	prompt := parseLeaderboardContentPrompt(content, tournamentName)

	resp, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
		Model: c.model,
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "leaderboard_response",
					Strict: openai.Bool(true),
					Type:   "json_schema",
					Schema: leaderboardResponseSchema(),
				},
			},
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse leaderboard content: %w", err)
	}

	var result LeaderboardResponse
	if err := json.Unmarshal([]byte(resp.OutputText()), &result); err != nil {
		return nil, fmt.Errorf("failed to parse leaderboard response: %w", err)
	}

	c.logger.Debug("parsed leaderboard", "entries", len(result.Entries))
	return &result, nil
}

func (c *Client) ResolveDuplicateEarnings(ctx context.Context, tournamentName string, inputs []DuplicateEarningsInput) (map[string]int, error) {
	c.logger.Info("resolving duplicate earnings", "tournament", tournamentName, "duplicates", len(inputs))

	if len(inputs) == 0 {
		return make(map[string]int), nil
	}

	duplicatesJSON, err := json.Marshal(inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal duplicates: %w", err)
	}

	prompt := resolveDuplicateEarningsPrompt(tournamentName, string(duplicatesJSON))

	resp, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
		Model: c.model,
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "duplicate_earnings_response",
					Strict: openai.Bool(true),
					Type:   "json_schema",
					Schema: duplicateEarningsResponseSchema(),
				},
			},
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve duplicate earnings: %w", err)
	}

	var result DuplicateEarningsResponse
	if err := json.Unmarshal([]byte(resp.OutputText()), &result); err != nil {
		return nil, fmt.Errorf("failed to parse duplicate earnings response: %w", err)
	}

	resolved := make(map[string]int)
	for _, r := range result.Results {
		resolved[r.GolferName] = r.Earnings
	}

	c.logger.Debug("resolved duplicate earnings", "resolved", len(resolved))
	return resolved, nil
}

func duplicateEarningsResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"results": map[string]any{
				"type":        "array",
				"description": "Resolved earnings for each golfer",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"golfer_name": map[string]any{
							"type":        "string",
							"description": "The golfer name from the input",
						},
						"earnings": map[string]any{
							"type":        "integer",
							"description": "The correct prize money earned in USD",
						},
					},
					"required":             []string{"golfer_name", "earnings"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"results"},
		"additionalProperties": false,
	}
}
