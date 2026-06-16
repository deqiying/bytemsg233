package libinstall

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func NormalizeLanguage(language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "ts", "js", "javascript", "typescript":
		return "typescript"
	case "cs", "c#", "unity", "csharp":
		return "csharp"
	default:
		return strings.ToLower(strings.TrimSpace(language))
	}
}

func CopyLibrary(repoRoot, language, targetDir string) error {
	language = NormalizeLanguage(language)
	sourceDir := filepath.Join(repoRoot, "libs", language)
	info, err := os.Stat(sourceDir)
	if err != nil {
		return fmt.Errorf("library %q is not available at %s: %w", language, sourceDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("library path is not a directory: %s", sourceDir)
	}

	return copyDir(sourceDir, targetDir)
}

func copyDir(sourceDir, targetDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := entry.Name()
		if name == ".git" || name == "node_modules" || name == "build" || name == "dist" || name == "target" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(targetDir, rel)
		if entry.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		return copyFile(path, dest)
	})
}

func copyFile(source, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
