#!/usr/bin/env python3
"""
Codegen CI/CD Workflow Script

This script implements a full CI/CD pipeline using the Codegen API, GitHub webhooks,
and Ngrok for exposing local endpoints. It follows a workflow pattern similar to
the AIGNE orchestrator, breaking down tasks into steps and distributing them to
specialized agents.

Features:
- Analyze REQUIREMENTS.md and create PRs implementing requirements
- Review PRs against requirements and approve or adjust
- Create test branches with full test coverage and deployment scripts
- Deploy and verify changes, then merge or fix issues
- Update REQUIREMENTS.md with progress

Usage:
    python codegen_cicd.py --setup    # Configure API keys and repository
    python codegen_cicd.py --run      # Start the webhook server and CI/CD cycle
    python codegen_cicd.py --teardown # Remove webhooks and stop tunnels
"""

import argparse
import json
import logging
import os
import re
import signal
import sys
import time
import threading
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
from typing import Dict, List, Optional, Union, Any

import requests

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler("codegen_cicd.log"),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger("codegen_cicd")

# Constants
CONFIG_FILE = "codegen_cicd_config.json"
DEFAULT_PORT = 8000
MAX_RETRIES = 3
WEBHOOK_EVENTS = ["pull_request", "check_suite", "push"]


class CodegenClient:
    """Client for interacting with the Codegen API."""

    def __init__(self, org_id: str, api_token: str):
        """Initialize the Codegen client.

        Args:
            org_id: The organization ID for Codegen.
            api_token: The API token for Codegen.
        """
        self.org_id = org_id
        self.api_token = api_token
        self.base_url = "https://api.codegen.com/v1"
        self.headers = {
            "Authorization": f"Bearer {api_token}",
            "Content-Type": "application/json",
        }
        logger.info("Initialized Codegen client")

    def run_agent(self, prompt: str, wait: bool = True) -> Dict[str, Any]:
        """Run a Codegen agent with the given prompt.

        Args:
            prompt: The prompt to send to the agent.
            wait: Whether to wait for the task to complete.

        Returns:
            The task result.
        """
        logger.info(f"Running Codegen agent with prompt: {prompt[:100]}...")
        
        # Create the task
        response = requests.post(
            f"{self.base_url}/tasks",
            headers=self.headers,
            json={
                "org_id": self.org_id,
                "prompt": prompt,
            },
        )
        response.raise_for_status()
        task = response.json()
        
        if not wait:
            return task
        
        # Wait for the task to complete
        task_id = task["id"]
        while True:
            response = requests.get(
                f"{self.base_url}/tasks/{task_id}",
                headers=self.headers,
            )
            response.raise_for_status()
            task = response.json()
            
            if task["status"] in ["completed", "failed"]:
                break
                
            logger.info(f"Task {task_id} status: {task['status']}")
            time.sleep(5)
        
        return task


class GitHubClient:
    """Client for interacting with the GitHub API."""

    def __init__(self, token: str):
        """Initialize the GitHub client.

        Args:
            token: The GitHub API token.
        """
        self.token = token
        self.base_url = "https://api.github.com"
        self.headers = {
            "Authorization": f"token {token}",
            "Accept": "application/vnd.github.v3+json",
        }
        logger.info("Initialized GitHub client")

    def get_repositories(self) -> List[Dict[str, Any]]:
        """Get a list of repositories the user has access to.

        Returns:
            A list of repository objects.
        """
        response = requests.get(
            f"{self.base_url}/user/repos",
            headers=self.headers,
        )
        response.raise_for_status()
        return response.json()

    def create_webhook(self, repo_owner: str, repo_name: str, webhook_url: str) -> Dict[str, Any]:
        """Create a webhook for the repository.

        Args:
            repo_owner: The owner of the repository.
            repo_name: The name of the repository.
            webhook_url: The URL to send webhook events to.

        Returns:
            The created webhook object.
        """
        logger.info(f"Creating webhook for {repo_owner}/{repo_name} to {webhook_url}")
        
        response = requests.post(
            f"{self.base_url}/repos/{repo_owner}/{repo_name}/hooks",
            headers=self.headers,
            json={
                "name": "web",
                "active": True,
                "events": WEBHOOK_EVENTS,
                "config": {
                    "url": webhook_url,
                    "content_type": "json",
                    "insecure_ssl": "0",
                },
            },
        )
        response.raise_for_status()
        return response.json()

    def delete_webhook(self, repo_owner: str, repo_name: str, webhook_id: int) -> None:
        """Delete a webhook from the repository.

        Args:
            repo_owner: The owner of the repository.
            repo_name: The name of the repository.
            webhook_id: The ID of the webhook to delete.
        """
        logger.info(f"Deleting webhook {webhook_id} from {repo_owner}/{repo_name}")
        
        response = requests.delete(
            f"{self.base_url}/repos/{repo_owner}/{repo_name}/hooks/{webhook_id}",
            headers=self.headers,
        )
        response.raise_for_status()

    def get_file_content(self, repo_owner: str, repo_name: str, path: str, ref: str = "main") -> str:
        """Get the content of a file from the repository.

        Args:
            repo_owner: The owner of the repository.
            repo_name: The name of the repository.
            path: The path to the file.
            ref: The branch or commit to get the file from.

        Returns:
            The content of the file.
        """
        logger.info(f"Getting content of {path} from {repo_owner}/{repo_name}@{ref}")
        
        response = requests.get(
            f"{self.base_url}/repos/{repo_owner}/{repo_name}/contents/{path}",
            headers=self.headers,
            params={"ref": ref},
        )
        response.raise_for_status()
        
        content = response.json()
        if content.get("type") != "file":
            raise ValueError(f"{path} is not a file")
            
        import base64
        return base64.b64decode(content["content"]).decode("utf-8")

    def create_pull_request(
        self, repo_owner: str, repo_name: str, title: str, body: str, head: str, base: str = "main"
    ) -> Dict[str, Any]:
        """Create a pull request.

        Args:
            repo_owner: The owner of the repository.
            repo_name: The name of the repository.
            title: The title of the pull request.
            body: The body of the pull request.
            head: The name of the branch where your changes are implemented.
            base: The name of the branch you want the changes pulled into.

        Returns:
            The created pull request object.
        """
        logger.info(f"Creating PR for {repo_owner}/{repo_name}: {title}")
        
        response = requests.post(
            f"{self.base_url}/repos/{repo_owner}/{repo_name}/pulls",
            headers=self.headers,
            json={
                "title": title,
                "body": body,
                "head": head,
                "base": base,
            },
        )
        response.raise_for_status()
        return response.json()

    def merge_pull_request(self, repo_owner: str, repo_name: str, pr_number: int) -> Dict[str, Any]:
        """Merge a pull request.

        Args:
            repo_owner: The owner of the repository.
            repo_name: The name of the repository.
            pr_number: The number of the pull request to merge.

        Returns:
            The merge result.
        """
        logger.info(f"Merging PR #{pr_number} for {repo_owner}/{repo_name}")
        
        response = requests.put(
            f"{self.base_url}/repos/{repo_owner}/{repo_name}/pulls/{pr_number}/merge",
            headers=self.headers,
            json={
                "merge_method": "merge",
            },
        )
        response.raise_for_status()
        return response.json()


class NgrokClient:
    """Client for interacting with the Ngrok API."""

    def __init__(self, api_key: str):
        """Initialize the Ngrok client.

        Args:
            api_key: The Ngrok API key.
        """
        self.api_key = api_key
        self.base_url = "https://api.ngrok.com"
        self.headers = {
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
            "Ngrok-Version": "2",
        }
        logger.info("Initialized Ngrok client")

    def create_tunnel(self, port: int) -> str:
        """Create a tunnel to the local port.

        Args:
            port: The local port to expose.

        Returns:
            The public URL of the tunnel.
        """
        logger.info(f"Creating Ngrok tunnel to port {port}")
        
        # First, check if we already have a tunnel for this port
        response = requests.get(
            f"{self.base_url}/tunnels",
            headers=self.headers,
        )
        response.raise_for_status()
        
        tunnels = response.json().get("tunnels", [])
        for tunnel in tunnels:
            if tunnel.get("forwards_to", "").endswith(f":{port}"):
                logger.info(f"Found existing tunnel: {tunnel['public_url']}")
                return tunnel["public_url"]
        
        # Create a new tunnel
        response = requests.post(
            f"{self.base_url}/tunnels",
            headers=self.headers,
            json={
                "name": f"codegen-cicd-{port}",
                "protocol": "http",
                "forwards_to": f"http://localhost:{port}",
            },
        )
        response.raise_for_status()
        
        tunnel = response.json()
        logger.info(f"Created tunnel: {tunnel['public_url']}")
        return tunnel["public_url"]

    def delete_tunnel(self, tunnel_id: str) -> None:
        """Delete a tunnel.

        Args:
            tunnel_id: The ID of the tunnel to delete.
        """
        logger.info(f"Deleting tunnel {tunnel_id}")
        
        response = requests.delete(
            f"{self.base_url}/tunnels/{tunnel_id}",
            headers=self.headers,
        )
        response.raise_for_status()


class WebhookHandler(BaseHTTPRequestHandler):
    """Handler for GitHub webhook events."""

    def do_POST(self):
        """Handle POST requests."""
        content_length = int(self.headers["Content-Length"])
        post_data = self.rfile.read(content_length)
        payload = json.loads(post_data.decode("utf-8"))
        
        # Get the event type from the headers
        event_type = self.headers.get("X-GitHub-Event")
        logger.info(f"Received {event_type} event")
        
        # Process the event
        try:
            self.process_event(event_type, payload)
            self.send_response(200)
            self.send_header("Content-type", "text/plain")
            self.end_headers()
            self.wfile.write(b"OK")
        except Exception as e:
            logger.error(f"Error processing event: {e}")
            self.send_response(500)
            self.send_header("Content-type", "text/plain")
            self.end_headers()
            self.wfile.write(f"Error: {e}".encode("utf-8"))

    def process_event(self, event_type: str, payload: Dict[str, Any]) -> None:
        """Process a webhook event.

        Args:
            event_type: The type of event.
            payload: The event payload.
        """
        # Load the configuration
        config = load_config()
        
        # Initialize clients
        codegen_client = CodegenClient(config["codegen_org_id"], config["codegen_api_token"])
        github_client = GitHubClient(config["github_token"])
        
        # Process the event based on its type
        if event_type == "pull_request":
            self.process_pull_request_event(payload, codegen_client, github_client, config)
        elif event_type == "check_suite":
            self.process_check_suite_event(payload, codegen_client, github_client, config)
        elif event_type == "push":
            self.process_push_event(payload, codegen_client, github_client, config)

    def process_pull_request_event(
        self, 
        payload: Dict[str, Any], 
        codegen_client: CodegenClient, 
        github_client: GitHubClient,
        config: Dict[str, Any]
    ) -> None:
        """Process a pull request event.

        Args:
            payload: The event payload.
            codegen_client: The Codegen client.
            github_client: The GitHub client.
            config: The configuration.
        """
        # Extract information from the payload
        action = payload.get("action")
        pr = payload.get("pull_request", {})
        pr_number = pr.get("number")
        repo = payload.get("repository", {})
        repo_owner = repo.get("owner", {}).get("login")
        repo_name = repo.get("name")
        
        logger.info(f"Processing PR #{pr_number} ({action}) for {repo_owner}/{repo_name}")
        
        # Only process opened or synchronized PRs
        if action not in ["opened", "synchronize"]:
            logger.info(f"Ignoring PR {action} event")
            return
            
        # Step 2: Review PR against requirements
        prompt = f"""
        Analyze PR #{pr_number} in the repository {repo_owner}/{repo_name} against the requirements in REQUIREMENTS.md.
        
        If the PR fully implements the requirements, approve it and merge it.
        If the PR needs adjustments, create a new PR with the necessary changes.
        
        PR URL: {pr.get('html_url')}
        """
        
        task = codegen_client.run_agent(prompt)
        
        if task.get("status") == "completed":
            logger.info(f"PR review completed: {task.get('result', {}).get('summary')}")
        else:
            logger.error(f"PR review failed: {task.get('error')}")

    def process_check_suite_event(
        self, 
        payload: Dict[str, Any], 
        codegen_client: CodegenClient, 
        github_client: GitHubClient,
        config: Dict[str, Any]
    ) -> None:
        """Process a check suite event.

        Args:
            payload: The event payload.
            codegen_client: The Codegen client.
            github_client: The GitHub client.
            config: The configuration.
        """
        # Extract information from the payload
        action = payload.get("action")
        check_suite = payload.get("check_suite", {})
        conclusion = check_suite.get("conclusion")
        repo = payload.get("repository", {})
        repo_owner = repo.get("owner", {}).get("login")
        repo_name = repo.get("name")
        
        logger.info(f"Processing check suite ({action}, {conclusion}) for {repo_owner}/{repo_name}")
        
        # Only process completed check suites
        if action != "completed":
            logger.info(f"Ignoring check suite {action} event")
            return
            
        # Only process failed check suites
        if conclusion != "failure":
            logger.info(f"Ignoring check suite with conclusion {conclusion}")
            return
            
        # Get the pull requests associated with the check suite
        pull_requests = check_suite.get("pull_requests", [])
        if not pull_requests:
            logger.info("No pull requests associated with the check suite")
            return
            
        # Process the first pull request
        pr = pull_requests[0]
        pr_number = pr.get("number")
        
        # Step 3: Fix failed checks
        prompt = f"""
        The check suite for PR #{pr_number} in the repository {repo_owner}/{repo_name} has failed.
        
        Please fix the issues and update the PR.
        
        Check suite URL: {check_suite.get('url')}
        PR URL: {pr.get('url')}
        """
        
        task = codegen_client.run_agent(prompt)
        
        if task.get("status") == "completed":
            logger.info(f"Check suite fix completed: {task.get('result', {}).get('summary')}")
        else:
            logger.error(f"Check suite fix failed: {task.get('error')}")

    def process_push_event(
        self, 
        payload: Dict[str, Any], 
        codegen_client: CodegenClient, 
        github_client: GitHubClient,
        config: Dict[str, Any]
    ) -> None:
        """Process a push event.

        Args:
            payload: The event payload.
            codegen_client: The Codegen client.
            github_client: The GitHub client.
            config: The configuration.
        """
        # Extract information from the payload
        ref = payload.get("ref")
        repo = payload.get("repository", {})
        repo_owner = repo.get("owner", {}).get("login")
        repo_name = repo.get("name")
        
        logger.info(f"Processing push to {ref} for {repo_owner}/{repo_name}")
        
        # Only process pushes to the main branch
        if ref != "refs/heads/main":
            logger.info(f"Ignoring push to {ref}")
            return
            
        # Check if REQUIREMENTS.md was modified
        try:
            requirements = github_client.get_file_content(repo_owner, repo_name, "REQUIREMENTS.md")
            
            # Step 1: Analyze requirements and create PR
            prompt = f"""
            Analyze the REQUIREMENTS.md file in the repository {repo_owner}/{repo_name} and create a PR implementing the requirements.
            
            Requirements:
            {requirements}
            """
            
            task = codegen_client.run_agent(prompt)
            
            if task.get("status") == "completed":
                logger.info(f"Requirements analysis completed: {task.get('result', {}).get('summary')}")
            else:
                logger.error(f"Requirements analysis failed: {task.get('error')}")
                
        except Exception as e:
            logger.error(f"Error getting REQUIREMENTS.md: {e}")


class WebhookServer:
    """Server for handling GitHub webhook events."""

    def __init__(self, port: int = DEFAULT_PORT):
        """Initialize the webhook server.

        Args:
            port: The port to listen on.
        """
        self.port = port
        self.server = HTTPServer(("", port), WebhookHandler)
        self.thread = None
        logger.info(f"Initialized webhook server on port {port}")

    def start(self) -> None:
        """Start the webhook server in a separate thread."""
        logger.info("Starting webhook server")
        
        self.thread = threading.Thread(target=self.server.serve_forever)
        self.thread.daemon = True
        self.thread.start()
        
        logger.info(f"Webhook server running on port {self.port}")

    def stop(self) -> None:
        """Stop the webhook server."""
        logger.info("Stopping webhook server")
        
        if self.server:
            self.server.shutdown()
            self.server.server_close()
            
        if self.thread:
            self.thread.join()
            
        logger.info("Webhook server stopped")


def load_config() -> Dict[str, Any]:
    """Load the configuration from the config file.

    Returns:
        The configuration.
    """
    if not os.path.exists(CONFIG_FILE):
        return {}
        
    with open(CONFIG_FILE, "r") as f:
        return json.load(f)


def save_config(config: Dict[str, Any]) -> None:
    """Save the configuration to the config file.

    Args:
        config: The configuration to save.
    """
    with open(CONFIG_FILE, "w") as f:
        json.dump(config, f, indent=2)


def setup_workflow() -> None:
    """Set up the CI/CD workflow."""
    logger.info("Setting up CI/CD workflow")
    
    # Load existing configuration
    config = load_config()
    
    # Get GitHub token
    github_token = config.get("github_token") or input("GitHub API token: ")
    
    # Initialize GitHub client
    github_client = GitHubClient(github_token)
    
    # Get repositories
    repos = github_client.get_repositories()
    print("Available repositories:")
    for i, repo in enumerate(repos):
        print(f"{i+1}. {repo['full_name']}")
        
    # Select repository
    repo_index = int(input("Select repository (number): ")) - 1
    repo = repos[repo_index]
    repo_owner, repo_name = repo["full_name"].split("/")
    
    # Get Ngrok API key
    ngrok_api_key = config.get("ngrok_api_key") or input("Ngrok API key: ")
    
    # Initialize Ngrok client
    ngrok_client = NgrokClient(ngrok_api_key)
    
    # Create tunnel
    tunnel_url = ngrok_client.create_tunnel(DEFAULT_PORT)
    
    # Create webhook
    webhook = github_client.create_webhook(repo_owner, repo_name, tunnel_url)
    
    # Get Codegen API token and org ID
    codegen_api_token = config.get("codegen_api_token") or input("Codegen API token: ")
    codegen_org_id = config.get("codegen_org_id") or input("Codegen organization ID: ")
    
    # Save configuration
    config.update({
        "github_token": github_token,
        "ngrok_api_key": ngrok_api_key,
        "repo_owner": repo_owner,
        "repo_name": repo_name,
        "webhook_id": webhook["id"],
        "tunnel_url": tunnel_url,
        "codegen_api_token": codegen_api_token,
        "codegen_org_id": codegen_org_id,
    })
    save_config(config)
    
    logger.info("CI/CD workflow set up successfully")
    print(f"Webhook URL: {tunnel_url}")
    print(f"Configuration saved to {CONFIG_FILE}")


def run_workflow() -> None:
    """Run the CI/CD workflow."""
    logger.info("Running CI/CD workflow")
    
    # Load configuration
    config = load_config()
    if not config:
        logger.error("No configuration found. Please run setup first.")
        return
        
    # Start webhook server
    server = WebhookServer()
    server.start()
    
    # Initialize clients
    codegen_client = CodegenClient(config["codegen_org_id"], config["codegen_api_token"])
    github_client = GitHubClient(config["github_token"])
    
    # Trigger initial requirements analysis
    try:
        requirements = github_client.get_file_content(
            config["repo_owner"], config["repo_name"], "REQUIREMENTS.md"
        )
        
        # Step 1: Analyze requirements and create PR
        prompt = f"""
        Analyze the REQUIREMENTS.md file in the repository {config["repo_owner"]}/{config["repo_name"]} and create a PR implementing the requirements.
        
        Requirements:
        {requirements}
        """
        
        task = codegen_client.run_agent(prompt)
        
        if task.get("status") == "completed":
            logger.info(f"Requirements analysis completed: {task.get('result', {}).get('summary')}")
        else:
            logger.error(f"Requirements analysis failed: {task.get('error')}")
            
    except Exception as e:
        logger.error(f"Error getting REQUIREMENTS.md: {e}")
    
    # Keep the server running until interrupted
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        logger.info("Interrupted by user")
    finally:
        server.stop()


def teardown_workflow() -> None:
    """Tear down the CI/CD workflow."""
    logger.info("Tearing down CI/CD workflow")
    
    # Load configuration
    config = load_config()
    if not config:
        logger.error("No configuration found. Nothing to tear down.")
        return
        
    # Initialize clients
    github_client = GitHubClient(config["github_token"])
    
    # Delete webhook
    try:
        github_client.delete_webhook(
            config["repo_owner"], config["repo_name"], config["webhook_id"]
        )
    except Exception as e:
        logger.error(f"Error deleting webhook: {e}")
    
    # Delete configuration
    try:
        os.remove(CONFIG_FILE)
    except Exception as e:
        logger.error(f"Error deleting configuration: {e}")
    
    logger.info("CI/CD workflow torn down successfully")


def main() -> None:
    """Main entry point."""
    parser = argparse.ArgumentParser(description="Codegen CI/CD Workflow")
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--setup", action="store_true", help="Set up the CI/CD workflow")
    group.add_argument("--run", action="store_true", help="Run the CI/CD workflow")
    group.add_argument("--teardown", action="store_true", help="Tear down the CI/CD workflow")
    
    args = parser.parse_args()
    
    if args.setup:
        setup_workflow()
    elif args.run:
        run_workflow()
    elif args.teardown:
        teardown_workflow()


if __name__ == "__main__":
    main()

