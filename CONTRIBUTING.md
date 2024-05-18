<!--
THIS FILE IS MANAGED BY IAC
AND WILL BE OVERWRITTEN IF MODIFIED
github-infra:/assets/contributing_oss.md
-->

# Contributing

Contributions are welcome and help create better software. As a contributor we have
a few guidelines to follow;

## Commit Message Format

Commit messages need to be in a specific format, there are CI steps to validate the format - these
are then used to autogenerate changelogs.

If the commit contains a breaking change - this needs to be included in the body;

```
BREAKING CHANGE: summary of breaking change
```

This will bump the release by a major version when generated.

### Type

Must be one of the following:

* **build**: Changes that affect the build system (github workflows) or external dependencies
* **docs**: Documentation only changes
* **feat**: A new feature
* **fix**: A bug fix
* **perf**: A code change that improves performance
* **refactor**: A code change that neither fixes a bug nor adds a feature, code style changes
* **test**: Adding missing tests or correcting existing tests
* **chore**: No production code changes, e.g. updating codeowners
