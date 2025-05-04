# Codegen CI/CD Workflow

This directory contains a Python script that implements a full CI/CD pipeline using the Codegen API, GitHub webhooks, and Ngrok for exposing local endpoints. It follows a workflow pattern similar to the AIGNE orchestrator, breaking down tasks into steps and distributing them to specialized agents.

## Features

- **Requirements Analysis**: Automatically analyze REQUIREMENTS.md and create PRs implementing requirements
- **PR Review**: Review PRs against requirements and approve or adjust as needed
- **Test Coverage**: Create test branches with full test coverage and deployment scripts
- **Deployment**: Deploy and verify changes, then merge or fix issues
- **Progress Tracking**: Update REQUIREMENTS.md with progress checkboxes

## Architecture

The workflow is implemented as a multi-step process:

1. **Step 1**: Analyze REQUIREMENTS.md and create a PR implementing the requirements
2. **Step 2**: When a PR is created, review against requirements and approve or adjust
3. **Step 3**: Create a test branch with full test coverage and deployment scripts
4. **Step 4**: Deploy and verify changes, then merge or fix issues

Each step is triggered by a GitHub webhook event and processed by a Codegen agent.

## Prerequisites

- Python 3.6 or higher
- `requests` library (`pip install requests`)
- GitHub API token with repo scope
- Ngrok API key
- Codegen API token and organization ID

## Installation

1. Clone this repository:
   ```bash
   git clone https://github.com/Zeeeepa/point.git
   cd point/cicd
   ```

2. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```

## Usage

### Setup

Set up the CI/CD workflow:

```bash
python codegen_cicd.py --setup
```

This will:
- Prompt for your GitHub API token, Ngrok API key, and Codegen credentials
- List your GitHub repositories and let you select one
- Create an Ngrok tunnel to expose your local webhook server
- Create a GitHub webhook for the selected repository
- Save the configuration to `codegen_cicd_config.json`

### Run

Run the CI/CD workflow:

```bash
python codegen_cicd.py --run
```

This will:
- Start a webhook server to receive GitHub events
- Trigger an initial requirements analysis
- Process webhook events as they come in
- Keep running until interrupted (Ctrl+C)

### Teardown

Tear down the CI/CD workflow:

```bash
python codegen_cicd.py --teardown
```

This will:
- Delete the GitHub webhook
- Delete the configuration file

## Configuration

The configuration is stored in `codegen_cicd_config.json` with the following structure:

```json
{
  "github_token": "your-github-token",
  "ngrok_api_key": "your-ngrok-api-key",
  "repo_owner": "repository-owner",
  "repo_name": "repository-name",
  "webhook_id": 12345678,
  "tunnel_url": "https://your-ngrok-tunnel.ngrok.io",
  "codegen_api_token": "your-codegen-api-token",
  "codegen_org_id": "your-codegen-org-id"
}
```

## Webhook Events

The workflow processes the following GitHub webhook events:

- **pull_request**: Triggered when a PR is opened or updated
- **check_suite**: Triggered when a check suite completes
- **push**: Triggered when code is pushed to the repository

## Example Workflow

1. Create a `REQUIREMENTS.md` file in your repository with a list of requirements
2. Run the CI/CD workflow
3. The workflow will analyze the requirements and create a PR
4. When the PR is created, the workflow will review it
5. If the PR is approved, it will be merged
6. If the PR needs adjustments, a new PR will be created
7. After merging, a test branch will be created with full test coverage
8. The changes will be deployed and verified
9. If successful, the test branch will be merged
10. If issues are found, a fix PR will be created
11. The `REQUIREMENTS.md` file will be updated with progress

## Logging

Logs are written to `codegen_cicd.log` and also displayed on the console.

## License

MIT

