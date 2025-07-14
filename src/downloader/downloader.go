package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/00mohamad00/telegram-downloader-bot/src/pkg/videoinfo"
)

type VideoDownloader struct {
	DownloadDir string
	Client      *http.Client
}

func NewVideoDownloader(downloadDir string, timeout time.Duration) *VideoDownloader {
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create download directory: %v\n", err)
	}

	return &VideoDownloader{
		DownloadDir: downloadDir,
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (vd *VideoDownloader) DownloadVideo(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add user agent to avoid blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := vd.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status: %s", resp.Status)
	}

	filename := vd.generateFilename(url, resp.Header.Get("Content-Type"))
	filePath := filepath.Join(vd.DownloadDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save video: %w", err)
	}

	return filePath, nil
}

func (vd *VideoDownloader) generateFilename(url, contentType string) string {
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	if !strings.Contains(filename, ".") {
		ext := vd.getExtensionFromContentType(contentType)
		if ext != "" {
			filename += ext
		} else {
			filename += ".mp4"
		}
	}

	if filename == "" || filename == ".mp4" {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("video_%s.mp4", timestamp)
	}

	return filename
}

func (vd *VideoDownloader) getExtensionFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "video/mp4"):
		return ".mp4"
	case strings.Contains(contentType, "video/webm"):
		return ".webm"
	case strings.Contains(contentType, "video/avi"):
		return ".avi"
	case strings.Contains(contentType, "video/mov"):
		return ".mov"
	case strings.Contains(contentType, "video/wmv"):
		return ".wmv"
	case strings.Contains(contentType, "video/flv"):
		return ".flv"
	case strings.Contains(contentType, "video/mkv"):
		return ".mkv"
	default:
		return ""
	}
}

func (vd *VideoDownloader) IsValidVideoURL(url string) bool {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}

	videoExtensions := []string{".mp4", ".webm", ".avi", ".mov", ".wmv", ".flv", ".mkv"}
	urlLower := strings.ToLower(url)

	for _, ext := range videoExtensions {
		if strings.Contains(urlLower, ext) {
			return true
		}
	}

	return true
}

func (vd *VideoDownloader) GetVideoInfo(url string) (*videoinfo.VideoInfo, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := vd.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %s", resp.Status)
	}

	return &videoinfo.VideoInfo{
		URL:         url,
		ContentType: resp.Header.Get("Content-Type"),
		Size:        resp.ContentLength,
		Filename:    vd.generateFilename(url, resp.Header.Get("Content-Type")),
	}, nil
}
