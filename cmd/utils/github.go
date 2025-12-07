package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	// GitHub API endpoint for latest release
	githubReleasesURL = "https://api.github.com/repos/lukaszraczylo/semver-generator/releases/latest"
	// Request timeout for HTTP requests
	requestTimeout = 10 * time.Second
)

// ReleaseInfo contains information about a GitHub release
type ReleaseInfo struct {
	TagName string         `json:"tag_name"`
	HTMLURL string         `json:"html_url"`
	Name    string         `json:"name"`
	Assets  []ReleaseAsset `json:"assets"`
}

// ReleaseAsset contains information about a release asset
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	DownloadURL    string
}

// httpClient is the HTTP client used for requests (allows mocking in tests)
var httpClient = &http.Client{
	Timeout: requestTimeout,
}

// CheckLatestRelease checks for the latest release version using REST API
// Returns the latest version tag and true if successful, empty string and false otherwise
func CheckLatestRelease() (string, bool) {
	release, err := fetchLatestRelease(context.Background())
	if err != nil {
		Debug("Unable to check latest release", map[string]interface{}{"error": err.Error()})
		return "", false
	}

	version := normalizeVersion(release.TagName)
	return version, true
}

// CheckForUpdate checks if a newer version is available
// Returns UpdateInfo if an update is available, nil otherwise
func CheckForUpdate(currentVersion string) *UpdateInfo {
	release, err := fetchLatestRelease(context.Background())
	if err != nil {
		return nil
	}

	latestVersion := normalizeVersion(release.TagName)
	current := normalizeVersion(currentVersion)

	if isNewerVersion(latestVersion, current) {
		downloadURL := findBinaryAsset(release.Assets)
		return &UpdateInfo{
			CurrentVersion: current,
			LatestVersion:  latestVersion,
			ReleaseURL:     release.HTMLURL,
			DownloadURL:    downloadURL,
		}
	}

	return nil
}

// UpdatePackage downloads and installs the latest version
func UpdatePackage() bool {
	Info("Checking for updates", nil)

	release, err := fetchLatestRelease(context.Background())
	if err != nil {
		Error("Unable to fetch latest release", map[string]interface{}{"error": err.Error()})
		return false
	}

	downloadURL := findBinaryAsset(release.Assets)
	if downloadURL == "" {
		Error("Unable to find binary for current platform", map[string]interface{}{
			"os":   runtime.GOOS,
			"arch": runtime.GOARCH,
		})
		return false
	}

	Info("Downloading update", map[string]interface{}{
		"version": release.TagName,
		"url":     downloadURL,
	})

	// Download to temp file
	tempFile, err := downloadBinary(downloadURL)
	if err != nil {
		Error("Unable to download binary", map[string]interface{}{"error": err.Error()})
		return false
	}
	defer os.Remove(tempFile) // Clean up temp file on failure

	// Get current binary path
	currentBinary, err := os.Executable()
	if err != nil {
		Error("Unable to get current binary path", map[string]interface{}{"error": err.Error()})
		return false
	}

	// Replace current binary
	if err := replaceBinary(tempFile, currentBinary); err != nil {
		Error("Unable to replace binary", map[string]interface{}{"error": err.Error()})
		return false
	}

	Info("Update successful", map[string]interface{}{
		"version": release.TagName,
	})
	return true
}

// fetchLatestRelease fetches the latest release info from GitHub REST API
func fetchLatestRelease(ctx context.Context) (*ReleaseInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubReleasesURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "semver-generator")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// findBinaryAsset finds the download URL for the current platform
func findBinaryAsset(assets []ReleaseAsset) string {
	// Build expected binary name pattern
	// Format: semver-gen-{version}-{os}-{arch}.tar.gz or just semver-gen-{os}-{arch}
	osName := runtime.GOOS
	archName := runtime.GOARCH

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		// Match patterns like "semver-gen-1.0.0-darwin-arm64.tar.gz" or "semver-gen-darwin-arm64"
		if strings.Contains(name, osName) && strings.Contains(name, archName) {
			// Prefer tar.gz archives
			if strings.HasSuffix(name, ".tar.gz") {
				return asset.BrowserDownloadURL
			}
		}
	}

	// Fallback: try to find any matching binary without tar.gz
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, osName) && strings.Contains(name, archName) {
			// Skip checksums
			if strings.Contains(name, "checksum") || strings.HasSuffix(name, ".sha256") || strings.HasSuffix(name, ".md5") {
				continue
			}
			return asset.BrowserDownloadURL
		}
	}

	return ""
}

// downloadBinary downloads the binary to a temp file and returns the path
func downloadBinary(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file
	tempFile, err := os.CreateTemp("", "semver-gen-update-*")
	if err != nil {
		return "", err
	}
	tempPath := tempFile.Name()

	// Check if it's a tar.gz archive
	if strings.HasSuffix(url, ".tar.gz") {
		// For tar.gz, we need to extract the binary
		if err := extractTarGz(resp.Body, tempFile); err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return "", err
		}
	} else {
		// Direct binary download
		if _, err := io.Copy(tempFile, resp.Body); err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return "", err
		}
	}

	tempFile.Close()
	return tempPath, nil
}

// extractTarGz extracts the semver-gen binary from a tar.gz archive
func extractTarGz(r io.Reader, destFile *os.File) error {
	// For simplicity, we'll download the whole archive to a temp file first,
	// then use tar command to extract. This avoids adding archive/tar dependency.

	// Create temp archive file
	archiveFile, err := os.CreateTemp("", "semver-gen-archive-*.tar.gz")
	if err != nil {
		return err
	}
	archivePath := archiveFile.Name()
	defer os.Remove(archivePath)

	if _, err := io.Copy(archiveFile, r); err != nil {
		archiveFile.Close()
		return err
	}
	archiveFile.Close()

	// Extract using tar command
	extractDir, err := os.MkdirTemp("", "semver-gen-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(extractDir)

	// Use tar to extract
	cmd := fmt.Sprintf("tar -xzf %s -C %s", archivePath, extractDir)
	if err := runCommand(cmd); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	// Find the semver-gen binary in the extracted files
	binaryPath := ""
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Name() == "semver-gen" || strings.HasPrefix(entry.Name(), "semver-gen") && !strings.Contains(entry.Name(), ".") {
			binaryPath = fmt.Sprintf("%s/%s", extractDir, entry.Name())
			break
		}
	}

	if binaryPath == "" {
		return fmt.Errorf("semver-gen binary not found in archive")
	}

	// Copy the binary to the destination
	srcFile, err := os.Open(binaryPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Seek to beginning of dest file and truncate
	if _, err := destFile.Seek(0, 0); err != nil {
		return err
	}
	if err := destFile.Truncate(0); err != nil {
		return err
	}

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	return nil
}

// runCommand runs a shell command
func runCommand(cmdStr string) error {
	return runCommandFunc(cmdStr)
}

// runCommandFunc is the function used to run commands (allows mocking in tests)
var runCommandFunc = func(cmdStr string) error {
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// replaceBinary replaces the current binary with the new one
func replaceBinary(newBinary, currentBinary string) error {
	// Make the new binary executable
	if err := os.Chmod(newBinary, 0755); err != nil {
		return err
	}

	// Rename (atomic on most systems)
	if err := os.Rename(newBinary, currentBinary); err != nil {
		// If rename fails (e.g., cross-device), try copy
		return copyFile(newBinary, currentBinary)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Make executable
	return os.Chmod(dst, 0755)
}

// normalizeVersion removes 'v' or 'V' prefix and trims whitespace
func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}

// isNewerVersion compares two semver-like versions
// Returns true if latest is newer than current
func isNewerVersion(latest, current string) bool {
	latestParts := parseVersionParts(latest)
	currentParts := parseVersionParts(current)

	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}

	return len(latestParts) > len(currentParts)
}

// parseVersionParts splits a version string into numeric parts
func parseVersionParts(v string) []int {
	// Remove any suffix like -beta, -rc1, etc.
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	result := make([]int, 0, len(parts))

	for _, p := range parts {
		var num int
		fmt.Sscanf(p, "%d", &num)
		result = append(result, num)
	}

	return result
}

// FormatUpdateMessage formats a user-friendly update notification
func (u *UpdateInfo) FormatUpdateMessage() string {
	return fmt.Sprintf("New version available: %s (current: %s) - %s",
		u.LatestVersion, u.CurrentVersion, u.ReleaseURL)
}
