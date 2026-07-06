package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultRepo   = "neko233-com/bytemsg233"
	defaultBinary = "bytemsg233"
)

type Options struct {
	CurrentVersion string
	Version        string
	Repo           string
	Binary         string
	Mirrors        []string
	Timeout        time.Duration
	Force          bool
	DryRun         bool
	Out            io.Writer
}

type Result struct {
	Version string
	Source  string
	Path    string
}

type releaseInfo struct {
	TagName string `json:"tag_name"`
}

func Run(ctx context.Context, opts Options) (*Result, error) {
	if opts.Repo == "" {
		opts.Repo = defaultRepo
	}
	if opts.Binary == "" {
		opts.Binary = defaultBinary
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	mirrors := normalizeMirrors(opts.Mirrors)
	client := &http.Client{Timeout: opts.Timeout}

	targetVersion := opts.Version
	if targetVersion == "" || targetVersion == "latest" {
		version, source, err := latestVersion(ctx, client, opts.Repo, mirrors)
		if err != nil {
			return nil, err
		}
		targetVersion = version
		fmt.Fprintf(opts.Out, "Latest version: %s (%s)\n", targetVersion, source)
	}
	if !strings.HasPrefix(targetVersion, "v") {
		targetVersion = "v" + targetVersion
	}

	if !opts.Force && opts.CurrentVersion != "" && opts.CurrentVersion != "dev" && opts.CurrentVersion == targetVersion {
		fmt.Fprintf(opts.Out, "%s already latest.\n", opts.Binary)
		return &Result{Version: targetVersion}, nil
	}

	assetName, err := assetName(opts.Binary)
	if err != nil {
		return nil, err
	}
	baseURL := "https://github.com/" + opts.Repo + "/releases/download/" + targetVersion + "/"
	assetURL := baseURL + assetName
	checksumURL := baseURL + "checksums.txt"

	if opts.DryRun {
		fmt.Fprintln(opts.Out, "Would try download URLs:")
		for _, candidate := range mirroredURLs(assetURL, mirrors) {
			fmt.Fprintf(opts.Out, "- %s\n", candidate)
		}
		return &Result{Version: targetVersion, Source: assetURL}, nil
	}

	tmpDir, err := os.MkdirTemp("", "bytemsg233-update-*")
	if err != nil {
		return nil, err
	}
	if runtime.GOOS != "windows" {
		defer os.RemoveAll(tmpDir)
	}

	archivePath := filepath.Join(tmpDir, assetName)
	checksumPath := filepath.Join(tmpDir, "checksums.txt")
	archiveSource, err := downloadWithFallback(ctx, client, assetURL, archivePath, mirrors, opts.Out)
	if err != nil {
		return nil, err
	}
	if checksumSource, err := downloadWithFallback(ctx, client, checksumURL, checksumPath, mirrors, opts.Out); err == nil {
		if err := verifyChecksum(checksumPath, assetName, archivePath); err != nil {
			return nil, err
		}
		fmt.Fprintf(opts.Out, "Checksum OK (%s)\n", checksumSource)
	} else {
		return nil, fmt.Errorf("download checksums.txt failed: %w", err)
	}

	newBinary := filepath.Join(tmpDir, exeName(opts.Binary))
	if err := extractBinary(archivePath, opts.Binary, newBinary); err != nil {
		return nil, err
	}

	currentPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		return nil, err
	}
	if err := installBinary(newBinary, currentPath); err != nil {
		return nil, err
	}
	fmt.Fprintf(opts.Out, "%s updated to %s\n", opts.Binary, targetVersion)
	return &Result{Version: targetVersion, Source: archiveSource, Path: currentPath}, nil
}

func normalizeMirrors(mirrors []string) []string {
	env := os.Getenv("BYTEMSG233_UPDATE_MIRROR")
	if env != "" {
		mirrors = append(strings.Split(env, ","), mirrors...)
	}
	if len(mirrors) == 0 {
		mirrors = []string{
			"https://gh-proxy.com/",
			"https://ghfast.top/",
			"https://gh.llkk.cc/",
		}
	}
	out := make([]string, 0, len(mirrors))
	for _, mirror := range mirrors {
		mirror = strings.TrimSpace(mirror)
		if mirror == "" {
			continue
		}
		out = append(out, mirror)
	}
	return out
}

func latestVersion(ctx context.Context, client *http.Client, repo string, mirrors []string) (string, string, error) {
	apiURL := "https://api.github.com/repos/" + repo + "/releases/latest"
	for _, candidate := range mirroredURLs(apiURL, mirrors) {
		body, finalURL, err := fetch(ctx, client, candidate)
		if err != nil {
			continue
		}
		var release releaseInfo
		if json.Unmarshal(body, &release) == nil && release.TagName != "" {
			return release.TagName, candidate, nil
		}
		if tag := tagFromLatestURL(finalURL); tag != "" {
			return tag, candidate, nil
		}
	}

	latestURL := "https://github.com/" + repo + "/releases/latest"
	for _, candidate := range mirroredURLs(latestURL, mirrors) {
		_, finalURL, err := fetch(ctx, client, candidate)
		if err != nil {
			continue
		}
		if tag := tagFromLatestURL(finalURL); tag != "" {
			return tag, candidate, nil
		}
	}
	return "", "", errors.New("latest version lookup failed; use --version vX.Y.Z or set BYTEMSG233_UPDATE_MIRROR")
}

func mirroredURLs(raw string, mirrors []string) []string {
	urls := []string{raw}
	for _, mirror := range mirrors {
		if strings.Contains(mirror, "{url}") {
			urls = append(urls, strings.ReplaceAll(mirror, "{url}", raw))
			continue
		}
		urls = append(urls, strings.TrimRight(mirror, "/")+"/"+raw)
	}
	return urls
}

func fetch(ctx context.Context, client *http.Client, rawURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "bytemsg233-updater")
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.Request.URL.String(), fmt.Errorf("GET %s: %s", rawURL, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Request.URL.String(), err
	}
	return body, resp.Request.URL.String(), nil
}

func tagFromLatestURL(raw string) string {
	i := strings.LastIndex(raw, "/releases/tag/")
	if i < 0 {
		return ""
	}
	tag := raw[i+len("/releases/tag/"):]
	if j := strings.IndexAny(tag, "?#"); j >= 0 {
		tag = tag[:j]
	}
	return tag
}

func assetName(binary string) (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	switch osName {
	case "linux", "darwin":
	case "windows":
	default:
		return "", fmt.Errorf("unsupported OS %q", osName)
	}
	switch arch {
	case "amd64", "arm64":
	default:
		return "", fmt.Errorf("unsupported architecture %q", arch)
	}
	if osName == "windows" {
		return binary + "_" + osName + "_" + arch + ".zip", nil
	}
	return binary + "_" + osName + "_" + arch + ".tar.gz", nil
}

func exeName(binary string) string {
	if runtime.GOOS == "windows" {
		return binary + ".exe"
	}
	return binary
}

func downloadWithFallback(ctx context.Context, client *http.Client, rawURL, dst string, mirrors []string, out io.Writer) (string, error) {
	var lastErr error
	for _, candidate := range mirroredURLs(rawURL, mirrors) {
		fmt.Fprintf(out, "Downloading %s\n", candidate)
		err := download(ctx, client, candidate, dst)
		if err == nil {
			return candidate, nil
		}
		lastErr = err
		fmt.Fprintf(out, "Download failed: %v\n", err)
	}
	return "", lastErr
}

func download(ctx context.Context, client *http.Client, rawURL, dst string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "bytemsg233-updater")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GET %s: %s", rawURL, resp.Status)
	}
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func verifyChecksum(checksumPath, assetName, archivePath string) error {
	data, err := os.ReadFile(checksumPath)
	if err != nil {
		return err
	}
	want := ""
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == assetName {
			want = fields[0]
			break
		}
	}
	if want == "" {
		return fmt.Errorf("checksum for %s not found", assetName)
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	sum := sha256.New()
	if _, err := io.Copy(sum, file); err != nil {
		return err
	}
	got := hex.EncodeToString(sum.Sum(nil))
	if !strings.EqualFold(want, got) {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	return nil
}

func extractBinary(archivePath, binary, dst string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, binary, dst)
	}
	return extractTarGz(archivePath, binary, dst)
}

func extractZip(archivePath, binary, dst string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	want := exeName(binary)
	for _, file := range reader.File {
		if filepath.Base(file.Name) != want {
			continue
		}
		in, err := file.Open()
		if err != nil {
			return err
		}
		defer in.Close()
		return writeExecutable(dst, in)
	}
	return fmt.Errorf("%s not found in archive", want)
}

func extractTarGz(archivePath, binary, dst string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	want := exeName(binary)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.Typeflag == tar.TypeReg && filepath.Base(header.Name) == want {
			return writeExecutable(dst, tr)
		}
	}
	return fmt.Errorf("%s not found in archive", want)
}

func writeExecutable(dst string, src io.Reader) error {
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func installBinary(src, dst string) error {
	if runtime.GOOS == "windows" {
		return installBinaryWindows(src, dst)
	}
	info, err := os.Stat(dst)
	if err != nil {
		return err
	}
	tmp := dst + ".new"
	if err := copyFile(src, tmp, info.Mode()); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func installBinaryWindows(src, dst string) error {
	tmp := dst + ".new"
	if err := copyFile(src, tmp, 0755); err != nil {
		return err
	}
	pid := os.Getpid()
	script := fmt.Sprintf(`$pidToWait = %d
$src = %s
$dst = %s
$old = "$dst.old"
Wait-Process -Id $pidToWait -ErrorAction SilentlyContinue
Start-Sleep -Milliseconds 200
Remove-Item -LiteralPath $old -Force -ErrorAction SilentlyContinue
Move-Item -LiteralPath $dst -Destination $old -Force
Move-Item -LiteralPath $src -Destination $dst -Force
Remove-Item -LiteralPath $old -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath (Split-Path -Parent $src) -Recurse -Force -ErrorAction SilentlyContinue
`, pid, psQuote(tmp), psQuote(dst))
	scriptPath := filepath.Join(os.TempDir(), fmt.Sprintf("bytemsg233-update-%d.ps1", pid))
	if err := os.WriteFile(scriptPath, []byte(script), 0600); err != nil {
		return err
	}
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	return cmd.Start()
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func psQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
