package sync

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/emailreminder"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/pick"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/ent/user"
	"github.com/rj-davidson/greenrats/internal/email"
)

func (i *Ingester) sendPickReminders(ctx context.Context) {
	log.Println("Checking for pick reminders to send...")

	now := time.Now().UTC()
	reminderWindowStart := now.Add(24 * time.Hour)
	reminderWindowEnd := now.Add(27 * time.Hour)

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtGTE(reminderWindowStart),
			tournament.PickWindowClosesAtLTE(reminderWindowEnd),
		).
		All(ctx)
	if err != nil {
		log.Printf("failed to query upcoming tournaments: %v", err)
		i.captureJobError("send_pick_reminders", err)
		return
	}

	if len(tournaments) == 0 {
		log.Println("No tournaments starting within reminder window")
		return
	}

	for _, t := range tournaments {
		i.sendRemindersForTournament(ctx, t)
	}

	log.Println("Pick reminder check completed")
}

func (i *Ingester) sendRemindersForTournament(ctx context.Context, t *ent.Tournament) {
	log.Printf("Sending pick reminders for tournament: %s", t.Name)

	leagues, err := i.db.League.Query().
		Where(league.SeasonYearEQ(t.SeasonYear)).
		All(ctx)
	if err != nil {
		log.Printf("failed to query leagues: %v", err)
		i.captureJobError("send_pick_reminders", err)
		return
	}

	for _, l := range leagues {
		i.sendRemindersForLeagueTournament(ctx, l, t)
	}
}

func (i *Ingester) sendRemindersForLeagueTournament(ctx context.Context, l *ent.League, t *ent.Tournament) {
	memberships, err := i.db.LeagueMembership.Query().
		Where(leaguemembership.HasLeagueWith(league.IDEQ(l.ID))).
		WithUser().
		All(ctx)
	if err != nil {
		log.Printf("failed to query league memberships: %v", err)
		i.captureJobError("send_pick_reminders", err)
		return
	}

	for _, m := range memberships {
		if m.Edges.User == nil {
			continue
		}
		u := m.Edges.User

		if u.DisplayName == nil {
			continue
		}

		hasPick, err := i.db.Pick.Query().
			Where(
				pick.HasUserWith(user.IDEQ(u.ID)),
				pick.HasTournamentWith(tournament.IDEQ(t.ID)),
				pick.HasLeagueWith(league.IDEQ(l.ID)),
			).
			Exist(ctx)
		if err != nil {
			log.Printf("failed to check pick: %v", err)
			i.captureJobError("send_pick_reminders", err)
			continue
		}
		if hasPick {
			continue
		}

		alreadySent, err := i.db.EmailReminder.Query().
			Where(
				emailreminder.HasUserWith(user.IDEQ(u.ID)),
				emailreminder.HasTournamentWith(tournament.IDEQ(t.ID)),
				emailreminder.HasLeagueWith(league.IDEQ(l.ID)),
				emailreminder.ReminderTypeEQ(emailreminder.ReminderTypePickReminder),
			).
			Exist(ctx)
		if err != nil {
			log.Printf("failed to check reminder status: %v", err)
			i.captureJobError("send_pick_reminders", err)
			continue
		}
		if alreadySent {
			continue
		}

		if err := i.sendPickReminderEmail(ctx, u, l, t); err != nil {
			log.Printf("failed to send reminder to %s: %v", u.Email, err)
			i.captureJobError("send_pick_reminders", err)
			continue
		}

		_, err = i.db.EmailReminder.Create().
			SetUserID(u.ID).
			SetTournamentID(t.ID).
			SetLeagueID(l.ID).
			SetReminderType(emailreminder.ReminderTypePickReminder).
			Save(ctx)
		if err != nil {
			log.Printf("failed to record reminder: %v", err)
			i.captureJobError("send_pick_reminders", err)
		}
	}
}

func (i *Ingester) sendPickReminderEmail(_ context.Context, u *ent.User, l *ent.League, t *ent.Tournament) error {
	displayName := ""
	if u.DisplayName != nil {
		displayName = *u.DisplayName
	}

	deadline := ""
	if t.PickWindowClosesAt != nil {
		deadline = t.PickWindowClosesAt.Format("Monday, January 2 at 3:04 PM MST")
	} else {
		deadline = t.StartDate.Format("Monday, January 2 at 3:04 PM MST")
	}

	data := email.PickReminderData{
		DisplayName:    displayName,
		TournamentName: t.Name,
		LeagueName:     l.Name,
		Deadline:       deadline,
	}

	return i.email.SendPickReminder(u.Email, data)
}

func (i *Ingester) sendTournamentResultsEmails(ctx context.Context, t *ent.Tournament) {
	log.Printf("Sending tournament results emails for: %s", t.Name)

	winnerEntry, err := i.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.IDEQ(t.ID)),
			tournamententry.PositionEQ(1),
		).
		WithGolfer().
		First(ctx)

	tournamentWinner := "Unknown"
	if err == nil && winnerEntry.Edges.Golfer != nil {
		tournamentWinner = winnerEntry.Edges.Golfer.Name
	}

	picks, err := i.db.Pick.Query().
		Where(pick.HasTournamentWith(tournament.IDEQ(t.ID))).
		WithUser().
		WithGolfer().
		WithLeague().
		All(ctx)
	if err != nil {
		log.Printf("failed to query picks: %v", err)
		i.captureJobError("send_tournament_results", err)
		return
	}

	for _, p := range picks {
		if p.Edges.User == nil || p.Edges.League == nil || p.Edges.Golfer == nil {
			continue
		}
		if p.Edges.User.DisplayName == nil {
			continue
		}

		alreadySent, err := i.db.EmailReminder.Query().
			Where(
				emailreminder.HasUserWith(user.IDEQ(p.Edges.User.ID)),
				emailreminder.HasTournamentWith(tournament.IDEQ(t.ID)),
				emailreminder.HasLeagueWith(league.IDEQ(p.Edges.League.ID)),
				emailreminder.ReminderTypeEQ(emailreminder.ReminderTypeTournamentResults),
			).
			Exist(ctx)
		if err != nil {
			log.Printf("failed to check reminder status: %v", err)
			i.captureJobError("send_tournament_results", err)
			continue
		}
		if alreadySent {
			continue
		}

		golferEntry, _ := i.db.TournamentEntry.Query().
			Where(
				tournamententry.HasTournamentWith(tournament.IDEQ(t.ID)),
				tournamententry.HasGolferWith(golfer.IDEQ(p.Edges.Golfer.ID)),
			).
			Only(ctx)

		position := "N/A"
		earnings := "$0"
		if golferEntry != nil {
			if golferEntry.Cut {
				position = "CUT"
			} else if golferEntry.Position > 0 {
				position = fmt.Sprintf("%d", golferEntry.Position)
			}
			earnings = formatCurrency(golferEntry.Earnings)
		}

		userRank, totalEarnings := i.calculateLeagueStandings(ctx, p.Edges.User.ID, p.Edges.League.ID, t.SeasonYear)

		data := &email.TournamentResultsData{
			DisplayName:      *p.Edges.User.DisplayName,
			TournamentName:   t.Name,
			TournamentWinner: tournamentWinner,
			LeagueName:       p.Edges.League.Name,
			GolferName:       p.Edges.Golfer.Name,
			GolferPosition:   position,
			GolferEarnings:   earnings,
			UserRank:         userRank,
			TotalEarnings:    formatCurrency(totalEarnings),
		}

		if err := i.email.SendTournamentResults(p.Edges.User.Email, data); err != nil {
			log.Printf("failed to send results to %s: %v", p.Edges.User.Email, err)
			i.captureJobError("send_tournament_results", err)
			continue
		}

		_, err = i.db.EmailReminder.Create().
			SetUserID(p.Edges.User.ID).
			SetTournamentID(t.ID).
			SetLeagueID(p.Edges.League.ID).
			SetReminderType(emailreminder.ReminderTypeTournamentResults).
			Save(ctx)
		if err != nil {
			log.Printf("failed to record reminder: %v", err)
			i.captureJobError("send_tournament_results", err)
		}
	}

	log.Printf("Finished sending tournament results emails for: %s", t.Name)
}

func (i *Ingester) calculateLeagueStandings(ctx context.Context, userID, leagueID uuid.UUID, seasonYear int) (rank, totalEarnings int) {
	type userEarnings struct {
		userID   uuid.UUID
		earnings int
	}

	picks, err := i.db.Pick.Query().
		Where(
			pick.HasLeagueWith(league.IDEQ(leagueID)),
			pick.SeasonYearEQ(seasonYear),
		).
		WithUser().
		WithGolfer().
		WithTournament().
		All(ctx)
	if err != nil {
		return 0, 0
	}

	earningsMap := make(map[uuid.UUID]int)
	for _, p := range picks {
		if p.Edges.User == nil || p.Edges.Golfer == nil || p.Edges.Tournament == nil {
			continue
		}
		entry, err := i.db.TournamentEntry.Query().
			Where(
				tournamententry.HasTournamentWith(tournament.IDEQ(p.Edges.Tournament.ID)),
				tournamententry.HasGolferWith(golfer.IDEQ(p.Edges.Golfer.ID)),
			).
			Only(ctx)
		if err == nil {
			earningsMap[p.Edges.User.ID] += entry.Earnings
		}
	}

	var allEarnings []userEarnings
	for uid, e := range earningsMap {
		allEarnings = append(allEarnings, userEarnings{userID: uid, earnings: e})
	}

	for j := 0; j < len(allEarnings); j++ {
		for k := j + 1; k < len(allEarnings); k++ {
			if allEarnings[k].earnings > allEarnings[j].earnings {
				allEarnings[j], allEarnings[k] = allEarnings[k], allEarnings[j]
			}
		}
	}

	userTotal := earningsMap[userID]
	userRank := 1
	for _, e := range allEarnings {
		if e.userID == userID {
			break
		}
		if e.earnings > userTotal {
			userRank++
		}
	}

	return userRank, userTotal
}
