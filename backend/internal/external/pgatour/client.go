package pgatour

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	DefaultBaseURL  = "https://orchestrator.pgatour.com/graphql"
	DefaultTourCode = "R"
)

type Client struct {
	client  *resty.Client
	apiKey  string
	baseURL string
}

func New(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json")

	if apiKey != "" {
		client.SetHeader("x-api-key", apiKey)
	}

	return &Client{
		client:  client,
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

type Field struct {
	TournamentName string          `json:"tournamentName"`
	ID             string          `json:"id"`
	LastUpdated    json.RawMessage `json:"lastUpdated"`
	Message        string          `json:"message"`
	Players        []FieldPlayer   `json:"players"`
	Alternates     []FieldPlayer   `json:"alternates"`
}

type FieldPlayer struct {
	ID          string      `json:"id"`
	DisplayName string      `json:"displayName"`
	FirstName   string      `json:"firstName"`
	LastName    string      `json:"lastName"`
	ShortName   string      `json:"shortName"`
	Country     string      `json:"country"`
	CountryFlag string      `json:"countryFlag"`
	Amateur     bool        `json:"amateur"`
	Withdrawn   bool        `json:"withdrawn"`
	Status      string      `json:"status"`
	OWGR        IntOrString `json:"owgr"`
}

type IntOrString struct {
	Int   int
	Valid bool
}

func (v *IntOrString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	var asInt int
	if err := json.Unmarshal(data, &asInt); err == nil {
		v.Int = asInt
		v.Valid = true
		return nil
	}

	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		if asString == "" {
			return nil
		}
		var parsed int
		if _, err := fmt.Sscanf(asString, "%d", &parsed); err == nil {
			v.Int = parsed
			v.Valid = true
		}
		return nil
	}

	return fmt.Errorf("invalid int/string value: %s", string(data))
}

type ScheduleTournament struct {
	ID               string          `json:"id"`
	TournamentName   string          `json:"tournamentName"`
	StartDate        json.RawMessage `json:"startDate"`
	TournamentStatus string          `json:"tournamentStatus"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type graphQLRequest struct {
	OperationName string                 `json:"operationName,omitempty"`
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

func (c *Client) GetField(ctx context.Context, fieldID string, includeWithdrawn bool, changesOnly bool) (*Field, error) {
	var response struct {
		Data struct {
			Field Field `json:"field"`
		} `json:"data"`
		Errors []graphQLError `json:"errors"`
	}

	req := graphQLRequest{
		OperationName: "Field",
		Query: `query Field($fieldId: ID!, $includeWithdrawn: Boolean, $changesOnly: Boolean) {
  field(id: $fieldId, includeWithdrawn: $includeWithdrawn, changesOnly: $changesOnly) {
    tournamentName
    id
    lastUpdated
    message
    players {
      id
      displayName
      firstName
      lastName
      shortName
      country
      countryFlag
      amateur
      withdrawn
      status
      owgr
    }
    alternates {
      id
      displayName
      firstName
      lastName
      shortName
      country
      countryFlag
      amateur
      withdrawn
      status
      owgr
    }
  }
}`,
		Variables: map[string]interface{}{
			"fieldId":          fieldID,
			"includeWithdrawn": includeWithdrawn,
			"changesOnly":      changesOnly,
		},
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		Post("")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch field: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("API error fetching field: %s", resp.Status())
	}
	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("API error fetching field: %s", response.Errors[0].Message)
	}

	return &response.Data.Field, nil
}

func (c *Client) GetUpcomingSchedule(ctx context.Context, tourCode string, year int) ([]ScheduleTournament, error) {
	var response struct {
		Data struct {
			UpcomingSchedule struct {
				Tournaments []ScheduleTournament `json:"tournaments"`
			} `json:"upcomingSchedule"`
		} `json:"data"`
		Errors []graphQLError `json:"errors"`
	}

	yearVar := ""
	if year > 0 {
		yearVar = fmt.Sprintf("%d", year)
	}

	req := graphQLRequest{
		OperationName: "UpcomingSchedule",
		Query: `query UpcomingSchedule($tourCode: String!, $year: String) {
  upcomingSchedule(tourCode: $tourCode, year: $year) {
    tournaments {
      id
      startDate
      tournamentName
      tournamentStatus
    }
  }
}`,
		Variables: map[string]interface{}{
			"tourCode": tourCode,
			"year":     yearVar,
		},
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		Post("")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch upcoming schedule: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("API error fetching upcoming schedule: %s", resp.Status())
	}
	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("API error fetching upcoming schedule: %s", response.Errors[0].Message)
	}

	return response.Data.UpcomingSchedule.Tournaments, nil
}
