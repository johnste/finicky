package browser

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// desktopEntry is the subset of a freedesktop.org .desktop file that Finicky
// needs to list and launch browsers.
// See https://specifications.freedesktop.org/desktop-entry-spec/latest/
type desktopEntry struct {
	ID       string // desktop file ID, e.g. "firefox.desktop"
	Name     string
	Exec     string
	MimeType []string
	Hidden   bool
	NoShow   bool
}

func (e desktopEntry) handlesHTTPS() bool {
	for _, m := range e.MimeType {
		if m == "x-scheme-handler/https" {
			return true
		}
	}
	return false
}

// isFinicky reports whether the entry launches Finicky itself, which must
// never be listed or launched as a browser.
func (e desktopEntry) isFinicky() bool {
	if strings.Contains(strings.ToLower(e.ID), "finicky") {
		return true
	}
	argv, err := parseExec(e.Exec, "")
	return err == nil && len(argv) > 0 && strings.EqualFold(filepath.Base(argv[0]), "finicky")
}

// applicationDirs returns the XDG applications directories in precedence
// order: $XDG_DATA_HOME first, then each of $XDG_DATA_DIRS.
func applicationDirs() []string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			dataHome = filepath.Join(home, ".local", "share")
		}
	}
	dataDirs := os.Getenv("XDG_DATA_DIRS")
	if dataDirs == "" {
		dataDirs = "/usr/local/share:/usr/share"
	}

	var dirs []string
	if dataHome != "" {
		dirs = append(dirs, filepath.Join(dataHome, "applications"))
	}
	for _, d := range strings.Split(dataDirs, ":") {
		if d != "" {
			dirs = append(dirs, filepath.Join(d, "applications"))
		}
	}
	return dirs
}

// loadDesktopEntries reads every application entry in dirs. When the same
// desktop file ID appears in several directories, the first one wins, as the
// spec requires, so a user's override in ~/.local/share hides the system one.
func loadDesktopEntries(dirs []string) []desktopEntry {
	seen := map[string]bool{}
	var entries []desktopEntry
	for _, dir := range dirs {
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".desktop") {
				return nil
			}
			rel, _ := filepath.Rel(dir, path)
			id := strings.ReplaceAll(rel, string(filepath.Separator), "-")
			if seen[id] {
				return nil
			}
			seen[id] = true
			entry, ok := readDesktopEntry(path)
			if !ok {
				return nil
			}
			entry.ID = id
			entries = append(entries, entry)
			return nil
		})
	}
	return entries
}

func readDesktopEntry(path string) (desktopEntry, bool) {
	f, err := os.Open(path)
	if err != nil {
		return desktopEntry{}, false
	}
	defer f.Close()

	var entry desktopEntry
	isApp := false
	inMain := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inMain = line == "[Desktop Entry]"
			continue
		}
		if !inMain {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), unescapeDesktopValue(strings.TrimSpace(value))
		switch key {
		case "Type":
			isApp = value == "Application"
		case "Name":
			entry.Name = value
		case "Exec":
			entry.Exec = value
		case "MimeType":
			for _, m := range strings.Split(value, ";") {
				if m != "" {
					entry.MimeType = append(entry.MimeType, m)
				}
			}
		case "Hidden":
			entry.Hidden = value == "true"
		case "NoDisplay":
			entry.NoShow = value == "true"
		}
	}
	return entry, isApp && !entry.Hidden && entry.Exec != ""
}

// unescapeDesktopValue applies the escapes allowed in desktop file string
// values (\s, \n, \t, \r, \\).
func unescapeDesktopValue(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 == len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 's':
			b.WriteByte(' ')
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// parseExec splits a desktop entry Exec value into argv. File and URL field
// codes (%u, %U, %f, %F) are replaced by url, or dropped when url is empty;
// %% becomes a literal %, and other field codes are dropped.
func parseExec(exec string, url string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inArg, inQuote := false, false

	flush := func() {
		if inArg {
			args = append(args, cur.String())
			cur.Reset()
			inArg = false
		}
	}

	for i := 0; i < len(exec); i++ {
		c := exec[i]
		switch {
		case inQuote && c == '"':
			inQuote = false
		case inQuote && c == '\\' && i+1 < len(exec) && strings.IndexByte("\"`$\\", exec[i+1]) >= 0:
			i++
			cur.WriteByte(exec[i])
		case inQuote:
			cur.WriteByte(c)
		case c == '"':
			inQuote, inArg = true, true
		case c == ' ' || c == '\t':
			flush()
		case c == '%' && i+1 < len(exec):
			i++
			switch exec[i] {
			case '%':
				cur.WriteByte('%')
				inArg = true
			case 'u', 'U', 'f', 'F':
				if url != "" {
					cur.WriteString(url)
					inArg = true
				}
			}
		default:
			cur.WriteByte(c)
			inArg = true
		}
	}
	if inQuote {
		return nil, fmt.Errorf("unterminated quote in Exec %q", exec)
	}
	flush()
	if len(args) == 0 {
		return nil, fmt.Errorf("empty Exec")
	}
	return args, nil
}

// findDesktopEntry returns the entry matching identifier, which may be the
// entry's Name ("Firefox"), its desktop file ID with or without the .desktop
// suffix ("firefox", "org.mozilla.firefox.desktop"), or the file name of its
// executable ("firefox").
func findDesktopEntry(entries []desktopEntry, identifier string) (desktopEntry, bool) {
	id := strings.TrimSuffix(identifier, ".desktop") + ".desktop"
	matchers := []func(desktopEntry) bool{
		func(e desktopEntry) bool { return e.ID == id },
		func(e desktopEntry) bool { return e.Name == identifier },
		func(e desktopEntry) bool { return strings.EqualFold(e.ID, id) || strings.EqualFold(e.Name, identifier) },
		func(e desktopEntry) bool {
			argv, err := parseExec(e.Exec, "")
			return err == nil && strings.EqualFold(filepath.Base(argv[0]), identifier)
		},
	}
	for _, match := range matchers {
		for _, e := range entries {
			if !e.isFinicky() && match(e) {
				return e, true
			}
		}
	}
	return desktopEntry{}, false
}
