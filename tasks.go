package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// task represents a parsed checkbox item from a note.
type task struct {
	Text string `json:"text"` // task text after the checkbox
	Done bool   `json:"done"` // true if [x] or [X]
	Line int    `json:"line"` // 1-based line number
	File string `json:"file"` // relative path (when searching vault-wide)
}

// taskPattern matches markdown checkboxes: - [ ] text or - [x] text
// Allows leading whitespace/tabs for nesting.
var taskPattern = regexp.MustCompile(`(?m)^[\t ]*- \[([ xX])\] (.+)$`)

// parseTasks extracts all checkbox items from text.
func parseTasks(text string) []task {
	lines := strings.Split(text, "\n")
	var tasks []task

	for i, line := range lines {
		m := taskPattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		tasks = append(tasks, task{
			Text: m[2],
			Done: m[1] == "x" || m[1] == "X",
			Line: i + 1,
		})
	}
	return tasks
}

// cmdTasks lists tasks (checkboxes) from one note or across the vault.
// Supports filters: done (only completed), pending (only incomplete).
// Supports path= to limit search to a subfolder.
func cmdTasks(vaultDir string, params map[string]string, flags map[string]bool) error {
	format := outputFormat(flags)
	filterDone := flags["done"]
	filterPending := flags["pending"]

	title := params["file"]
	pathFilter := params["path"]

	// Single file mode
	if title != "" {
		path, err := resolveNote(vaultDir, title)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(vaultDir, path)
		tasks := parseTasks(string(data))
		tasks = filterTasks(tasks, filterDone, filterPending)

		for i := range tasks {
			tasks[i].File = relPath
		}

		outputTasks(tasks, format)
		return nil
	}

	// Vault-wide mode
	searchRoot := vaultDir
	if pathFilter != "" {
		searchRoot = filepath.Join(vaultDir, pathFilter)
		if _, err := os.Stat(searchRoot); os.IsNotExist(err) {
			return fmt.Errorf("path filter %q not found in vault", pathFilter)
		}
	}

	var allTasks []task

	err := filepath.WalkDir(searchRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		name := d.Name()
		if d.IsDir() && (strings.HasPrefix(name, ".") || name == ".trash") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(name, ".md") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(vaultDir, path)
		tasks := parseTasks(string(data))

		for i := range tasks {
			tasks[i].File = relPath
		}

		allTasks = append(allTasks, tasks...)
		return nil
	})

	if err != nil {
		return err
	}

	allTasks = filterTasks(allTasks, filterDone, filterPending)
	outputTasks(allTasks, format)
	return nil
}

// filterTasks applies done/pending filters.
func filterTasks(tasks []task, done, pending bool) []task {
	if !done && !pending {
		return tasks
	}

	var result []task
	for _, t := range tasks {
		if done && t.Done {
			result = append(result, t)
		}
		if pending && !t.Done {
			result = append(result, t)
		}
	}
	return result
}

// outputTasks prints tasks in the requested format.
func outputTasks(tasks []task, format string) {
	switch format {
	case "json":
		data, _ := json.Marshal(tasks)
		fmt.Println(string(data))
	case "csv":
		fmt.Println("done,text,line,file")
		for _, t := range tasks {
			done := "false"
			if t.Done {
				done = "true"
			}
			fmt.Printf("%s,%q,%d,%s\n", done, t.Text, t.Line, t.File)
		}
	case "yaml":
		for _, t := range tasks {
			fmt.Printf("- text: %s\n  done: %v\n  line: %d\n  file: %s\n", yamlEscapeValue(t.Text), t.Done, t.Line, t.File)
		}
	default:
		for _, t := range tasks {
			check := " "
			if t.Done {
				check = "x"
			}
			fmt.Printf("- [%s] %s (%s:%d)\n", check, t.Text, t.File, t.Line)
		}
	}
}
