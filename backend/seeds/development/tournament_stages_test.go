package development

import (
	"testing"
	"time"
	"unicode/utf8"

	"preditto/internal/model"
)

func TestChampionsLeagueStageSchedule(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	stages, err := championsLeagueStages(now)
	if err != nil {
		t.Fatal(err)
	}
	// Expected UTC deadlines and estimated finishes derived from the UEFA
	// calendar, including the autumn/spring DST changes and the early final.
	want := []struct {
		code, close, end string
		points           int
	}{
		{"league-1", "2026-09-08 14:45", "2026-09-10 22:00", 1},
		{"league-2", "2026-10-13 14:45", "2026-10-14 22:00", 1},
		{"league-3", "2026-10-20 14:45", "2026-10-21 22:00", 1},
		{"league-4", "2026-11-03 15:45", "2026-11-04 23:00", 1},
		{"league-5", "2026-11-24 15:45", "2026-11-25 23:00", 1},
		{"league-6", "2026-12-08 15:45", "2026-12-09 23:00", 1},
		{"league-7", "2027-01-19 15:45", "2027-01-20 23:00", 1},
		{"league-8", "2027-01-27 18:00", "2027-01-27 23:00", 1},
		{"round-32-1", "2027-02-16 15:45", "2027-02-17 23:00", 2},
		{"round-32-2", "2027-02-23 15:45", "2027-02-24 23:00", 2},
		{"round-16-1", "2027-03-09 15:45", "2027-03-10 23:00", 3},
		{"round-16-2", "2027-03-16 15:45", "2027-03-17 23:00", 3},
		{"quarter-final-1", "2027-04-06 17:00", "2027-04-07 22:00", 4},
		{"quarter-final-2", "2027-04-13 17:00", "2027-04-14 22:00", 4},
		{"semi-final-1", "2027-04-27 17:00", "2027-04-28 22:00", 5},
		{"semi-final-2", "2027-05-04 17:00", "2027-05-05 22:00", 5},
		{"final", "2027-06-05 14:00", "2027-06-05 19:00", 6},
	}
	if len(stages) != 17 {
		t.Fatalf("got %d stages, want 17", len(stages))
	}
	seen := make(map[string]bool)
	for i, stage := range stages {
		expected := want[i]
		kind, status := "play-off", model.TournamentStatusPlanned
		if i < 8 {
			kind = "group"
		}
		if i < 4 {
			status = model.TournamentStatusFinished
		} else if i == 4 {
			status = model.TournamentStatusActive
		}
		if seen[stage.code] || stage.code != expected.code || stage.sequenceNo != i+1 || stage.kind != kind ||
			stage.status != status || stage.name == "" || utf8.RuneCountInString(stage.name) > 150 ||
			len(stage.code) > 50 || stage.outcomePoints != expected.points || stage.exactScorePoints != 3*expected.points {
			t.Fatalf("invalid stage %d: %+v", i+1, stage)
		}
		seen[stage.code] = true
		open := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
		if i > 0 {
			previousEnd, err := time.Parse("2006-01-02 15:04", want[i-1].end)
			if err != nil {
				t.Fatal(err)
			}
			// The fixture end includes three hours; opening is kickoff + eight.
			open = previousEnd.Add(5 * time.Hour)
		}
		if !stage.predictionsOpenAt.Equal(open) || !stage.predictionsOpenAt.Before(stage.predictionsCloseAt) ||
			stage.firstKickoff.Sub(stage.predictionsCloseAt) != 2*time.Hour ||
			stage.predictionsCloseAt.Format("2006-01-02 15:04") != expected.close ||
			stage.endsAt.Format("2006-01-02 15:04") != expected.end ||
			stage.predictionsOpenAt.Location() != time.UTC || stage.predictionsCloseAt.Location() != time.UTC {
			t.Fatalf("invalid UTC window for %s: %+v", stage.code, stage)
		}
	}
}

func TestChampionsLeagueStageStatusBoundaries(t *testing.T) {
	stages, err := championsLeagueStages(time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	check := func(now time.Time, finished int) {
		t.Helper()
		got, err := championsLeagueStages(now)
		if err != nil {
			t.Fatal(err)
		}
		for i, stage := range got {
			want := model.TournamentStatusPlanned
			if i < finished {
				want = model.TournamentStatusFinished
			} else if i == finished {
				want = model.TournamentStatusActive
			}
			if stage.status != want {
				t.Fatalf("at %s stage %s: status=%s want=%s", now, stage.code, stage.status, want)
			}
		}
	}
	check(time.Time{}, 0) // The nearest stage is active even before predictions open.
	for i, stage := range stages {
		check(stage.predictionsOpenAt, i)
		check(stage.predictionsCloseAt, i)
		check(stage.firstKickoff, i)
		check(stage.endsAt.Add(-time.Nanosecond), i)
		check(stage.endsAt, i+1)
		check(stage.endsAt.In(time.FixedZone("UTC+4", 4*60*60)), i+1)
	}
	check(stages[16].endsAt.Add(365*24*time.Hour), 17)
}
