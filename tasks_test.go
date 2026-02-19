package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTasks(t *testing.T) {
	text := `# My Note

- [ ] Buy groceries
- [x] Review PR
- [X] Deploy changes
  - [ ] Nested task
Some random text
- [ ] Another task
`
	tasks := parseTasks(text)

	if len(tasks) != 5 {
		t.Fatalf("got %d tasks, want 5", len(tasks))
	}

	// First task
	if tasks[0].Text != "Buy groceries" || tasks[0].Done || tasks[0].Line != 3 {
		t.Errorf("task[0] = %+v, want Buy groceries, done=false, line=3", tasks[0])
	}

	// Second task (done)
	if tasks[1].Text != "Review PR" || !tasks[1].Done || tasks[1].Line != 4 {
		t.Errorf("task[1] = %+v, want Review PR, done=true, line=4", tasks[1])
	}

	// Third task (X uppercase)
	if !tasks[2].Done {
		t.Errorf("task[2] should be done (uppercase X)")
	}

	// Fourth task (nested)
	if tasks[3].Text != "Nested task" || tasks[3].Done {
		t.Errorf("task[3] = %+v, want Nested task, done=false", tasks[3])
	}
}

func TestParseTasks_Empty(t *testing.T) {
	tasks := parseTasks("# No tasks here\n\nJust text.\n")
	if len(tasks) != 0 {
		t.Errorf("got %d tasks, want 0", len(tasks))
	}
}

func TestFilterTasks(t *testing.T) {
	tasks := []task{
		{Text: "Done task", Done: true},
		{Text: "Pending task", Done: false},
		{Text: "Another done", Done: true},
	}

	done := filterTasks(tasks, true, false)
	if len(done) != 2 {
		t.Errorf("done filter: got %d, want 2", len(done))
	}

	pending := filterTasks(tasks, false, true)
	if len(pending) != 1 {
		t.Errorf("pending filter: got %d, want 1", len(pending))
	}

	all := filterTasks(tasks, false, false)
	if len(all) != 3 {
		t.Errorf("no filter: got %d, want 3", len(all))
	}
}

func TestCmdTasks_SingleFile(t *testing.T) {
	vaultDir := t.TempDir()

	os.WriteFile(
		filepath.Join(vaultDir, "Tasks.md"),
		[]byte("# Tasks\n\n- [ ] Do thing 1\n- [x] Done thing\n- [ ] Do thing 2\n"),
		0644,
	)

	params := map[string]string{"file": "Tasks"}
	flags := map[string]bool{}
	if err := cmdTasks(vaultDir, params, flags); err != nil {
		t.Fatalf("tasks single file: %v", err)
	}
}

func TestCmdTasks_VaultWide(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "projects"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)

	os.WriteFile(
		filepath.Join(vaultDir, "Daily.md"),
		[]byte("- [ ] Buy groceries\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, "projects", "Plan.md"),
		[]byte("- [x] Review PR\n- [ ] Deploy\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, ".obsidian", "hidden.md"),
		[]byte("- [ ] Should be skipped\n"),
		0644,
	)

	params := map[string]string{}
	flags := map[string]bool{}
	if err := cmdTasks(vaultDir, params, flags); err != nil {
		t.Fatalf("tasks vault-wide: %v", err)
	}
}

func TestCmdTasks_FilterDone(t *testing.T) {
	vaultDir := t.TempDir()

	os.WriteFile(
		filepath.Join(vaultDir, "Tasks.md"),
		[]byte("- [ ] Pending\n- [x] Done\n"),
		0644,
	)

	params := map[string]string{"file": "Tasks"}
	flags := map[string]bool{"done": true}
	if err := cmdTasks(vaultDir, params, flags); err != nil {
		t.Fatalf("tasks filter done: %v", err)
	}
}

func TestCmdTasks_FilterPending(t *testing.T) {
	vaultDir := t.TempDir()

	os.WriteFile(
		filepath.Join(vaultDir, "Tasks.md"),
		[]byte("- [ ] Pending\n- [x] Done\n"),
		0644,
	)

	params := map[string]string{"file": "Tasks"}
	flags := map[string]bool{"pending": true}
	if err := cmdTasks(vaultDir, params, flags); err != nil {
		t.Fatalf("tasks filter pending: %v", err)
	}
}

func TestCmdTasks_PathFilter(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "projects"), 0755)

	os.WriteFile(
		filepath.Join(vaultDir, "Root.md"),
		[]byte("- [ ] Root task\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, "projects", "Project.md"),
		[]byte("- [ ] Project task\n"),
		0644,
	)

	params := map[string]string{"path": "projects"}
	flags := map[string]bool{}
	if err := cmdTasks(vaultDir, params, flags); err != nil {
		t.Fatalf("tasks path filter: %v", err)
	}
}

func TestCmdTasks_JSONOutput(t *testing.T) {
	vaultDir := t.TempDir()

	os.WriteFile(
		filepath.Join(vaultDir, "Tasks.md"),
		[]byte("- [ ] Buy groceries\n- [x] Review PR\n"),
		0644,
	)

	params := map[string]string{"file": "Tasks"}
	flags := map[string]bool{"--json": true}

	got := captureStdout(func() {
		if err := cmdTasks(vaultDir, params, flags); err != nil {
			t.Fatalf("tasks json: %v", err)
		}
	})

	if got == "" {
		t.Fatal("expected json output, got empty")
	}
	if got[0] != '[' {
		t.Errorf("expected json array, got: %q", got[:20])
	}
}
