#!/usr/bin/env python3
"""
Test script to verify execution logs functionality.

This script will:
1. Create a new test workflow with multiple nodes
2. Execute the workflow
3. Wait for execution to complete
4. Fetch execution logs from ClickHouse
5. Verify logs are present and correct
"""

import requests
import time
import json
import sys
from datetime import datetime

# Configuration
BASE_URL = "http://localhost:3001/api"
TIMEOUT = 30  # seconds

# Colors for output
class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    RESET = '\033[0m'
    BOLD = '\033[1m'

def print_step(step, message):
    """Print a formatted step message."""
    print(f"\n{Colors.BLUE}{Colors.BOLD}[Step {step}]{Colors.RESET} {message}")

def print_success(message):
    """Print a success message."""
    print(f"{Colors.GREEN}✓ {message}{Colors.RESET}")

def print_error(message):
    """Print an error message."""
    print(f"{Colors.RED}✗ {message}{Colors.RESET}")

def print_info(message):
    """Print an info message."""
    print(f"{Colors.YELLOW}ℹ {message}{Colors.RESET}")

def create_test_workflow():
    """Create a test workflow with multiple nodes."""
    print_step(1, "Creating test workflow...")

    workflow = {
        "name": f"Test Workflow - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
        "description": "Automated test workflow to verify execution logs",
        "nodes": [
            {
                "id": "trigger-1",
                "type": "chat_trigger",
                "position": {"x": 100, "y": 100},
                "data": {
                    "label": "Chat Trigger"
                }
            },
            {
                "id": "transform-1",
                "type": "transform",
                "position": {"x": 300, "y": 100},
                "data": {
                    "label": "Transform Data",
                    "expression": "{ message: input.message, timestamp: new Date().toISOString() }"
                }
            },
            {
                "id": "delay-1",
                "type": "delay",
                "position": {"x": 500, "y": 100},
                "data": {
                    "label": "Delay 1s",
                    "duration": 1000
                }
            },
            {
                "id": "response-1",
                "type": "response",
                "position": {"x": 700, "y": 100},
                "data": {
                    "label": "Send Response",
                    "message": "Test completed successfully"
                }
            }
        ],
        "edges": [
            {
                "id": "e1",
                "source": "trigger-1",
                "target": "transform-1"
            },
            {
                "id": "e2",
                "source": "transform-1",
                "target": "delay-1"
            },
            {
                "id": "e3",
                "source": "delay-1",
                "target": "response-1"
            }
        ],
        "active": True
    }

    try:
        response = requests.post(f"{BASE_URL}/workflows", json=workflow, timeout=10)
        response.raise_for_status()

        workflow_data = response.json()
        workflow_id = workflow_data["data"]["id"]

        print_success(f"Created workflow with ID: {workflow_id}")
        print_info(f"Workflow name: {workflow['name']}")
        print_info(f"Nodes: {len(workflow['nodes'])}")

        return workflow_id
    except requests.exceptions.RequestException as e:
        print_error(f"Failed to create workflow: {e}")
        sys.exit(1)

def execute_workflow(workflow_id):
    """Execute the workflow and return execution ID."""
    print_step(2, "Executing workflow...")

    context = {
        "message": "This is a test execution",
        "testData": {
            "value": 123,
            "timestamp": datetime.now().isoformat()
        }
    }

    try:
        response = requests.post(
            f"{BASE_URL}/workflows/{workflow_id}/execute",
            json={"context": context},
            timeout=10
        )
        response.raise_for_status()

        execution_data = response.json()
        execution_id = execution_data["data"]["id"]

        print_success(f"Started execution with ID: {execution_id}")
        print_info(f"Status: {execution_data['data']['status']}")

        return execution_id
    except requests.exceptions.RequestException as e:
        print_error(f"Failed to execute workflow: {e}")
        sys.exit(1)

def wait_for_execution(workflow_id, execution_id):
    """Wait for execution to complete."""
    print_step(3, "Waiting for execution to complete...")

    start_time = time.time()

    while time.time() - start_time < TIMEOUT:
        try:
            response = requests.get(
                f"{BASE_URL}/workflows/{workflow_id}/executions/{execution_id}",
                timeout=5
            )
            response.raise_for_status()

            execution_data = response.json()["data"]["execution"]
            status = execution_data["status"]

            print(f"  Status: {status}", end='\r')

            if status in ["success", "error"]:
                print()  # New line
                if status == "success":
                    print_success(f"Execution completed successfully")
                else:
                    print_error(f"Execution failed: {execution_data.get('error', 'Unknown error')}")

                if execution_data.get("finishedAt"):
                    try:
                        started = datetime.fromisoformat(execution_data["startedAt"].replace('Z', '+00:00'))
                        finished = datetime.fromisoformat(execution_data["finishedAt"].replace('Z', '+00:00'))
                        duration = (finished - started).total_seconds()
                        print_info(f"Duration: {duration:.2f} seconds")
                    except ValueError:
                        # Handle Go's time format which may have different microsecond precision
                        pass

                return status

            time.sleep(1)
        except requests.exceptions.RequestException as e:
            print_error(f"Failed to check execution status: {e}")
            sys.exit(1)

    print_error(f"Execution timed out after {TIMEOUT} seconds")
    return "timeout"

def fetch_execution_logs(execution_id):
    """Fetch execution logs from ClickHouse."""
    print_step(4, "Fetching execution logs from ClickHouse...")

    try:
        # Fetch execution logs
        response = requests.get(
            f"{BASE_URL}/logs/executions",
            params={"execution_id": execution_id, "limit": 100},
            timeout=10
        )
        response.raise_for_status()

        logs = response.json()["data"]

        if not logs:
            print_error("No execution logs found!")
            return False

        print_success(f"Found {len(logs)} execution log entries")

        # Group logs by level (handle both PascalCase and snake_case)
        log_levels = {}
        for log in logs:
            level = log.get("level") or log.get("Level", "UNKNOWN")
            log_levels[level] = log_levels.get(level, 0) + 1

        print_info("Log levels breakdown:")
        for level, count in sorted(log_levels.items()):
            print(f"  - {level}: {count}")

        # Print sample logs (handle both PascalCase and snake_case)
        print_info("\nSample log entries:")
        for i, log in enumerate(logs[:5]):
            level = log.get('level') or log.get('Level', 'UNKNOWN')
            message = log.get('message') or log.get('Message', '(no message)')
            print(f"  {i+1}. [{level}] {message[:80]}")
            node_id = log.get('node_id') or log.get('NodeID')
            if node_id:
                print(f"     Node: {node_id}")
            duration = log.get('duration_ms') or log.get('DurationMs')
            if duration:
                print(f"     Duration: {duration}ms")

        # Debug: print first log structure
        if logs:
            print_info("\nFirst log entry structure:")
            for key in logs[0].keys():
                print(f"  - {key}: {type(logs[0][key]).__name__}")

        return True
    except requests.exceptions.RequestException as e:
        print_error(f"Failed to fetch execution logs: {e}")
        return False

def fetch_node_executions(execution_id):
    """Fetch node execution logs."""
    print_step(5, "Fetching node execution logs...")

    try:
        response = requests.get(
            f"{BASE_URL}/logs/nodes/{execution_id}",
            timeout=10
        )
        response.raise_for_status()

        node_logs = response.json()["data"]

        if not node_logs:
            print_error("No node execution logs found!")
            return False

        print_success(f"Found {len(node_logs)} node execution records")

        # Analyze node executions (handle both PascalCase and snake_case)
        successful = sum(1 for n in node_logs if (n.get("status") or n.get("Status")) == "success")
        failed = sum(1 for n in node_logs if (n.get("status") or n.get("Status")) == "error")

        print_info(f"Node execution results:")
        print(f"  - Successful: {successful}")
        print(f"  - Failed: {failed}")

        print_info("\nNode execution details:")
        for node in node_logs:
            status = node.get("status") or node.get("Status", "unknown")
            node_id = node.get("node_id") or node.get("NodeID", "unknown")
            node_type = node.get("node_type") or node.get("NodeType", "unknown")
            duration_ms = node.get('duration_ms') or node.get('DurationMs', 0)
            error = node.get('error') or node.get('Error')

            status_icon = "✓" if status == "success" else "✗"
            duration = f"{duration_ms}ms"
            print(f"  {status_icon} {node_id} ({node_type}) - {duration}")
            if error:
                print(f"    Error: {error}")

        return True
    except requests.exceptions.RequestException as e:
        print_error(f"Failed to fetch node executions: {e}")
        return False

def fetch_workflow_execution_history(workflow_id):
    """Fetch workflow execution history."""
    print_step(6, "Fetching workflow execution history...")

    try:
        response = requests.get(
            f"{BASE_URL}/logs/workflows/{workflow_id}",
            timeout=10
        )
        response.raise_for_status()

        history = response.json()["data"]

        if not history:
            print_error("No workflow execution history found!")
            return False

        print_success(f"Found {len(history)} workflow execution records")

        print_info("\nWorkflow execution summary:")
        for i, exec_summary in enumerate(history[:3]):
            exec_id = exec_summary.get('execution_id') or exec_summary.get('ExecutionID', 'unknown')
            status = exec_summary.get('status') or exec_summary.get('Status', 'unknown')
            total = exec_summary.get('total_nodes') or exec_summary.get('TotalNodes', 0)
            success = exec_summary.get('successful_nodes') or exec_summary.get('SuccessfulNodes', 0)
            failed = exec_summary.get('failed_nodes') or exec_summary.get('FailedNodes', 0)
            duration = exec_summary.get('duration_ms') or exec_summary.get('DurationMs')

            print(f"  {i+1}. Execution {exec_id[:8]}...")
            print(f"     Status: {status}")
            print(f"     Total nodes: {total}")
            print(f"     Success/Failed: {success}/{failed}")
            if duration:
                print(f"     Duration: {duration}ms")

        return True
    except requests.exceptions.RequestException as e:
        print_error(f"Failed to fetch workflow history: {e}")
        return False

def cleanup_workflow(workflow_id):
    """Delete the test workflow."""
    print_step(7, "Cleaning up test workflow...")

    try:
        response = requests.delete(f"{BASE_URL}/workflows/{workflow_id}", timeout=10)
        response.raise_for_status()
        print_success("Test workflow deleted successfully")
    except requests.exceptions.RequestException as e:
        print_error(f"Failed to delete workflow: {e}")

def main():
    """Main test execution."""
    print(f"\n{Colors.BOLD}{'='*60}{Colors.RESET}")
    print(f"{Colors.BOLD}Workflow Execution Logs Test Suite{Colors.RESET}")
    print(f"{Colors.BOLD}{'='*60}{Colors.RESET}")

    start_time = time.time()

    # Run tests
    workflow_id = create_test_workflow()
    execution_id = execute_workflow(workflow_id)
    status = wait_for_execution(workflow_id, execution_id)

    # Verify logs
    logs_ok = fetch_execution_logs(execution_id)
    nodes_ok = fetch_node_executions(execution_id)
    history_ok = fetch_workflow_execution_history(workflow_id)

    # Cleanup
    cleanup_workflow(workflow_id)

    # Summary
    total_time = time.time() - start_time

    print(f"\n{Colors.BOLD}{'='*60}{Colors.RESET}")
    print(f"{Colors.BOLD}Test Results Summary{Colors.RESET}")
    print(f"{Colors.BOLD}{'='*60}{Colors.RESET}")

    tests = [
        ("Workflow Creation", True),
        ("Workflow Execution", status == "success"),
        ("Execution Logs (ClickHouse)", logs_ok),
        ("Node Execution Logs", nodes_ok),
        ("Workflow History", history_ok)
    ]

    passed = sum(1 for _, result in tests if result)
    total = len(tests)

    for test_name, result in tests:
        icon = "✓" if result else "✗"
        color = Colors.GREEN if result else Colors.RED
        print(f"{color}{icon} {test_name}{Colors.RESET}")

    print(f"\n{Colors.BOLD}Tests Passed: {passed}/{total}{Colors.RESET}")
    print(f"{Colors.BOLD}Total Time: {total_time:.2f}s{Colors.RESET}")

    if passed == total:
        print(f"\n{Colors.GREEN}{Colors.BOLD}🎉 All tests passed!{Colors.RESET}")
        sys.exit(0)
    else:
        print(f"\n{Colors.RED}{Colors.BOLD}❌ Some tests failed!{Colors.RESET}")
        sys.exit(1)

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print(f"\n{Colors.YELLOW}Test interrupted by user{Colors.RESET}")
        sys.exit(1)
    except Exception as e:
        print_error(f"Unexpected error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
