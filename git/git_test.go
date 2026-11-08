package git

import (
	"os/exec"
	"strings"
	"testing"
)

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestPushInBatches(t *testing.T) {
	remote, local := t.TempDir(), t.TempDir()
	run(t, remote, "init", "-q", "--bare", "-b", "main")
	run(t, remote, "config", "core.logAllRefUpdates", "true") // reflog = one entry per push
	run(t, local, "init", "-q", "-b", "main")
	run(t, local, "remote", "add", "origin", remote)

	commit := func(n int) {
		for i := 0; i < n; i++ {
			run(t, local, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "c")
		}
	}

	// 5 commits, batch 2 -> pushes at #2, #4, then main (#5)
	commit(5)
	if err := pushInBatches(local, 2); err != nil {
		t.Fatal(err)
	}
	// later run: 1 new commit -> single push
	commit(1)
	if err := pushInBatches(local, 2); err != nil {
		t.Fatal(err)
	}

	if got, want := run(t, remote, "rev-parse", "main"), run(t, local, "rev-parse", "HEAD"); got != want {
		t.Fatalf("remote main = %s, want %s", got, want)
	}
	if got := len(strings.Split(run(t, remote, "reflog", "main"), "\n")); got != 4 {
		t.Fatalf("pushes = %d, want 4", got)
	}
}
