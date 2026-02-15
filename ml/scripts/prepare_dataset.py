#!/usr/bin/env python3
"""Prepare dataset for YOLOv8 training.

Converts CVAT annotations to YOLO format and creates train/val/test splits.
"""

import argparse
import random
import shutil
import sys
from pathlib import Path

import yaml


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Prepare dataset for training")

    parser.add_argument(
        "--images-dir",
        type=str,
        default="data/raw",
        help="Directory with raw images",
    )
    parser.add_argument(
        "--annotations-dir",
        type=str,
        default="data/annotations",
        help="Directory with CVAT YOLO export",
    )
    parser.add_argument(
        "--output-dir",
        type=str,
        default="data/dataset",
        help="Output dataset directory",
    )
    parser.add_argument(
        "--train-ratio",
        type=float,
        default=0.8,
        help="Training set ratio",
    )
    parser.add_argument(
        "--val-ratio",
        type=float,
        default=0.15,
        help="Validation set ratio",
    )
    parser.add_argument(
        "--seed",
        type=int,
        default=42,
        help="Random seed for reproducibility",
    )
    parser.add_argument(
        "--config-output",
        type=str,
        default="training/configs/beverages.yaml",
        help="Output YAML config path",
    )

    return parser.parse_args()


def find_image_label_pairs(
    images_dir: Path,
    annotations_dir: Path,
) -> list[tuple[Path, Path]]:
    """Find matching image and label file pairs."""
    pairs = []
    image_extensions = {".jpg", ".jpeg", ".png", ".bmp"}

    # Find all images
    for img_path in images_dir.iterdir():
        if img_path.suffix.lower() not in image_extensions:
            continue

        # Look for corresponding label file
        label_name = img_path.stem + ".txt"
        label_path = annotations_dir / label_name

        if label_path.exists():
            pairs.append((img_path, label_path))
        else:
            print(f"Warning: No label for {img_path.name}")

    return pairs


def split_dataset(
    pairs: list[tuple[Path, Path]],
    train_ratio: float,
    val_ratio: float,
    seed: int,
) -> tuple[list, list, list]:
    """Split dataset into train/val/test sets."""
    random.seed(seed)
    random.shuffle(pairs)

    n_total = len(pairs)
    n_train = int(n_total * train_ratio)
    n_val = int(n_total * val_ratio)

    train_pairs = pairs[:n_train]
    val_pairs = pairs[n_train : n_train + n_val]
    test_pairs = pairs[n_train + n_val :]

    return train_pairs, val_pairs, test_pairs


def copy_pairs_to_split(
    pairs: list[tuple[Path, Path]],
    output_dir: Path,
    split_name: str,
) -> None:
    """Copy image/label pairs to split directory."""
    images_dir = output_dir / split_name / "images"
    labels_dir = output_dir / split_name / "labels"

    images_dir.mkdir(parents=True, exist_ok=True)
    labels_dir.mkdir(parents=True, exist_ok=True)

    for img_path, label_path in pairs:
        shutil.copy(img_path, images_dir / img_path.name)
        shutil.copy(label_path, labels_dir / label_path.name)


def extract_classes_from_annotations(annotations_dir: Path) -> dict[int, str]:
    """Extract class names from CVAT obj.names file or label files."""
    # Try to find obj.names from CVAT export
    obj_names_path = annotations_dir / "obj.names"
    if obj_names_path.exists():
        with open(obj_names_path) as f:
            names = [line.strip() for line in f if line.strip()]
        return {i: name for i, name in enumerate(names)}

    # Fall back to extracting unique class IDs from label files
    class_ids = set()
    for label_file in annotations_dir.glob("*.txt"):
        with open(label_file) as f:
            for line in f:
                parts = line.strip().split()
                if parts:
                    class_ids.add(int(parts[0]))

    return {i: f"class_{i}" for i in sorted(class_ids)}


def create_dataset_config(
    output_path: Path,
    dataset_dir: Path,
    class_names: dict[int, str],
) -> None:
    """Create YOLO dataset configuration file."""
    config = {
        "path": str(dataset_dir.resolve()),
        "train": "train/images",
        "val": "val/images",
        "test": "test/images",
        "nc": len(class_names),
        "names": class_names,
    }

    output_path.parent.mkdir(parents=True, exist_ok=True)
    with open(output_path, "w") as f:
        yaml.dump(config, f, default_flow_style=False, sort_keys=False)


def main() -> int:
    args = parse_args()

    images_dir = Path(args.images_dir)
    annotations_dir = Path(args.annotations_dir)
    output_dir = Path(args.output_dir)
    config_output = Path(args.config_output)

    # Validate inputs
    if not images_dir.exists():
        print(f"Error: Images directory not found: {images_dir}")
        return 1

    if not annotations_dir.exists():
        print(f"Error: Annotations directory not found: {annotations_dir}")
        return 1

    # Find pairs
    print(f"Searching for images in: {images_dir}")
    print(f"Searching for annotations in: {annotations_dir}")

    pairs = find_image_label_pairs(images_dir, annotations_dir)
    print(f"Found {len(pairs)} image/label pairs")

    if not pairs:
        print("Error: No image/label pairs found")
        return 1

    # Split dataset
    train_pairs, val_pairs, test_pairs = split_dataset(
        pairs,
        args.train_ratio,
        args.val_ratio,
        args.seed,
    )
    print(f"Split: {len(train_pairs)} train, {len(val_pairs)} val, {len(test_pairs)} test")

    # Clean output directory
    if output_dir.exists():
        shutil.rmtree(output_dir)

    # Copy files
    print("Copying files...")
    copy_pairs_to_split(train_pairs, output_dir, "train")
    copy_pairs_to_split(val_pairs, output_dir, "val")
    copy_pairs_to_split(test_pairs, output_dir, "test")

    # Extract class names
    class_names = extract_classes_from_annotations(annotations_dir)
    print(f"Found {len(class_names)} classes: {list(class_names.values())}")

    # Create config
    create_dataset_config(config_output, output_dir, class_names)
    print(f"Dataset config saved to: {config_output}")

    print("\nDataset preparation complete!")
    return 0


if __name__ == "__main__":
    sys.exit(main())
