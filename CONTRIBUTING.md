# Contributing to This Project

Whether you're a Go newbie or an expert, contributions are welcome.

---

## Getting Started

Before making changes, open an issue to discuss your idea. This helps avoid duplicate work and ensures your contribution aligns with the project's direction.

---

## All Code Changes Happen via Pull Requests

Here's the workflow:

1. Fork the repo and create a branch off `master`.
2. Write code. Add tests if needed.
3. Update documentation if you add or change features.
4. Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines.
5. Test your changes locally.
6. Write a [clear commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).
7. Open a pull request describing your changes and referencing related issues.

---

## Dependency Management

We use Go modules for dependency management. If you need to update or add dependencies for your PR:

1. Make your code changes as needed.
2. To update or add a dependency, run:
   ```sh
   go get -u github.com/username/modulename
   go mod tidy
   ```
3. Commit the changes to `go.mod` and `go.sum`.

If you have questions or issues with dependency management, feel free to reach out in your pull request or open an issue.

---

## Code of Conduct

By participating in this community, you agree to follow our [Code of Conduct][code of conduct]. Please help us keep this community open, inclusive, and respectful.

[code-of-conduct]: https://github.com/jr-k/d4s/blob/master/CONDUCT.md

---

## License

All contributions are made under the [MIT License](http://choosealicense.com/licenses/mit/). By submitting code, you agree that your contributions are released under the same license as this project. Except for [jr-k](https://github.com/jr-k) who is the maintainer of this project and has any right on the project.

If you have concerns about licensing, please [open an issue](https://github.com/jr-k/d4s/issues) or reach out to the maintainers.

---

## Reporting Bugs

Please use [GitHub Issues](https://github.com/jr-k/d4s/issues) to report bugs or request features.

- Open a new issue with a **clear title** and **detailed description**.
- Include logs, screenshots, steps to reproduce, and environment info when possible.

---

Thanks for contributing.
