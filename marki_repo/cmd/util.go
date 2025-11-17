package cmd

import (
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func mkdirIfNotExists(outDir string) error {
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		return err
	}
	return nil
}

func dirClear(dirName string) error {
	err := filepath.WalkDir(dirName, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		err = os.Remove(path)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(result)
}
func dirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func isValidTheme(theme string) error {
	validThemes := []string{
		"abap", "algol", "algol_nu", "arduino", "autumn", "average", "base16-snazzy",
		"borland", "bw", "catppuccin-frappe", "catppuccin-latte", "catppuccin-macchiato",
		"catppuccin-mocha", "colorful", "doom-one", "doom-one2", "dracula", "emacs",
		"evergarden", "friendly", "fruity", "github-dark", "github", "gruvbox-light",
		"gruvbox", "hr_high_contrast", "hrdark", "igor", "lovelace", "manni", "modus-operandi",
		"modus-vivendi", "monokai", "monokailight", "murphy", "native", "nord", "nordic",
		"onedark", "onesenterprise", "paraiso-dark", "paraiso-light", "pastie", "perldoc",
		"pygments", "rainbow_dash", "rose-pine-dawn", "rose-pine-moon", "rose-pine", "rpgle",
		"rrt", "solarized-dark", "solarized-dark256", "solarized-light", "swapoff", "tango",
		"tokyonight-day", "tokyonight-moon", "tokyonight-night", "tokyonight-storm", "trac",
		"vim", "vs", "vulcan", "witchhazel", "xcode-dark", "xcode",
	}
	for _, vTheme := range validThemes {
		if theme == vTheme {
			return nil
		}
	}
	return fmt.Errorf("<THEME> [%s] is not a valid theme\nfor a list of valid themes see https://github.com/phillip-england/marki", theme)
}
