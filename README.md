# gh-createpr
### Github CLI Pull Request Creator

GitHub Pull Request Tool is a command-line interface (CLI) tool written in Go to streamline the process of creating pull requests on GitHub, adding reviewers, and managing reviewers' configurations.

### Installation
Since it's GitHub CLI Extension, make sure you have [GitHub CLI](https://cli.github.com) installed and authenticated (`gh auth login`).

To install the extension, please execute the following command:

```bash
$ gh extension install ataberkcanitez/gh-createpr
```

To upgrade to the latest version:

```bash
$ gh extension upgrade createpr
```

### Usage

From a branch you want to open a pull request for, run:

```bash
$ gh createpr
```

The tool will:

1. Let you pick the commit to use as the pull request title.
2. Ask for the title (pre-filled with the chosen commit) and an optional body.
3. Create the pull request, then add your configured reviewers and assignee.

#### Choosing the title

When your branch has more than one commit, you can choose which one the title comes from:

```
? Which commit should the title come from?  [Use arrows to move, type to filter]
  Write my own title
> 5145d77  AT: Execute go mod tidy
  16a35c9  AT: Real job done
```

- Press **Enter** to use the latest commit, as before.
- Use the arrow keys to pick an older commit, or **type to filter** the list.
- Choose **Write my own title** to type a title from scratch.

Only commits on your branch are listed (commits that are not yet on the target branch). With a single commit, the picker is skipped and its message is suggested as the title.

#### Skipping the prompts

Pass `--title` and/or `--body` to skip the matching prompt, for example in scripts:

```bash
$ gh createpr --title "Fix login redirect" --body "Closes #42"
```

When the input is not a terminal (e.g. piped), the tool falls back to simple line-by-line prompts.

#### Setting Target Branch

You can configure a default target branch for your pull requests. This is useful when you frequently create PRs against a specific branch (e.g., `develop`, `staging`):

```bash
# Set the default target branch
$ gh createpr --set-target-branch develop

# Create a PR - it will automatically target the configured branch
$ gh createpr
```

If no target branch is configured, the pull request will use GitHub's default base branch for the repository.

> **Note:** the target branch is a global setting. If a repository does not have that branch, creating the pull request will fail. Change it with `--set-target-branch` for that repository.

#### Managing reviewers and assignee

```bash
# Reviewers are requested on every pull request you create
$ gh createpr --add-reviewer octocat
$ gh createpr --add-reviewer my-org/my-team
$ gh createpr --remove-reviewer octocat
$ gh createpr --list-reviewers

# Change the default assignee (saved, and used for this and future PRs)
$ gh createpr --assignee octocat
```

### Options

* `-h --help`: Show help for command.
* `--title <string>`: Title of the pull request (skips the title prompt).
* `--body <string>`: Body of the pull request (skips the body prompt).
* `--add-reviewer <string>`: Add a reviewer to your configuration.
* `--remove-reviewer <string>`: Remove a reviewer from your configuration.
* `--list-reviewers`: List the reviewers in your configuration.
* `--assignee <string>`: Set the default assignee in your configuration, then create the pull request.
* `--set-target-branch <string>`: Set the default target branch for pull requests.
* `--list`: List all configurations for the cli tool.

### Configuration

Settings are stored in `~/gh-createpr-configuration.yml`, which is created automatically on first run:

```yaml
reviewers:
    - octocat
    - my-org/my-team
assignee: '@me'
target_branch: develop
```

### Colors

Output is colored when running in a terminal. Colors follow the same rules as `gh`: set `NO_COLOR=1` to disable them, or `CLICOLOR_FORCE=1` to force them.

### Development

Requires the Go version listed in `go.mod`.

```bash
# Run the tests
$ go test ./...

# Build and run the local version
$ go build -o /tmp/gh-createpr . && /tmp/gh-createpr
```

To try the full flow without creating a real pull request, point `GH_PATH` at a fake `gh` that only logs its arguments, and use a temporary `HOME` so your real configuration is untouched:

```bash
$ printf '#!/bin/sh\necho "gh $*" >> /tmp/fake-gh.log\n[ "$1 $2" = "pr create" ] && echo "https://github.com/example/repo/pull/1"\nexit 0\n' > /tmp/fake-gh && chmod +x /tmp/fake-gh
$ mkdir -p /tmp/gh-createpr-home
$ GH_PATH=/tmp/fake-gh HOME=/tmp/gh-createpr-home /tmp/gh-createpr
$ cat /tmp/fake-gh.log
```

Releases are built by GitHub Actions when a `v*` tag is pushed.

### License
This project is licensed under the MIT License - see the [LICENSE](LICENSE.md) file for details.


### Acknowledgments

- GitHub CLI for providing the foundation for interacting with GitHub from the command line.

### Contributing

Feel free to contribute by opening issues or creating pull requests. Your feedback and involvement are highly encouraged!

---
Enjoy using `createpr` and feel free to reach out with any feedback or suggestions!
