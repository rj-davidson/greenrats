package openai

import "fmt"

func earningsAgentPrompt(tournament string, year int, golfersJSON string) string {
	return fmt.Sprintf(`You are a golf data research agent. Find the final results and prize money for the %d %s.

Search authoritative sources like pgatour.com, espn.com, or golfchannel.com.

Here are the golfers to look up, each with a unique golfer_id:
%s

For each golfer you find in the tournament results, return:
- golfer_id: Copy the exact golfer_id value from the input list (this is a database identifier, not related to placement)
- earnings: Prize money in USD as an integer

Match golfers by comparing the name, first_name, and last_name fields from the input to the names in the search results.
If the tournament is not yet complete or results are unavailable, return an empty results array.`, year, tournament, golfersJSON)
}

func leaderboardSearchPrompt(tournament string, year int) string {
	return fmt.Sprintf(`You are a golf data research agent. Find the complete final leaderboard with prize money for the %d %s.

Search authoritative sources like pgatour.com, espn.com, or golfchannel.com.

Return ALL players who earned prize money in the tournament. For each player, return:
- name: The player's full name as shown in the results
- earnings: Prize money earned in USD as an integer

If the tournament is not yet complete or results are unavailable, return an empty entries array.`, year, tournament)
}

func matchPlayersPrompt(leaderboardJSON, golfersJSON string) string {
	return fmt.Sprintf(`You are a name matching agent. Match golfers from the input list to the tournament leaderboard.

Tournament leaderboard (names and earnings):
%s

Golfers to match (each with a unique golfer_id):
%s

For each golfer in the input list, find the matching entry in the leaderboard by comparing names. Use fuzzy matching to handle name variations (e.g., "J.J. Spaun" matches "JJ Spaun", "Byeong Hun An" matches "Byeong-Hun An").

For each matched golfer, return:
- golfer_id: Copy the exact golfer_id value from the input list
- earnings: The earnings from the matched leaderboard entry

Only return golfers that you can confidently match. If no match is found for a golfer, omit them from the results.`, leaderboardJSON, golfersJSON)
}

func parseLeaderboardContentPrompt(content, tournamentName string) string {
	return fmt.Sprintf(`You are a golf data extraction agent. Parse the following webpage content to extract the leaderboard with prize money for the %s.

Content:
%s

Return ALL players who earned prize money in the tournament. For each player, return:
- name: The player's full name as shown in the results
- earnings: Prize money earned in USD as an integer (parse from strings like "$3,600,000" to 3600000)

If the content does not contain earnings data or the tournament results are not available, return an empty entries array.`, tournamentName, content)
}
