package email

import "fmt"

const baseStyle = `
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
  .header { background: #16a34a; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
  .content { background: #f9fafb; padding: 20px; border-radius: 0 0 8px 8px; }
  .button { display: inline-block; background: #16a34a; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 10px 0; }
  .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 20px; }
  .highlight { background: #ecfdf5; padding: 15px; border-radius: 6px; margin: 15px 0; }
</style>
`

func wrapHTML(body string) string {
	return fmt.Sprintf(`<!DOCTYPE html><html><head>%s</head><body>%s<div class="footer"><p>GreenRats - Golf Pick'em</p></div></body></html>`, baseStyle, body)
}

func renderWelcome(data WelcomeData) string {
	body := fmt.Sprintf(`
<div class="header">
  <h1>Welcome to GreenRats!</h1>
</div>
<div class="content">
  <p>Hey %s,</p>
  <p>Welcome to GreenRats! You're all set to start picking golfers and competing with friends.</p>
  <div class="highlight">
    <strong>How it works:</strong>
    <ul>
      <li>Pick one golfer per tournament</li>
      <li>Once picked, you can't use that golfer again this season</li>
      <li>Earn points based on your golfer's performance</li>
      <li>Compete on league leaderboards</li>
    </ul>
  </div>
  <p>Join or create a league to get started!</p>
  <a href="https://greenrats.com" class="button">Go to GreenRats</a>
</div>
`, data.DisplayName)
	return wrapHTML(body)
}

func renderLeagueJoin(data LeagueJoinData) string {
	var body string
	if data.IsNewMember {
		body = fmt.Sprintf(`
<div class="header">
  <h1>You're In!</h1>
</div>
<div class="content">
  <p>Hey %s,</p>
  <p>You've successfully joined <strong>%s</strong>!</p>
  <p>You can now make picks and compete on the league leaderboard.</p>
  <a href="https://greenrats.com" class="button">View League</a>
</div>
`, data.DisplayName, data.LeagueName)
	} else {
		body = fmt.Sprintf(`
<div class="header">
  <h1>New Member!</h1>
</div>
<div class="content">
  <p>Hey %s,</p>
  <p><strong>%s</strong> has joined your league <strong>%s</strong>!</p>
  <a href="https://greenrats.com" class="button">View League</a>
</div>
`, data.DisplayName, data.NewMemberName, data.LeagueName)
	}
	return wrapHTML(body)
}

func renderPickReminder(data PickReminderData) string {
	body := fmt.Sprintf(`
<div class="header">
  <h1>Pick Reminder</h1>
</div>
<div class="content">
  <p>Hey %s,</p>
  <p>Don't forget to make your pick for <strong>%s</strong> in <strong>%s</strong>!</p>
  <div class="highlight">
    <strong>Deadline:</strong> %s
  </div>
  <p>Once the tournament starts, you won't be able to make or change your pick.</p>
  <a href="https://greenrats.com" class="button">Make Your Pick</a>
</div>
`, data.DisplayName, data.TournamentName, data.LeagueName, data.Deadline)
	return wrapHTML(body)
}

func renderTournamentResults(data *TournamentResultsData) string {
	body := fmt.Sprintf(`
<div class="header">
  <h1>%s Results</h1>
</div>
<div class="content">
  <p>Hey %s,</p>
  <p>The tournament has concluded! Here's how your pick performed:</p>
  <div class="highlight">
    <p><strong>Tournament Winner:</strong> %s</p>
    <p><strong>Your Pick:</strong> %s</p>
    <p><strong>Position:</strong> %s</p>
    <p><strong>Earnings:</strong> %s</p>
  </div>
  <p>In <strong>%s</strong>, you're now ranked <strong>#%d</strong> with total earnings of <strong>%s</strong>.</p>
  <a href="https://greenrats.com" class="button">View Leaderboard</a>
</div>
`, data.TournamentName, data.DisplayName, data.TournamentWinner, data.GolferName, data.GolferPosition, data.GolferEarnings, data.LeagueName, data.UserRank, data.TotalEarnings)
	return wrapHTML(body)
}

func renderCommissionerAction(data CommissionerActionData) string {
	body := fmt.Sprintf(`
<div class="header">
  <h1>Commissioner Update</h1>
</div>
<div class="content">
  <p>Hey %s,</p>
  <p>The commissioner of <strong>%s</strong> has made a change:</p>
  <div class="highlight">
    <p><strong>Commissioner:</strong> %s</p>
    <p><strong>Action:</strong> %s</p>
  </div>
  <a href="https://greenrats.com" class="button">View League</a>
</div>
`, data.DisplayName, data.LeagueName, data.CommissionerName, data.ActionDescription)
	return wrapHTML(body)
}
