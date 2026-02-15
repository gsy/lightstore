#!/usr/bin/env python3
"""Export trained YOLOv8 model to production format."""

import argparse
import shutil
import sys
from pathlib import Path

from ultralytics import YOLO


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Export YOLOv8 model for deployment")

    parser.add_argument(
        "--model",
        type=str,
        required=True,
        help="Path to trained model (best.pt)",
    )
    parser.add_argument(
        "--output-dir",
        type=str,
        default="../models",
        help="Output directory for exported model",
    )
    parser.add_argument(
        "--format",
        type=str,
        default="pytorch",
        choices=["pytorch", "onnx", "tflite", "torchscript"],
        help="Export format",
    )
    parser.add_argument("--imgsz", type=int, default=640, help="Image size")
    parser.add_argument("--half", action="store_true", help="FP16 quantization")
    parser.add_argument("--int8", action="store_true", help="INT8 quantization (requires calibration)")
    parser.add_argument("--simplify", action="store_true", help="Simplify ONNX model")
    parser.add_argument("--opset", type=int, default=17, help="ONNX opset version")

    return parser.parse_args()


def main() -> int:
    args = parse_args()

    print(f"Exporting model: {args.model}")
    print(f"  Format: {args.format}")
    print(f"  Output: {args.output_dir}")
    print(f"  Image size: {args.imgsz}")

    # Load model
    model = YOLO(args.model)

    # Create output directory
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    if args.format == "pytorch":
        # Just copy the model
        output_path = output_dir / "best.pt"
        shutil.copy(args.model, output_path)
        print(f"Model copied to: {output_path}")

    elif args.format == "onnx":
        exported_path = model.export(
            format="onnx",
            imgsz=args.imgsz,
            half=args.half,
            simplify=args.simplify,
            opset=args.opset,
        )
        output_path = output_dir / "model.onnx"
        shutil.move(exported_path, output_path)
        print(f"ONNX model exported to: {output_path}")

    elif args.format == "tflite":
        exported_path = model.export(
            format="tflite",
            imgsz=args.imgsz,
            half=args.half,
            int8=args.int8,
        )
        output_path = output_dir / "model.tflite"
        shutil.move(exported_path, output_path)
        print(f"TFLite model exported to: {output_path}")

    elif args.format == "torchscript":
        exported_path = model.export(
            format="torchscript",
            imgsz=args.imgsz,
        )
        output_path = output_dir / "model.torchscript"
        shutil.move(exported_path, output_path)
        print(f"TorchScript model exported to: {output_path}")

    # Print model info
    print("\nModel info:")
    print(f"  Classes: {len(model.names)}")
    print(f"  Names: {list(model.names.values())}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
