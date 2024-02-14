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

### Options

* `-h --help`: Show help for command.
* `--title <string>`: Title of the pull request.
* `--body <string>`: Body of the pull request.
* `--add-reviewer`: Add reviewer to the pull request.
* `--remove-reviewer`: Remove reviewer from the pull request.
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
