#!/usr/bin/env python3
"""Export annotations from CVAT via API.

Requires CVAT API token. Set via environment variable CVAT_API_TOKEN.
"""

import argparse
import os
import sys
import time
import zipfile
from io import BytesIO
from pathlib import Path

import requests


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Export annotations from CVAT")

    parser.add_argument(
        "--cvat-url",
        type=str,
        default=os.getenv("CVAT_URL", "http://localhost:8080"),
        help="CVAT server URL",
    )
    parser.add_argument(
        "--project-id",
        type=int,
        required=True,
        help="CVAT project ID",
    )
    parser.add_argument(
        "--output-dir",
        type=str,
        default="data/annotations",
        help="Output directory for annotations",
    )
    parser.add_argument(
        "--format",
        type=str,
        default="YOLO 1.1",
        help="Export format (YOLO 1.1, COCO 1.0, etc.)",
    )
    parser.add_argument(
        "--include-images",
        action="store_true",
        help="Include images in export",
    )

    return parser.parse_args()


class CVATClient:
    """Simple CVAT API client."""

    def __init__(self, base_url: str, token: str):
        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()
        self.session.headers["Authorization"] = f"Token {token}"

    def get_project(self, project_id: int) -> dict:
        """Get project details."""
        resp = self.session.get(f"{self.base_url}/api/projects/{project_id}")
        resp.raise_for_status()
        return resp.json()

    def export_project(
        self,
        project_id: int,
        export_format: str = "YOLO 1.1",
        include_images: bool = False,
    ) -> bytes:
        """Export project annotations."""
        # Start export task
        params = {
            "format": export_format,
            "save_images": str(include_images).lower(),
        }
        resp = self.session.get(
            f"{self.base_url}/api/projects/{project_id}/dataset",
            params=params,
        )

        # If 202, export is processing
        if resp.status_code == 202:
            print("Export started, waiting for completion...")
            while True:
                time.sleep(2)
                resp = self.session.get(
                    f"{self.base_url}/api/projects/{project_id}/dataset",
                    params=params,
                )
                if resp.status_code == 200:
                    break
                elif resp.status_code != 202:
                    resp.raise_for_status()
                print(".", end="", flush=True)
            print()

        resp.raise_for_status()
        return resp.content

    def list_tasks(self, project_id: int) -> list[dict]:
        """List tasks in project."""
        resp = self.session.get(
            f"{self.base_url}/api/tasks",
            params={"project_id": project_id},
        )
        resp.raise_for_status()
        return resp.json().get("results", [])


def main() -> int:
    args = parse_args()

    # Get API token
    token = os.getenv("CVAT_API_TOKEN")
    if not token:
        print("Error: CVAT_API_TOKEN environment variable not set")
        print("\nTo get a token:")
        print("  1. Log into CVAT")
        print("  2. Go to your profile settings")
        print("  3. Generate an API token")
        print("  4. export CVAT_API_TOKEN=<your-token>")
        return 1

    client = CVATClient(args.cvat_url, token)
    output_dir = Path(args.output_dir)

    try:
        # Get project info
        print(f"Fetching project {args.project_id}...")
        project = client.get_project(args.project_id)
        print(f"Project: {project['name']}")

        # List tasks
        tasks = client.list_tasks(args.project_id)
        print(f"Found {len(tasks)} tasks")

        # Export annotations
        print(f"\nExporting in format: {args.format}")
        export_data = client.export_project(
            args.project_id,
            args.format,
            args.include_images,
        )

        # Extract zip
        output_dir.mkdir(parents=True, exist_ok=True)

        with zipfile.ZipFile(BytesIO(export_data)) as zf:
            zf.extractall(output_dir)

        print(f"\nAnnotations exported to: {output_dir}")

        # List exported files
        exported_files = list(output_dir.glob("*"))
        print(f"Exported {len(exported_files)} files")

        return 0

    except requests.exceptions.RequestException as e:
        print(f"API error: {e}")
        return 1


if __name__ == "__main__":
    sys.exit(main())
