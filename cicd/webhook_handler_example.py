#!/usr/bin/env python3
"""
Example Webhook Handler for Codegen CI/CD Workflow

This script demonstrates how to handle GitHub webhook events and process them
using the Codegen API. It's a simplified version of the webhook handler in
the main codegen_cicd.py script.
"""

import json
import logging
import sys
from http.server import HTTPServer, BaseHTTPRequestHandler
from typing import Dict, Any

import requests

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler("webhook_handler.log"),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger("webhook_handler")

# Configuration
CONFIG = {
    "github_token": "your-github-token",
    "codegen_api_token": "your-codegen-api-token",
    "codegen_org_id": "your-codegen-org-id",
    "repo_owner": "repository-owner",
    "repo_name": "repository-name",
}

# Constants
PORT = 8000


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

    def run_agent(self, prompt: str) -> Dict[str, Any]:
        """Run a Codegen agent with the given prompt.

        Args:
            prompt: The prompt to send to the agent.

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
            import time
            time.sleep(5)
        
        return task


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
        # Initialize Codegen client
        codegen_client = CodegenClient(CONFIG["codegen_org_id"], CONFIG["codegen_api_token"])
        
        # Process the event based on its type
        if event_type == "pull_request":
            self.process_pull_request_event(payload, codegen_client)
        elif event_type == "check_suite":
            self.process_check_suite_event(payload, codegen_client)
        elif event_type == "push":
            self.process_push_event(payload, codegen_client)

    def process_pull_request_event(
        self, 
        payload: Dict[str, Any], 
        codegen_client: CodegenClient
    ) -> None:
        """Process a pull request event.

        Args:
            payload: The event payload.
            codegen_client: The Codegen client.
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
        codegen_client: CodegenClient
    ) -> None:
        """Process a check suite event.

        Args:
            payload: The event payload.
            codegen_client: The Codegen client.
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
        codegen_client: CodegenClient
    ) -> None:
        """Process a push event.

        Args:
            payload: The event payload.
            codegen_client: The Codegen client.
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
            
        # Step 1: Analyze requirements and create PR
        prompt = f"""
        Analyze the REQUIREMENTS.md file in the repository {repo_owner}/{repo_name} and create a PR implementing the requirements.
        
        Repository: {repo_owner}/{repo_name}
        """
        
        task = codegen_client.run_agent(prompt)
        
        if task.get("status") == "completed":
            logger.info(f"Requirements analysis completed: {task.get('result', {}).get('summary')}")
        else:
            logger.error(f"Requirements analysis failed: {task.get('error')}")


def main() -> None:
    """Main entry point."""
    logger.info(f"Starting webhook server on port {PORT}")
    
    server = HTTPServer(("", PORT), WebhookHandler)
    
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        logger.info("Interrupted by user")
    finally:
        server.server_close()
        logger.info("Webhook server stopped")


if __name__ == "__main__":
    main()

