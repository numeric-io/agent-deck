package session

import (
	"os"
	"path/filepath"
	"strings"
)

// notesDirName is the subdirectory within a profile for group notes
const notesDirName = "notes"

// LoadGroupNotes reads the notes file for a group. Returns "" if the file is missing.
func LoadGroupNotes(profile, groupPath string) string {
	path, err := notesFilePath(profile, groupPath)
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// SaveGroupNotes writes notes for a group. If content is empty the file is removed.
func SaveGroupNotes(profile, groupPath, content string) error {
	path, err := notesFilePath(profile, groupPath)
	if err != nil {
		return err
	}

	if strings.TrimSpace(content) == "" {
		// Remove the file; ignore "not exists" errors.
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

// NotesPreview returns the first non-empty line of the notes for list display.
func NotesPreview(notes string) string {
	for _, line := range strings.Split(notes, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// notesFilePath returns the filesystem path for a group's notes file.
// Slashes in group paths are replaced with "__" to keep a flat directory.
func notesFilePath(profile, groupPath string) (string, error) {
	profileDir, err := GetProfileDir(profile)
	if err != nil {
		return "", err
	}
	// Replace "/" with "__" so "projects/devops" → "projects__devops.md"
	safe := strings.ReplaceAll(groupPath, "/", "__")
	return filepath.Join(profileDir, notesDirName, safe+".md"), nil
}
