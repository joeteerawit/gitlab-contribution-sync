# Gitlab Contribution Sync

A simple tool to sync your GitLab contribution count to your GitHub profile.
Only public data on your GitLab profile is used to determine the contribution
counts, so it is safe to sync private contributions.

Caveat: Due to the nature of the public GitLab API, only contributions from the
last 12 months can be synchronised.

## Usage

### Prerequisites

1. Create a blank private repo in your GitHub account.
2. Ensure you have [`git`](https://git-scm.com/) installed locally.
3. Ensure you are able to push to the private repo you created in step 1 (e.g.
   [via
   SSH](https://docs.github.com/en/authentication/connecting-to-github-with-ssh)).
4. If you wish to sync private contribution counts from your GitLab profile too,
   ensure you have enabled the "Include private contributions on your profile"
   option in your GitLab profile settings.
   
### Installation

Install from a clone of this fork. `go.mod` still declares the upstream module
path, so `go install github.com/joeteerawit/gitlab-contribution-sync@latest`
fails.

```bash
git clone https://github.com/joeteerawit/gitlab-contribution-sync && cd gitlab-contribution-sync && go install .
```

The binary lands in `$(go env GOBIN)` (or `$(go env GOPATH)/bin` if `GOBIN` is
unset), which must be on your `PATH`. With asdf, run `asdf reshim golang`.

### CLI

```bash
gitlab-contribution-sync <gitlab-username> <local-target-directory> <remote-github-repo>
```

For a self-hosted GitLab, set `GITLAB_HOST` (default `gitlab.com`):

```bash
GITLAB_HOST=git.example.com gitlab-contribution-sync <gitlab-username> <local-target-directory> <remote-github-repo>
```

`GITLAB_HOST` must be a host name only, with no scheme and no path (e.g.
`git.example.com`, not `https://git.example.com/`).

where:

- `<gitlab-username>` is a placeholder for your GitLab username.
- `<local-target-directory>` is a placeholder for a path on your local
  filesystem, where the tool will create and update a local repository. This
  directory will be created upon first usage.
- `<remote-github-repo>` is a placeholder for the remote (push) address of the
  GitHub repo you want to sync your contributions to.

e.g.

```bash
gitlab-contribution-sync "dawidl022" "$HOME/Development/my-gitlab-contributions" "git@github.com:dawidl022/my-gitlab-contributions.git"
```

To re-sync your contributions, simply re-run the command with the same
arguments. You can also set up a cron job to periodically run the tool.

If `<local-target-directory>` is already a git repo, the tool reuses it and its
existing `origin` remote (ignoring `<remote-github-repo>`), and pushes to branch
`main`.

## Troubleshooting

- **Nothing syncs or the calendar is empty:** open
  `https://<host>/users/<username>/calendar.json` in a private window. If it
  returns `{}`, enable "Include private contributions on your profile" (User
  Settings → Profile) and make sure the profile is not private.
- **The run stops with `exit status 128` partway through:** progress is saved in
  `contributions.json` after every commit, so re-running the same command
  resumes. Git errors are not printed. Run `git push -u origin main` in the
  target directory to see push errors.
- **Commits don't show on the GitHub graph:** the commit author email
  (`git config user.email` in the target directory) must be verified on your
  GitHub account. For a private target repo, enable "Include private
  contributions on my profile" on GitHub.
- **Push fails with `remote unpack failed` / `missing blob`:** git's automatic
  maintenance repacks after every commit and can race with the next commit,
  corrupting the local repo. Run the first (large) sync with auto-maintenance
  off for that run only:

  ```bash
  GIT_CONFIG_COUNT=2 GIT_CONFIG_KEY_0=maintenance.auto GIT_CONFIG_VALUE_0=false \
    GIT_CONFIG_KEY_1=gc.auto GIT_CONFIG_VALUE_1=0 \
    gitlab-contribution-sync <gitlab-username> <local-target-directory> <remote-github-repo>
  ```

  If the repo is already corrupt (`git fsck` reports missing objects) and
  nothing was pushed, reset to the last pushed commit, delete
  `contributions.json` and re-run.
- A first sync can create thousands of commits and take several minutes.
- **Only the latest ~1,000 commits show on the GitHub graph:** GitHub appears to
  count at most ~1,000 commits per push (observed, not documented). For a large
  first sync, point `<remote-github-repo>` at a local bare repo so the tool's
  single push stays local, then push to GitHub oldest-first in batches of 900:

  ```bash
  for n in $(seq 900 900 $(git rev-list --count HEAD)) $(git rev-list --count HEAD); do
    git push origin "$(git rev-list --reverse HEAD | sed -n "${n}p"):refs/heads/main"
  done
  ```

  Commits GitHub already received are not recounted, so if a big push was
  already made, rebuild the history in a fresh directory (new commit IDs) and
  force-push the first batch.

## How it works

The tool uses the same API endpoint GitLab uses to render the "calendar" on a
user's profile: `https://gitlab.com/users/<username>/calendar.json`. For each
contribution count in the calendar, a commit is created in the private GitHub
repository with the correct date (but not time) of the GitLab contribution.

## Future development

- [go-git](https://github.com/go-git/go-git) could be used to move the file and
git operations into memory, in order to improve performance.

- descriptive error messages when something fails
