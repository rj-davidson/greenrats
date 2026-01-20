package testutil

import "time"

const PickWindowDays = 3

func PickWindowOpen(tournamentStart time.Time) time.Time {
	return tournamentStart.AddDate(0, 0, -PickWindowDays).Add(time.Hour)
}

func PickWindowClosed(tournamentStart time.Time) time.Time {
	return tournamentStart.Add(time.Hour)
}

func BeforePickWindow(tournamentStart time.Time) time.Time {
	return tournamentStart.AddDate(0, 0, -PickWindowDays-1)
}

func TournamentInPickWindow() time.Time {
	return time.Now().AddDate(0, 0, 2)
}

func TournamentOutsidePickWindow() time.Time {
	return time.Now().AddDate(0, 0, 10)
}

func TournamentStarted() time.Time {
	return time.Now().AddDate(0, 0, -1)
}
