package seed

import "testing"

func TestValidateTeams(t *testing.T) {
	if err := validateTeams(teams); err != nil {
		t.Fatal(err)
	}
	if len(teams) != 36 {
		t.Fatalf("expected 36 seed teams, got %d", len(teams))
	}
}

func TestValidateTeamsRejectsInvalidShortName(t *testing.T) {
	for _, value := range []string{"", "AB", "ABCD", "AbC", "A1C", "A C", "АБВ"} {
		t.Run(value, func(t *testing.T) {
			if err := validateTeams([]team{{"Team", value, "team"}}); err == nil {
				t.Fatal("invalid short_name accepted")
			}
		})
	}
}

func TestValidateTeamsRejectsDuplicateSlug(t *testing.T) {
	if err := validateTeams([]team{{"First", "ONE", "same"}, {"Second", "TWO", "same"}}); err == nil {
		t.Fatal("duplicate slug accepted")
	}
}

func TestValidateTeamsRequiresNameAndSlug(t *testing.T) {
	for _, item := range []team{{"", "ONE", "one"}, {"One", "ONE", ""}} {
		if err := validateTeams([]team{item}); err == nil {
			t.Fatal("missing name or slug accepted")
		}
	}
}
