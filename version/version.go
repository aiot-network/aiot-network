package version

import (
	"fmt"
)

var (
	GitCommitLog = "unknown_unknown"
	GitStatus    = "unknown_unknown"
	Version      = "v0.2.4"
)

func StringifySingleLine(appName string) string {
	if len(GitCommitLog) < 7 {
		return ""
	}
	if GitStatus != "" {
		GitCommitLog = GitCommitLog[0:7] + "-dirty"
	} else {
		GitCommitLog = GitCommitLog[0:7]
	}
	return fmt.Sprintf("%s-%s-%s", appName, Version, GitCommitLog)
}
