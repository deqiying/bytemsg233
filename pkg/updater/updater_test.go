package updater

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMirroredURLs(t *testing.T) {
	got := mirroredURLs("https://github.com/neko233-com/bytemsg233/releases/latest", []string{
		"https://gh-proxy.com/",
		"https://mirror.example/{url}",
	})
	want := []string{
		"https://github.com/neko233-com/bytemsg233/releases/latest",
		"https://gh-proxy.com/https://github.com/neko233-com/bytemsg233/releases/latest",
		"https://mirror.example/https://github.com/neko233-com/bytemsg233/releases/latest",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d URLs, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("URL %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestTagFromLatestURL(t *testing.T) {
	got := tagFromLatestURL("https://github.com/neko233-com/bytemsg233/releases/tag/v1.2.3?x=1")
	if got != "v1.2.3" {
		t.Fatalf("tag = %q", got)
	}
}

func TestAssetName(t *testing.T) {
	name, err := assetName("bytemsg233")
	if err != nil {
		t.Fatal(err)
	}
	wantExt := ".tar.gz"
	if runtime.GOOS == "windows" {
		wantExt = ".zip"
	}
	if !hasSuffix(name, "_"+runtime.GOOS+"_"+runtime.GOARCH+wantExt) {
		t.Fatalf("asset name = %q, want current OS/arch suffix", name)
	}
}

func TestVerifyChecksum(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "bytemsg233_linux_amd64.tar.gz")
	if err := os.WriteFile(archive, []byte("archive"), 0644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte("archive"))
	checksums := filepath.Join(dir, "checksums.txt")
	content := fmt.Sprintf("%x  bytemsg233_linux_amd64.tar.gz\n", sum)
	if err := os.WriteFile(checksums, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := verifyChecksum(checksums, "bytemsg233_linux_amd64.tar.gz", archive); err != nil {
		t.Fatal(err)
	}
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
