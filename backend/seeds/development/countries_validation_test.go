package development

import "testing"

func TestValidateCountries(t *testing.T) {
	if len(countries) != 16 {
		t.Fatalf("countries=%d, want 16", len(countries))
	}
	if err := validateCountries(countries); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCountriesRejectsInvalidData(t *testing.T) {
	for _, code := range []string{"", "EN", "ENGL", "Eng", "E1G", "E_G"} {
		t.Run("code_"+code, func(t *testing.T) {
			if err := validateCountries([]country{{"England", code}}); err == nil {
				t.Fatal("expected invalid code error")
			}
		})
	}
	if err := validateCountries([]country{{"", "ENG"}}); err == nil {
		t.Fatal("expected missing name error")
	}
	if err := validateCountries([]country{{"England", "ENG"}, {"Other", "ENG"}}); err == nil {
		t.Fatal("expected duplicate code error")
	}
}
