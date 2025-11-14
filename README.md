# gh-createpr
### Github CLI Pull Request Creator

GitHub Pull Request Tool is a command-line interface (CLI) tool written in Go to streamline the process of creating pull requests on GitHub, adding reviewers, and managing reviewers' configurations.

### Installation
Since it's GitHub CLI Extension, make sure you have [GitHub CLI](https://cli.github.com) installed on your system.

To install the extension, please execute the following command:

```bash
$ gh extension install ataberkcanitez/gh-createpr
```

### Usage

After installing the extension, please execute following command to use:

```bash
$ gh createpr
```

#### Setting Target Branch

You can configure a default target branch for your pull requests. This is useful when you frequently create PRs against a specific branch (e.g., `develop`, `staging`):

```bash
# Set the default target branch
$ gh createpr --set-target-branch develop

# Create a PR - it will automatically target the configured branch
$ gh createpr
```

If no target branch is configured, the pull request will use GitHub's default base branch for the repository.

### Options

* `-h --help`: Show help for command.
* `--title <string>`: Title of the pull request.
* `--body <string>`: Body of the pull request.
* `--add-reviewer <string>`: Add reviewer to the pull request.
* `--assignee <string>`: Assignee of the pull request.
* `--remove-reviewer <string>`: Remove reviewer from the pull request.
* `--set-target-branch <string>`: Set the default target branch for pull requests.
* `--list`: List all configuartions for the cli tool.
* `--list-reviewers`: List reviewers of the pull request.

### License
This project is licensed under the MIT License - see the [LICENSE](LICENSE.md) file for details.


### Acknowledgments

- GitHub CLI for providing the foundation for interacting with GitHub from the command line.

### Contributing

Feel free to contribute by opening issues or creating pull requests. Your feedback and involvement are highly encouraged!

---
Enjoy using `createpr` and feel free to reach out with any feedback or suggestions!
