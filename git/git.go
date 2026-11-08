package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dawidl022/gitlab-contribution-sync/config"
	"github.com/dawidl022/gitlab-contribution-sync/contributions"
)

// a commit is made with the message "Contribution #<count> on <date>", and is
// timestamped with midday local time on that date (naive attempt at mitigating
// timezone issues).
func CommitContributions(targetDir string, date contributions.Date, count int) error {
	err := gitAdd(targetDir)
	if err != nil {
		return err
	}

	// auto maintenance repacks after every commit and races the next one, corrupting the repo
	cmd := exec.Command("git", "-c", "maintenance.auto=false", "-c", "gc.auto=0",
		"commit", "-m", fmt.Sprintf("\"Contribution #%d on %s\"", count, date))
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_AUTHOR_DATE=%sT12:00:00:00", date))
	cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_COMMITTER_DATE=%sT12:00:00:00", date))
	cmd.Dir = targetDir

	return cmd.Run()
}

func gitAdd(targetDir string) error {
	cmd := exec.Command("git", "add", contributions.Filename)
	cmd.Dir = targetDir

	return cmd.Run()
}

// GitHub only counts ~1000 commits per push on the contribution graph (observed),
// so unpushed commits go up oldest-first in batches.
const PushBatchSize = 900

func PushContributions(targetDir string) error {
	return pushInBatches(targetDir, PushBatchSize)
}

func pushInBatches(targetDir string, batchSize int) error {
	cmd := exec.Command("git", "rev-list", "--reverse", "HEAD", "--not", "--remotes=origin")
	cmd.Dir = targetDir
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	unpushed := strings.Fields(string(out))

	for i := batchSize; i < len(unpushed); i += batchSize {
		err := push(targetDir, unpushed[i-1]+":refs/heads/main")
		if err != nil {
			return err
		}
	}
	return push(targetDir, "main")
}

func push(targetDir string, refspec string) error {
	cmd := exec.Command("git", "push", "-u", "origin", refspec)
	cmd.Dir = targetDir
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func InitRepo(config config.Config) error {
	if !repoExists(config.TargetDir) {
		err := initGitRepo(config.TargetDir)
		if err != nil {
			return err
		}
		return addRemote(config)
	}
	return nil
}

func repoExists(targetDir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = targetDir

	err := cmd.Run()
	return err == nil
}

func initGitRepo(targetDir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = targetDir

	return cmd.Run()
}

func addRemote(config config.Config) error {
	cmd := exec.Command("git", "remote", "add", "origin", config.GithubRepo)
	cmd.Dir = config.TargetDir

	return cmd.Run()
}
