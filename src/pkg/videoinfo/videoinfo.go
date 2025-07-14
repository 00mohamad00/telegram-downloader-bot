package videoinfo

import "fmt"

type VideoInfo struct {
	URL         string
	ContentType string
	Size        int64
	Filename    string
}

// FormatSize returns the size of the video in a human-readable format
func (vi *VideoInfo) FormatSize() string {
	if vi.Size == -1 {
		return "Unknown size"
	}
	if vi.Size == 0 {
		return "0 B"
	}

	const unit = 1024
	if vi.Size < unit {
		return fmt.Sprintf("%d B", vi.Size)
	}

	div, exp := int64(unit), 0
	for n := vi.Size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(vi.Size)/float64(div), "KMGTPE"[exp])
}
