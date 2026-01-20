package openai

import "fmt"

func earningsAgentPrompt(tournament string, year int, golfersJSON string) string {
	return fmt.Sprintf(`You are a golf data research agent. Find the final results and prize money for the %d %s.

Search authoritative sources like pgatour.com, espn.com, or golfchannel.com.

Here are the golfers to look up, each with a unique golfer_id:
%s

For each golfer you find in the tournament results, return:
- golfer_id: Copy the exact golfer_id value from the input list (this is a database identifier, not related to placement)
- position: Their finishing position in the tournament (1 = winner, 2 = second place, etc)
- earnings: Prize money in USD as an integer

Match golfers by comparing the name, first_name, and last_name fields from the input to the names in the search results.
If the tournament is not yet complete or results are unavailable, return an empty results array.`, year, tournament, golfersJSON)
}
