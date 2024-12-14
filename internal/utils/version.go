package utils

import "fmt"

func FormatVersion(sha, branch string) string {
	if sha == "" {
		sha = "unknown"
	}
	if branch == "" {
		branch = "unknown"
	}
	return fmt.Sprintf("%s (%s)", branch, sha)
}
