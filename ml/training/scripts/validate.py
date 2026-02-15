#!/usr/bin/env python3
"""Validate trained YOLOv8 model."""

import argparse
import json
import sys
from pathlib import Path

from ultralytics import YOLO


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Validate YOLOv8 beverage detection model")

    parser.add_argument(
        "--model",
        type=str,
        required=True,
        help="Path to trained model (best.pt)",
    )
    parser.add_argument(
        "--data",
        type=str,
        default="configs/beverages.yaml",
        help="Dataset config file",
    )
    parser.add_argument("--imgsz", type=int, default=640, help="Image size")
    parser.add_argument("--batch", type=int, default=16, help="Batch size")
    parser.add_argument("--device", type=str, default="0", help="Device")
    parser.add_argument("--split", type=str, default="val", help="Dataset split (val, test)")
    parser.add_argument(
        "--output",
        type=str,
        default=None,
        help="Output JSON file for metrics",
    )
    parser.add_argument("--conf", type=float, default=0.001, help="Confidence threshold")
    parser.add_argument("--iou", type=float, default=0.6, help="IoU threshold for NMS")
    parser.add_argument("--plots", action="store_true", help="Generate plots")

    return parser.parse_args()


def main() -> int:
    args = parse_args()

    print(f"Validating model: {args.model}")
    print(f"  Dataset: {args.data}")
    print(f"  Split: {args.split}")
    print(f"  Image size: {args.imgsz}")

    # Load model
    model = YOLO(args.model)

    # Run validation
    metrics = model.val(
        data=args.data,
        imgsz=args.imgsz,
        batch=args.batch,
        device=args.device,
        split=args.split,
        conf=args.conf,
        iou=args.iou,
        plots=args.plots,
    )

    # Print results
    print("\n" + "=" * 50)
    print("Validation Results")
    print("=" * 50)
    print(f"  Precision: {metrics.box.mp:.4f}")
    print(f"  Recall: {metrics.box.mr:.4f}")
    print(f"  mAP50: {metrics.box.map50:.4f}")
    print(f"  mAP50-95: {metrics.box.map:.4f}")

    # Per-class metrics
    if hasattr(metrics.box, "ap_class_index") and metrics.box.ap_class_index is not None:
        print("\nPer-class mAP50:")
        names = model.names
        for i, class_idx in enumerate(metrics.box.ap_class_index):
            class_name = names.get(int(class_idx), f"class_{class_idx}")
            ap50 = metrics.box.ap50[i] if i < len(metrics.box.ap50) else 0
            print(f"  {class_name}: {ap50:.4f}")

    # Save metrics to JSON
    if args.output:
        output_data = {
            "model": args.model,
            "dataset": args.data,
            "split": args.split,
            "metrics": {
                "precision": float(metrics.box.mp),
                "recall": float(metrics.box.mr),
                "mAP50": float(metrics.box.map50),
                "mAP50_95": float(metrics.box.map),
            },
        }

        output_path = Path(args.output)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, "w") as f:
            json.dump(output_data, f, indent=2)
        print(f"\nMetrics saved to: {args.output}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
