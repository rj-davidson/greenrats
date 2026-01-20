package demo

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/ent/user"
	"github.com/rj-davidson/greenrats/internal/resources"
)

const (
	demoLeagueIDString   = "00000000-0000-0000-0000-000000000007"
	demoLeagueName       = "Prothero Demo League"
	demoJoinCode         = "000000"
	demoOwnerWorkosID    = "user_01KFA68NF21EWENM8PFYHCQBGE"
	demoOwnerEmail       = "dvdsn.rj@gmail.com"
	demoOwnerDisplayName = "RJ Davidson"
	demoTournamentBDLID  = 7
)

type demoPickRow struct {
	username   string
	golferName string
}

func EnsureDemoLeague(ctx context.Context, db *ent.Client) error {
	demoLeagueID, err := uuid.FromString(demoLeagueIDString)
	if err != nil {
		return fmt.Errorf("invalid demo league id: %w", err)
	}

	exists, err := db.League.Query().Where(league.IDEQ(demoLeagueID)).Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check demo league: %w", err)
	}
	if exists {
		return nil
	}

	joinCodeTaken, err := db.League.Query().Where(league.CodeEQ(demoJoinCode)).Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check demo join code: %w", err)
	}
	if joinCodeTaken {
		return fmt.Errorf("demo join code %s already in use", demoJoinCode)
	}

	tournamentEnt, err := db.Tournament.Query().
		Where(tournament.BdlIDEQ(demoTournamentBDLID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("demo tournament bdl_id=%d not found", demoTournamentBDLID)
		}
		return fmt.Errorf("failed to load demo tournament: %w", err)
	}

	rows, err := loadDemoPicks()
	if err != nil {
		return err
	}

	tx, err := db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start demo league transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	owner, err := ensureDemoOwner(ctx, tx)
	if err != nil {
		return err
	}

	leagueEnt, err := tx.League.Create().
		SetID(demoLeagueID).
		SetName(demoLeagueName).
		SetCode(demoJoinCode).
		SetSeasonYear(tournamentEnt.SeasonYear).
		SetJoiningEnabled(true).
		SetCreatedByID(owner.ID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create demo league: %w", err)
	}

	if _, err := tx.LeagueMembership.Create().
		SetUserID(owner.ID).
		SetLeagueID(leagueEnt.ID).
		SetRole(leaguemembership.RoleOwner).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to create demo owner membership: %w", err)
	}

	created := 0
	skippedGolfer := 0
	skippedEntry := 0

	for _, row := range rows {
		member, err := ensureDemoMember(ctx, tx, row.username)
		if err != nil {
			return err
		}

		if _, err := tx.LeagueMembership.Create().
			SetUserID(member.ID).
			SetLeagueID(leagueEnt.ID).
			SetRole(leaguemembership.RoleMember).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to create demo member %s: %w", row.username, err)
		}

		golferEnt, err := findGolferByName(ctx, tx, row.golferName)
		if err != nil {
			return err
		}
		if golferEnt == nil {
			skippedGolfer++
			continue
		}

		inField, err := tx.TournamentEntry.Query().
			Where(
				tournamententry.HasTournamentWith(tournament.IDEQ(tournamentEnt.ID)),
				tournamententry.HasGolferWith(golfer.IDEQ(golferEnt.ID)),
			).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check tournament entry for %s: %w", row.golferName, err)
		}
		if !inField {
			skippedEntry++
			continue
		}

		if _, err := tx.Pick.Create().
			SetUserID(member.ID).
			SetLeagueID(leagueEnt.ID).
			SetTournamentID(tournamentEnt.ID).
			SetGolferID(golferEnt.ID).
			SetSeasonYear(tournamentEnt.SeasonYear).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to create pick for %s: %w", row.username, err)
		}
		created++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit demo league: %w", err)
	}
	committed = true

	log.Printf(
		"Demo league seeded: league_id=%s picks=%d skipped_golfers=%d skipped_entries=%d",
		demoLeagueIDString,
		created,
		skippedGolfer,
		skippedEntry,
	)

	return nil
}

func ensureDemoOwner(ctx context.Context, tx *ent.Tx) (*ent.User, error) {
	owner, err := tx.User.Query().
		Where(user.WorkosIDEQ(demoOwnerWorkosID)).
		Only(ctx)
	if err == nil {
		return owner, nil
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to query demo owner by workos_id: %w", err)
	}

	owner, err = tx.User.Query().
		Where(user.EmailEQ(demoOwnerEmail)).
		Only(ctx)
	if err == nil {
		return owner, nil
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to query demo owner by email: %w", err)
	}

	owner, err = tx.User.Create().
		SetWorkosID(demoOwnerWorkosID).
		SetEmail(demoOwnerEmail).
		SetDisplayName(demoOwnerDisplayName).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create demo owner: %w", err)
	}

	return owner, nil
}

func ensureDemoMember(ctx context.Context, tx *ent.Tx, username string) (*ent.User, error) {
	normalized := strings.TrimSpace(username)
	workosID := demoWorkosID(normalized)

	member, err := tx.User.Query().
		Where(user.WorkosIDEQ(workosID)).
		Only(ctx)
	if err == nil {
		return member, nil
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to query demo member %s: %w", normalized, err)
	}

	member, err = tx.User.Query().
		Where(user.DisplayNameEQ(normalized)).
		Only(ctx)
	if err == nil {
		return member, nil
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to query demo member by display_name %s: %w", normalized, err)
	}

	email := demoEmail(normalized)
	member, err = tx.User.Create().
		SetWorkosID(workosID).
		SetEmail(email).
		SetDisplayName(normalized).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create demo member %s: %w", normalized, err)
	}

	return member, nil
}

func demoWorkosID(username string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(username)))
	return fmt.Sprintf("demo:%x", sum)
}

func demoEmail(username string) string {
	if strings.Contains(username, "@") {
		return username
	}
	sum := sha256.Sum256([]byte(strings.ToLower(username)))
	return fmt.Sprintf("demo+%x@greenrats.local", sum[:8])
}

func loadDemoPicks() ([]demoPickRow, error) {
	raw := strings.TrimPrefix(string(resources.ProtheroLeagueCSV), "\ufeff")
	reader := csv.NewReader(strings.NewReader(raw))
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read demo picks header: %w", err)
	}

	index := make(map[string]int, len(header))
	for i, col := range header {
		index[strings.ToLower(strings.TrimSpace(col))] = i
	}

	usernameIdx, ok := index["username"]
	if !ok {
		return nil, errors.New("demo picks missing username column")
	}
	golferIdx, ok := index["pick_player_name"]
	if !ok {
		return nil, errors.New("demo picks missing pick_player_name column")
	}

	rows := make([]demoPickRow, 0, 256)
	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if !errors.Is(err, csv.ErrFieldCount) {
				return nil, fmt.Errorf("failed to read demo picks row: %w", err)
			}
			if len(record) == 0 {
				continue
			}
		}
		if len(record) <= golferIdx || len(record) <= usernameIdx {
			continue
		}
		username := strings.TrimSpace(record[usernameIdx])
		golferName := strings.TrimSpace(record[golferIdx])
		if username == "" || golferName == "" {
			continue
		}
		rows = append(rows, demoPickRow{
			username:   username,
			golferName: golferName,
		})
	}

	return rows, nil
}

func findGolferByName(ctx context.Context, tx *ent.Tx, name string) (*ent.Golfer, error) {
	for _, candidate := range golferNameCandidates(name) {
		golferEnt, err := tx.Golfer.Query().
			Where(golfer.NameEQ(candidate)).
			Only(ctx)
		if err == nil {
			return golferEnt, nil
		}
		if !ent.IsNotFound(err) {
			return nil, fmt.Errorf("failed to query golfer %s: %w", candidate, err)
		}
	}
	return nil, nil
}

func golferNameCandidates(name string) []string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil
	}

	candidates := []string{trimmed}
	if !strings.Contains(trimmed, ",") {
		return candidates
	}

	parts := strings.SplitN(trimmed, ",", 2)
	last := strings.TrimSpace(parts[0])
	first := strings.TrimSpace(parts[1])
	if last == "" || first == "" {
		return candidates
	}

	return append([]string{fmt.Sprintf("%s %s", first, last)}, candidates...)
}
