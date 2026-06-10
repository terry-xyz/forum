package static_test

import (
	"os"
	"strings"
	"testing"
)

func TestCheckboxesAreExcludedFromFullWidthInputControls(t *testing.T) {
	css, err := os.ReadFile("style.css")
	if err != nil {
		t.Fatal(err)
	}

	body := string(css)
	if !strings.Contains(body, `input:not([type="hidden"]):not([type="checkbox"])`) {
		t.Fatal("style.css should exclude checkboxes from full-width text input styles")
	}
	if strings.Contains(body, "input:not([type=\"hidden\"]) {\n") {
		t.Fatal("style.css still applies full-width text input styles to checkboxes")
	}
}
