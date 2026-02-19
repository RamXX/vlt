package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMomentToGoFormat(t *testing.T) {
	tests := []struct {
		moment string
		want   string
	}{
		{"YYYY-MM-DD", "2006-01-02"},
		{"YY-M-D", "06-1-2"},
		{"YYYY/MM/DD", "2006/01/02"},
		{"dddd, MMMM D, YYYY", "Monday, January 2, 2006"},
		{"ddd MMM DD", "Mon Jan 02"},
		{"YYYY-MM-DD HH:mm", "2006-01-02 15:04"},
	}

	for _, tt := range tests {
		t.Run(tt.moment, func(t *testing.T) {
			got := momentToGoFormat(tt.moment)
			if got != tt.want {
				t.Errorf("momentToGoFormat(%q) = %q, want %q", tt.moment, got, tt.want)
			}
		})
	}
}

func TestLoadDailyConfig_Default(t *testing.T) {
	vaultDir := t.TempDir()

	config := loadDailyConfig(vaultDir)

	if config.Format != "2006-01-02" {
		t.Errorf("default format = %q, want %q", config.Format, "2006-01-02")
	}
	if config.Folder != "" {
		t.Errorf("default folder = %q, want empty", config.Folder)
	}
	if config.Template != "" {
		t.Errorf("default template = %q, want empty", config.Template)
	}
}

func TestLoadDailyConfig_FromFile(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)
	os.WriteFile(
		filepath.Join(vaultDir, ".obsidian", "daily-notes.json"),
		[]byte(`{"folder":"daily","format":"YYYY/MM/DD","template":"_templates/daily"}`),
		0644,
	)

	config := loadDailyConfig(vaultDir)

	if config.Folder != "daily" {
		t.Errorf("folder = %q, want %q", config.Folder, "daily")
	}
	if config.Format != "2006/01/02" {
		t.Errorf("format = %q, want %q", config.Format, "2006/01/02")
	}
	if config.Template != "_templates/daily" {
		t.Errorf("template = %q, want %q", config.Template, "_templates/daily")
	}
}

func TestLoadDailyConfig_PeriodicNotes(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, ".obsidian", "plugins", "periodic-notes"), 0755)
	os.WriteFile(
		filepath.Join(vaultDir, ".obsidian", "plugins", "periodic-notes", "data.json"),
		[]byte(`{"daily":{"folder":"journal","format":"YYYY-MM-DD"}}`),
		0644,
	)

	config := loadDailyConfig(vaultDir)

	if config.Folder != "journal" {
		t.Errorf("folder = %q, want %q", config.Folder, "journal")
	}
	if config.Format != "2006-01-02" {
		t.Errorf("format = %q, want %q", config.Format, "2006-01-02")
	}
}

func TestCmdDaily_CreateNew(t *testing.T) {
	vaultDir := t.TempDir()

	params := map[string]string{}
	if err := cmdDaily(vaultDir, params); err != nil {
		t.Fatalf("daily create: %v", err)
	}

	// Should create today's note
	today := time.Now().Format("2006-01-02")
	notePath := filepath.Join(vaultDir, today+".md")

	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("daily note not created: %v", err)
	}

	if !strings.Contains(string(data), "# "+today) {
		t.Errorf("daily note content: %q, expected heading with date", string(data))
	}
}

func TestCmdDaily_ReadExisting(t *testing.T) {
	vaultDir := t.TempDir()

	today := time.Now().Format("2006-01-02")
	content := "# Existing Note\n\nSome content.\n"
	os.WriteFile(
		filepath.Join(vaultDir, today+".md"),
		[]byte(content),
		0644,
	)

	got := captureStdout(func() {
		if err := cmdDaily(vaultDir, map[string]string{}); err != nil {
			t.Fatalf("daily read: %v", err)
		}
	})

	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestCmdDaily_SpecificDate(t *testing.T) {
	vaultDir := t.TempDir()

	params := map[string]string{"date": "2025-06-15"}
	if err := cmdDaily(vaultDir, params); err != nil {
		t.Fatalf("daily specific date: %v", err)
	}

	notePath := filepath.Join(vaultDir, "2025-06-15.md")
	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("daily note not created: %v", err)
	}

	if !strings.Contains(string(data), "# 2025-06-15") {
		t.Errorf("daily note content: %q", string(data))
	}
}

func TestCmdDaily_WithTemplate(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, "_templates"), 0755)

	os.WriteFile(
		filepath.Join(vaultDir, ".obsidian", "daily-notes.json"),
		[]byte(`{"template":"_templates/daily"}`),
		0644,
	)

	os.WriteFile(
		filepath.Join(vaultDir, "_templates", "daily.md"),
		[]byte("---\ndate: {{date}}\n---\n\n# {{title}}\n\n## Tasks\n\n## Notes\n"),
		0644,
	)

	params := map[string]string{"date": "2025-03-20"}
	if err := cmdDaily(vaultDir, params); err != nil {
		t.Fatalf("daily with template: %v", err)
	}

	notePath := filepath.Join(vaultDir, "2025-03-20.md")
	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("daily note not created: %v", err)
	}

	got := string(data)
	if !strings.Contains(got, "date: 2025-03-20") {
		t.Errorf("template date not substituted: %q", got)
	}
	if !strings.Contains(got, "# 2025-03-20") {
		t.Errorf("template title not substituted: %q", got)
	}
	if !strings.Contains(got, "## Tasks") {
		t.Errorf("template structure not preserved: %q", got)
	}
}

func TestCmdDaily_WithFolder(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)
	os.WriteFile(
		filepath.Join(vaultDir, ".obsidian", "daily-notes.json"),
		[]byte(`{"folder":"journal"}`),
		0644,
	)

	params := map[string]string{"date": "2025-06-15"}
	if err := cmdDaily(vaultDir, params); err != nil {
		t.Fatalf("daily with folder: %v", err)
	}

	notePath := filepath.Join(vaultDir, "journal", "2025-06-15.md")
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		t.Errorf("daily note not created in folder")
	}
}

func TestCmdDaily_InvalidDate(t *testing.T) {
	vaultDir := t.TempDir()

	params := map[string]string{"date": "not-a-date"}
	if err := cmdDaily(vaultDir, params); err == nil {
		t.Fatal("expected error for invalid date")
	}
}
