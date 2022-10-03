# gh-prx

<img src="https://github.com/ilaif/gh-prx/raw/main/assets/logo.png" width="200">

A GitHub (`gh`) CLI extension to automate the daily work with **branches**, **commits** and **pull requests**.

## Usage

1. Checking out to an automatically generated branch:

    ```sh
    gh prx checkout-new 1234 # Where 1234 is the issue's number/code
    ```

2. Creating a new PR with automatically generated title/body and checklist prompt:

    ```sh
    gh prx create
    ```

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
   gh extension install .
   ```

</details>

## Configuration

[TODO]
