# gh-prx

<img src="https://github.com/ilaif/gh-prx/raw/main/assets/logo.png" width="200">

A GitHub (`gh`) CLI extension to automate the daily work with **branches**, **commits** and **pull requests**.

## Usage

[Jump to installation](#installation)

1. Checking out an automatically generated branch from issues:

   ```sh
   gh prx checkout-new
   ```

   <img src="https://github.com/ilaif/gh-prx/raw/main/assets/gh-prx-checkout-new.gif" width="700">

2. Checking out an automatically generated branch based on an issue id:

   ```sh
   gh prx checkout-new 1234 # Where 1234 is the issue's key. If not provided, a list of issues will be prompted.
   ```

   <img src="https://github.com/ilaif/gh-prx/raw/main/assets/gh-prx-checkout-new-issue.gif" width="700">

3. Creating a new PR with automatically generated title/body and checklist prompt:

   ```sh
   gh prx create
   ```

   <img src="https://github.com/ilaif/gh-prx/raw/main/assets/gh-prx-create.gif" width="700">

> Explore further by running `gh prx --help`

## Why?

As developers, we rely heavily on git and our git provider (in this case, GitHub).

Many of us find the terminal and CLI applications as our main toolkit.

`gh-prx` helps with automating and standardizing how we work with git and GitHub to create a faster and more streamlined workflow for individuals and teams.

## Features

- Automatically creating new branches named based on issues fetched from project management tools
  - Currently supported: GitHub, Jira, Linear
- Extended PR creation:
  - Automatically push branch to origin
  - Parse branch names by a pattern into a customized PR title and description template
  - Add labels based on issue types
  - Filter commits and display them in the PR description
  - Interactively answer PR checklists before creating the PR
  - Use AI (🔮) to summarize the PR's changes
  - All `gh pr create` original flags are extended into the tool

> `gh-prx` is an early-stage project. Got a new feature in mind? Open a pull request or a [feature request](https://github.com/ilaif/gh-prx/issues/new) 🙏

## Configuration

Configuration is provided from `.github/.gh-prx.yaml` and is advised to be committed to git to maintain standardization across the team.

The default values for `.gh-prx.yaml` are:

```yaml
branch:
   template: "{{.Type}}/{{with .Issue}}{{.}}-{{end}}{{.Description}}" # Branch name template
   pattern: "{{.Type}}\\/({{.Issue}}-)?{{.Description}}" # Branch name pattern
   variable_patterns: # A map of patterns to match for each template variable
      Type: "fix|feat|chore|docs|refactor|test|style|build|ci|perf|revert"
      Issue: "([a-zA-Z]+\-)*[0-9]+"
      Author: "[a-zA-Z0-9]+"
      Description: ".*"
   token_separators: ["-", "_"] # Characters used to separate branch name into a human-readable string
   max_length: 60 # Max characters to allow for branch length without prompting for changing it
pr:
   title: "{{.Type}}{{with .Issue}}({{.}}){{end}}: {{humanize .Description}}" # PR title template
   ignore_commits_patterns: ["^wip"] # Patterns to filter out a commits from the {{.Commits}} variable
   answer_checklist: true # Whether to prompt the user to answer PR description checklists. Possible answers: yes, no, skip (remove the item)
   push_to_remote: true # Whether to push the local changes to remote before creating the PR
issue:
   provider: github # The provider to use for fetching issue details (supported: github,jira,linear)
   types: ["fix", "feat", "chore", "docs", "refactor", "test", "style", "build", "ci", "perf", "revert"] # The issue types to prompt the user when creating a new branch
checkout_new:
   jira:
      project: "" # The Jira project key to use when creating a new branch
      issue_jql: "[<jira_project>+AND+]assignee=currentUser()+AND+statusCategory!=Done+ORDER+BY+updated+DESC" # The Jira JQL to use when fetching issues. <jira_project> is optional and will be replaced with the project key that is configured in the `project` field.
   github:
      issue_list_flags: ["--state", open", "--assignee", "@me"] # The flags to use when fetching issues from GitHub
   # linear: # Due to Linear's GraphQL API, the issue list is not configurable. The default is: `assignedIssues(orderBy: updatedAt, filter: { state: { type: { neq: \"completed\" } } })`
pull_request_template_path: "./pull_request_template.md" # The pull request template file to use when creating a new PR. Relative to the repository root.
```

### PR Description (Body)

The PR description is based on the `pull_request_template_path` variable which defaults to the repo's `.github/pull_request_template.md`. If this file does not exist, a default template is used:

```markdown
{{with .Issue}}Closes #{{.}}.

{{end}}## Description

{{if .AISummary}}{{.AISummary}}{{ else }}{{humanize .Description}}

Change(s) in this PR:
{{range $commit := .Commits}}

- {{$commit}}
  {{- end}}
  {{- end}}

## PR Checklist

- [ ] Tests are included
- [ ] Documentation is changed or added
```

## Templating

The templates are based on [Go text template](https://pkg.go.dev/text/template).

### Additional template functions

`humanize`:

Humanizes a string by separating it into tokens (words) based on `branch.token_separators`.

Example:

Given:

- `token_separators: ["-"]`
- `{{.Description}}`: "my-dashed-string"
- Template:

  ```go-template
  This is "{{ humanize .Description}}"
  ```

Result:

```txt
This is "my dashed string"
```

`title`:

Capitalizes the first letter of each word in a string.

`lower`:

Lower cases a string.

`upper`:

Upper cases a string.

### Special template variable names

- `{{.Type}}` - Used to interpret GitHub labels to add to the PR and issue type to add the branch name.
- `{{.Issue}}` - Used as a placeholder for the issue key when creating a new branch.
- `{{.Description}}` - Used as a placeholder for the issue title when creating a new branch.
- `{{.Commits}}` - Used as a placeholder in a PR description (body) to iterate over filtered commits.
- `{{.AISummary}}` - Used as a placeholder in a PR description (body) to add a summary of the PR's changes based on AI.

## AI summary configuration

Just export your OpenAI key as an env var named `OPENAI_API_KEY` and you're good to go.

If `OPENAI_API_KEY` is not set, the AI summary will not be used.

To disable the AI summary explicitly, use the `--no-ai-summary` flag.

## Providers

There are currently 3 providers supported: GitHub, Jira and Linear.

### GitHub

The GitHub provider is the default provider and is configured simply by running `gh auth login`.

### Jira

To setup, run `gh prx setup provider jira --endpoint <endpoint> --user <email> --token <token>`.

Alternatively, set the `JIRA_ENDPOINT`, `JIRA_USER` and `JIRA_TOKEN` env vars.

### Linear

To setup, run `gh prx setup provider linear --api-key <api-key>`.

Alternatively, set the `LINEAR_API_KEY` env var.

## Installation

1. Install the `gh` CLI - see the [installation](https://github.com/cli/cli#installation)

   _Installation requires a minimum version (2.0.0) of the the GitHub CLI that supports extensions._

2. Install this extension:

   ```sh
   gh extension install ilaif/gh-prx
   ```

<details>
   <summary><strong>Installing Manually</strong></summary>

> If you want to install this extension **manually**, follow these steps:

1. Clone the repo

   ```bash
   # git
   git clone https://github.com/ilaif/gh-prx
   # GitHub CLI
   gh repo clone ilaif/gh-prx
   ```

2. Cd into it

   ```bash
   cd gh-prx
   ```

3. Install it locally

   ```bash
   make install-extension-local
   ```

</details>

## Questions, bug reporting and feature requests

You're more than welcome to [Create a new issue](https://github.com/ilaif/gh-prx/issues/new) or contribute.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change. See [CONTRIBUTING.md](CONTRIBUTING.md) for more information.

## License

gh-prx is licensed under the [MIT](https://choosealicense.com/licenses/mit/) license. For more information, please see the [LICENSE](LICENSE) file.
