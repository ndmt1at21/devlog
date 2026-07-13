package mysql

import (
	"strings"
	"testing"

	"github.com/ndmt1at21/devlog/backend/migrations"
)

func TestSplitStatementsDropsCommentSemicolons(t *testing.T) {
	sql := "-- header with a semicolon; prose continues\n" +
		"-- more comment\n" +
		"CREATE TABLE a (\n    id INT -- keep inline comments\n);\n\n" +
		"CREATE INDEX i ON a(id);\n"
	got := splitStatements(sql)
	if len(got) != 2 {
		t.Fatalf("statements = %d, want 2: %q", len(got), got)
	}
	if !strings.Contains(got[0], "CREATE TABLE a") || strings.Contains(got[0], "header") {
		t.Errorf("first statement mangled: %q", got[0])
	}
	if !strings.Contains(got[1], "CREATE INDEX i") {
		t.Errorf("second statement mangled: %q", got[1])
	}
}

// Every embedded migration must split into statements that don't start inside
// a comment — the failure mode that broke 0003_reactions.up.sql (its header
// comment contains a ";", so the naive splitter executed comment prose).
func TestSplitStatementsOnEmbeddedMigrations(t *testing.T) {
	entries, err := migrations.MySQL.ReadDir("mysql")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		content, err := migrations.MySQL.ReadFile("mysql/" + e.Name())
		if err != nil {
			t.Fatal(err)
		}
		for _, stmt := range splitStatements(string(content)) {
			first := strings.TrimSpace(stmt)
			if first == "" {
				t.Errorf("%s: empty statement survived", e.Name())
				continue
			}
			up := strings.ToUpper(first)
			if !strings.HasPrefix(up, "CREATE") && !strings.HasPrefix(up, "ALTER") &&
				!strings.HasPrefix(up, "DROP") && !strings.HasPrefix(up, "INSERT") &&
				!strings.HasPrefix(up, "UPDATE") {
				t.Errorf("%s: statement starts with non-SQL text: %.60q", e.Name(), first)
			}
		}
	}
}
