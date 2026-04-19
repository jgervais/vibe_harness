package rules

import (
	"testing"
)

func TestViolationValidate(t *testing.T) {
	tests := []struct {
		name    string
		v       Violation
		wantErr bool
	}{
		{
			name: "valid violation",
			v: Violation{
				RuleID:   "VH-G001",
				File:     "main.go",
				Line:     1,
				Column:   1,
				EndLine:  1,
				Message:  "file too long",
				Severity: "error",
			},
			wantErr: false,
		},
		{
			name: "valid warning",
			v: Violation{
				RuleID:   "VH-G005",
				File:     "secret.go",
				Line:     10,
				Severity: "warning",
			},
			wantErr: false,
		},
		{
			name: "valid note",
			v: Violation{
				RuleID:   "VH-G008",
				File:     "comment.go",
				Line:     1,
				Severity: "note",
			},
			wantErr: false,
		},
		{
			name: "invalid rule ID bad prefix",
			v: Violation{
				RuleID:   "XX-G001",
				File:     "main.go",
				Line:     1,
				Severity: "error",
			},
			wantErr: true,
		},
		{
			name: "invalid rule ID too few digits",
			v: Violation{
				RuleID:   "VH-G01",
				File:     "main.go",
				Line:     1,
				Severity: "error",
			},
			wantErr: true,
		},
		{
			name: "invalid rule ID too many digits",
			v: Violation{
				RuleID:   "VH-G0011",
				File:     "main.go",
				Line:     1,
				Severity: "error",
			},
			wantErr: true,
		},
		{
			name: "empty file",
			v: Violation{
				RuleID:   "VH-G001",
				File:     "",
				Line:     1,
				Severity: "error",
			},
			wantErr: true,
		},
		{
			name: "line zero",
			v: Violation{
				RuleID:   "VH-G001",
				File:     "main.go",
				Line:     0,
				Severity: "error",
			},
			wantErr: true,
		},
		{
			name: "negative line",
			v: Violation{
				RuleID:   "VH-G001",
				File:     "main.go",
				Line:     -5,
				Severity: "error",
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			v: Violation{
				RuleID:   "VH-G001",
				File:     "main.go",
				Line:     1,
				Severity: "critical",
			},
			wantErr: true,
		},
		{
			name: "empty severity",
			v: Violation{
				RuleID:   "VH-G001",
				File:     "main.go",
				Line:     1,
				Severity: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.v.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Violation.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckConstruction(t *testing.T) {
	c := Check{
		ID:          "VH-G001",
		Name:        "File Length",
		Description: "Files must not exceed 300 non-blank, non-comment lines",
		Severity:    "warning",
		RequiresAST: false,
		Threshold:   "300 lines",
	}
	if c.ID != "VH-G001" {
		t.Errorf("Check.ID = %q, want %q", c.ID, "VH-G001")
	}
	if c.Name != "File Length" {
		t.Errorf("Check.Name = %q, want %q", c.Name, "File Length")
	}
	if c.RequiresAST != false {
		t.Errorf("Check.RequiresAST = %v, want false", c.RequiresAST)
	}
}

func TestChecks(t *testing.T) {
	checks := Checks()
	if len(checks) != 6 {
		t.Fatalf("Checks() returned %d entries, want 6", len(checks))
	}

	wantIDs := []string{"VH-G001", "VH-G005", "VH-G006", "VH-G007", "VH-G008", "VH-G011"}
	wantSeverities := []string{"warning", "error", "warning", "warning", "note", "error"}

	for i, c := range checks {
		if c.ID != wantIDs[i] {
			t.Errorf("Checks()[%d].ID = %q, want %q", i, c.ID, wantIDs[i])
		}
		if c.Severity != wantSeverities[i] {
			t.Errorf("Checks()[%d].Severity = %q, want %q", i, c.Severity, wantSeverities[i])
		}
		if c.RequiresAST != false {
			t.Errorf("Checks()[%d].RequiresAST = %v, want false", i, c.RequiresAST)
		}
	}
}