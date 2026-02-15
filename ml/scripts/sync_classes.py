#!/usr/bin/env python3
"""Sync class names from catalog DB to ML server and training config.

Fetches SKUs from the Go backend and updates:
1. ML server class mapping via gRPC
2. Training config YAML file
"""

import argparse
import os
import sys
from pathlib import Path

import grpc
import requests
import yaml


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Sync classes from catalog")

    parser.add_argument(
        "--catalog-url",
        type=str,
        default=os.getenv("CATALOG_URL", "http://localhost:8080"),
        help="Catalog service base URL",
    )
    parser.add_argument(
        "--ml-server",
        type=str,
        default=os.getenv("ML_SERVER", "localhost:50051"),
        help="ML server gRPC address",
    )
    parser.add_argument(
        "--config-path",
        type=str,
        default="training/configs/beverages.yaml",
        help="Training config YAML path",
    )
    parser.add_argument(
        "--update-config",
        action="store_true",
        help="Update training config file",
    )
    parser.add_argument(
        "--update-server",
        action="store_true",
        help="Update ML server class mapping",
    )

    return parser.parse_args()


def fetch_skus(catalog_url: str) -> list[dict]:
    """Fetch SKUs from catalog service."""
    resp = requests.get(f"{catalog_url}/api/v1/skus")
    resp.raise_for_status()
    return resp.json().get("skus", [])


def update_training_config(config_path: Path, skus: list[dict]) -> None:
    """Update training config with class names from SKUs."""
    # Load existing config
    if config_path.exists():
        with open(config_path) as f:
            config = yaml.safe_load(f)
    else:
        config = {}

    # Build class mapping
    # SKUs should have a 'class_id' field that maps to YOLO class index
    # If not, we assign based on order
    names = {}
    sku_mapping = {}

    for i, sku in enumerate(skus):
        class_id = sku.get("class_id", i)
        sku_id = sku.get("id", str(i))
        name = sku.get("name", f"beverage_{i}")

        names[class_id] = name
        sku_mapping[class_id] = sku_id

    config["nc"] = len(names)
    config["names"] = names
    config["sku_mapping"] = sku_mapping

    # Save config
    config_path.parent.mkdir(parents=True, exist_ok=True)
    with open(config_path, "w") as f:
        yaml.dump(config, f, default_flow_style=False, sort_keys=False)

    print(f"Updated config with {len(names)} classes")


def update_ml_server(ml_server: str, skus: list[dict]) -> None:
    """Update ML server class mapping via gRPC."""
    # Import generated code
    sys.path.insert(0, str(Path(__file__).parent.parent / "server"))
    from app.generated import detection_pb2, detection_pb2_grpc

    channel = grpc.insecure_channel(ml_server)
    stub = detection_pb2_grpc.DetectionServiceStub(channel)

    # Build class mappings
    classes = []
    for i, sku in enumerate(skus):
        class_id = sku.get("class_id", i)
        classes.append(
            detection_pb2.ClassMapping(
                class_id=class_id,
                sku_id=sku.get("id", str(i)),
                class_name=sku.get("name", f"beverage_{i}"),
            )
        )

    request = detection_pb2.SyncClassesRequest(classes=classes)

    try:
        response = stub.SyncClasses(request)
        if response.success:
            print(f"ML server updated with {response.class_count} classes")
        else:
            print("ML server sync failed")
    except grpc.RpcError as e:
        print(f"gRPC error: {e.details()}")
        raise


def main() -> int:
    args = parse_args()

    if not args.update_config and not args.update_server:
        print("Error: Specify --update-config and/or --update-server")
        return 1

    # Fetch SKUs
    print(f"Fetching SKUs from {args.catalog_url}...")
    try:
        skus = fetch_skus(args.catalog_url)
    except requests.exceptions.RequestException as e:
        print(f"Failed to fetch SKUs: {e}")
        return 1

    print(f"Found {len(skus)} SKUs")

    if not skus:
        print("Warning: No SKUs found")
        return 0

    # Update training config
    if args.update_config:
        print(f"\nUpdating training config: {args.config_path}")
        update_training_config(Path(args.config_path), skus)

    # Update ML server
    if args.update_server:
        print(f"\nUpdating ML server: {args.ml_server}")
        try:
            update_ml_server(args.ml_server, skus)
        except Exception as e:
            print(f"Failed to update ML server: {e}")
            return 1

    print("\nSync complete!")
    return 0


if __name__ == "__main__":
    sys.exit(main())
