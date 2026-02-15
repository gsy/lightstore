#!/usr/bin/env python3
"""YOLOv8 training script for beverage detection."""

import argparse
import os
import sys
from datetime import datetime
from pathlib import Path

from ultralytics import YOLO


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Train YOLOv8 for beverage detection")

    # Model
    parser.add_argument(
        "--model",
        type=str,
        default="yolov8s.pt",
        help="Base model (yolov8n.pt, yolov8s.pt, yolov8m.pt)",
    )

    # Dataset
    parser.add_argument(
        "--data",
        type=str,
        default="configs/beverages.yaml",
        help="Dataset config file",
    )

    # Training params
    parser.add_argument("--epochs", type=int, default=100, help="Number of epochs")
    parser.add_argument("--batch", type=int, default=16, help="Batch size")
    parser.add_argument("--imgsz", type=int, default=640, help="Image size")
    parser.add_argument("--patience", type=int, default=20, help="Early stopping patience")

    # Hardware
    parser.add_argument("--device", type=str, default="0", help="CUDA device (0, 1, cpu)")
    parser.add_argument("--workers", type=int, default=8, help="Dataloader workers")

    # Output
    parser.add_argument("--project", type=str, default="runs/train", help="Project directory")
    parser.add_argument("--name", type=str, default=None, help="Experiment name")

    # Augmentation
    parser.add_argument("--augment", action="store_true", help="Enable extra augmentation")

    # Resume
    parser.add_argument("--resume", type=str, default=None, help="Resume from checkpoint")

    return parser.parse_args()


def main() -> int:
    args = parse_args()

    # Set experiment name
    if args.name is None:
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        args.name = f"beverages_{timestamp}"

    print(f"Starting training: {args.name}")
    print(f"  Model: {args.model}")
    print(f"  Data: {args.data}")
    print(f"  Epochs: {args.epochs}")
    print(f"  Batch size: {args.batch}")
    print(f"  Image size: {args.imgsz}")
    print(f"  Device: {args.device}")

    # Load model
    if args.resume:
        model = YOLO(args.resume)
        print(f"Resuming from: {args.resume}")
    else:
        model = YOLO(args.model)
        print(f"Loaded base model: {args.model}")

    # Augmentation settings
    augment_params = {}
    if args.augment:
        augment_params = {
            "hsv_h": 0.015,  # Hue augmentation
            "hsv_s": 0.7,  # Saturation augmentation
            "hsv_v": 0.4,  # Value augmentation
            "degrees": 10.0,  # Rotation
            "translate": 0.1,  # Translation
            "scale": 0.5,  # Scale
            "shear": 5.0,  # Shear
            "perspective": 0.0005,  # Perspective
            "flipud": 0.0,  # Vertical flip (disabled for beverages)
            "fliplr": 0.5,  # Horizontal flip
            "mosaic": 1.0,  # Mosaic augmentation
            "mixup": 0.1,  # Mixup augmentation
        }
        print("Extra augmentation enabled")

    # Train
    results = model.train(
        data=args.data,
        epochs=args.epochs,
        batch=args.batch,
        imgsz=args.imgsz,
        patience=args.patience,
        device=args.device,
        workers=args.workers,
        project=args.project,
        name=args.name,
        exist_ok=True,
        pretrained=True,
        optimizer="AdamW",
        lr0=0.001,
        lrf=0.01,
        momentum=0.937,
        weight_decay=0.0005,
        warmup_epochs=3,
        warmup_momentum=0.8,
        warmup_bias_lr=0.1,
        box=7.5,
        cls=0.5,
        dfl=1.5,
        close_mosaic=10,
        amp=True,  # Automatic mixed precision
        **augment_params,
    )

    # Print results
    print("\nTraining complete!")
    print(f"Best model saved to: {Path(args.project) / args.name / 'weights' / 'best.pt'}")

    # Validate
    print("\nRunning validation...")
    metrics = model.val()
    print(f"  mAP50: {metrics.box.map50:.4f}")
    print(f"  mAP50-95: {metrics.box.map:.4f}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
