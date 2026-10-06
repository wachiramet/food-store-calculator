package main

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "items with member card",
			input: "orange 5\nred\n\ny\n",
			want: []string{
				"Subtotal: 650.00 THB",
				"Bundle orange x2 (5%): -24.00 THB",
				"Member card (10%): -62.60 THB",
				"Total: 563.40 THB",
			},
		},
		{
			name:  "quantity defaults to one",
			input: "red\ngreen\n\nn\n",
			want:  []string{"Total: 90.00 THB"},
		},
		{
			name:  "same item entered twice adds up",
			input: "orange\norange\n\nn\n",
			want:  []string{"Total: 228.00 THB"},
		},
		{
			name:  "item name and answer ignore case",
			input: "Orange 2\n\nYES\n",
			want:  []string{"Total: 205.20 THB"},
		},
		{
			name:  "unknown item is reported and skipped",
			input: "black 2\nred\n\nn\n",
			want:  []string{"unknown item: black", "Total: 50.00 THB"},
		},
		{
			name:  "invalid quantity is reported and skipped",
			input: "red x\nred 0\nred -1\nred\n\nn\n",
			want: []string{
				"invalid quantity: x",
				"invalid quantity: 0",
				"invalid quantity: -1",
				"Total: 50.00 THB",
			},
		},
		{
			name:  "too many fields is reported and skipped",
			input: "red 1 2\nred\n\nn\n",
			want:  []string{`invalid entry "red 1 2"`, "Total: 50.00 THB"},
		},
		{
			name:  "empty order",
			input: "\nn\n",
			want:  []string{"Total: 0.00 THB"},
		},
		{
			name:  "input ends without answers",
			input: "red 2",
			want:  []string{"Total: 100.00 THB"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runOutput(t, tt.input)

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("output %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestRunHidesDiscountsThatDoNotApply(t *testing.T) {
	got := runOutput(t, "red\n\nn\n")

	for _, unwanted := range []string{"(5%)", "(10%)"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("output %q contains %q", got, unwanted)
		}
	}
}

func TestRunShowsMenu(t *testing.T) {
	got := runOutput(t, "")

	want := "Menu: blue 30.00, green 40.00, orange 120.00, pink 80.00, purple 90.00, red 50.00, yellow 50.00"
	if !strings.Contains(got, want) {
		t.Errorf("output %q does not contain %q", got, want)
	}
}

func TestRunShowsBundles(t *testing.T) {
	got := runOutput(t, "")

	want := "Bundle: orange, pink, green (5% off every 2 of the same item)"
	if !strings.Contains(got, want) {
		t.Errorf("output %q does not contain %q", got, want)
	}
}

func runOutput(t *testing.T, input string) string {
	t.Helper()

	var out strings.Builder
	if err := run(strings.NewReader(input), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return strings.Join(strings.Fields(out.String()), " ")
}
