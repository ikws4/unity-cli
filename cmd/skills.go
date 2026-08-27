package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

var embeddedSkillContent fs.FS

// SetEmbeddedSkillContent registers the skill tree compiled into the binary.
func SetEmbeddedSkillContent(content fs.FS) {
	embeddedSkillContent = content
}

type skillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

type skillEntry struct {
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
}

func skillsCmd(args []string) error {
	return runSkillsCmd(embeddedSkillContent, args, os.Stdout)
}

func runSkillsCmd(content fs.FS, args []string, out io.Writer) error {
	if content == nil {
		return fmt.Errorf("skill content is not embedded in this build")
	}
	if len(args) == 0 {
		return fmt.Errorf("skills requires a subcommand: list or read")
	}

	switch args[0] {
	case "list":
		return listSkills(content, args[1:], out)
	case "read":
		return readSkill(content, args[1:], out)
	default:
		return fmt.Errorf("unknown skills subcommand %q; expected list or read", args[0])
	}
}

func listSkills(content fs.FS, args []string, out io.Writer) error {
	if len(args) > 1 {
		return fmt.Errorf("skills list accepts at most one skill path")
	}
	if len(args) == 1 {
		entries, listedPath, err := listSkillPath(content, args[0])
		if err != nil {
			return err
		}
		return writeJSON(out, struct {
			Path    string       `json:"path"`
			Entries []skillEntry `json:"entries"`
			Count   int          `json:"count"`
		}{listedPath, entries, len(entries)})
	}

	entries, err := fs.ReadDir(content, ".")
	if err != nil {
		return fmt.Errorf("read embedded skills: %w", err)
	}

	infos := make([]skillInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, err := fs.ReadFile(content, entry.Name()+"/SKILL.md")
		if err != nil {
			continue
		}
		description, version := parseSkillFrontmatter(data)
		infos = append(infos, skillInfo{
			Name:        entry.Name(),
			Description: description,
			Version:     version,
		})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })

	return writeJSON(out, struct {
		Skills []skillInfo `json:"skills"`
		Count  int         `json:"count"`
	}{infos, len(infos)})
}

func listSkillPath(content fs.FS, target string) ([]skillEntry, string, error) {
	cleaned, err := cleanSkillPath(target)
	if err != nil {
		return nil, "", err
	}
	if err := ensureSkill(content, strings.Split(cleaned, "/")[0]); err != nil {
		return nil, "", err
	}
	info, err := fs.Stat(content, cleaned)
	if err != nil {
		return nil, "", fmt.Errorf("skill path %q not found", target)
	}
	if !info.IsDir() {
		return nil, "", fmt.Errorf("skill path %q is a file; use skills read", target)
	}

	dirEntries, err := fs.ReadDir(content, cleaned)
	if err != nil {
		return nil, "", fmt.Errorf("read skill path %q: %w", target, err)
	}
	entries := make([]skillEntry, 0, len(dirEntries))
	for _, entry := range dirEntries {
		entries = append(entries, skillEntry{
			Path:  cleaned + "/" + entry.Name(),
			IsDir: entry.IsDir(),
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, cleaned, nil
}

func readSkill(content fs.FS, args []string, out io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("skills read requires one skill name or skill path")
	}
	cleaned, err := cleanSkillPath(args[0])
	if err != nil {
		return err
	}
	name := strings.Split(cleaned, "/")[0]
	if err := ensureSkill(content, name); err != nil {
		return err
	}
	if cleaned == name {
		cleaned += "/SKILL.md"
	}

	info, err := fs.Stat(content, cleaned)
	if err != nil {
		return fmt.Errorf("skill file %q not found", args[0])
	}
	if info.IsDir() {
		return fmt.Errorf("skill path %q is a directory; use skills list", args[0])
	}
	data, err := fs.ReadFile(content, cleaned)
	if err != nil {
		return fmt.Errorf("read skill file %q: %w", args[0], err)
	}
	_, err = out.Write(data)
	return err
}

func ensureSkill(content fs.FS, name string) error {
	if name == "" || strings.ContainsAny(name, `/\`) || name == "." || name == ".." {
		return fmt.Errorf("unknown skill %q", name)
	}
	info, err := fs.Stat(content, name)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("unknown skill %q; run skills list", name)
	}
	if _, err := fs.Stat(content, name+"/SKILL.md"); err != nil {
		return fmt.Errorf("unknown skill %q; run skills list", name)
	}
	return nil
}

func cleanSkillPath(target string) (string, error) {
	if target == "" || strings.Contains(target, `\`) || path.IsAbs(target) {
		return "", fmt.Errorf("invalid skill path %q", target)
	}
	cleaned := path.Clean(target)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("invalid skill path %q: path traversal is not allowed", target)
	}
	return cleaned, nil
}

func parseSkillFrontmatter(data []byte) (description, version string) {
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", ""
	}
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
		switch key {
		case "description":
			description = value
		case "version":
			version = value
		}
	}
	return description, version
}

func writeJSON(out io.Writer, value interface{}) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
