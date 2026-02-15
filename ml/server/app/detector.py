"""YOLOv8 detector with hot reload support."""

import io
import logging
import threading
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Optional

import numpy as np
from PIL import Image
from ultralytics import YOLO

from .config import ModelConfig

logger = logging.getLogger(__name__)


@dataclass
class DetectionResult:
    """Single detection result."""

    class_id: int
    class_name: str
    confidence: float
    x1: float
    y1: float
    x2: float
    y2: float


@dataclass
class InferenceResult:
    """Complete inference result."""

    detections: list[DetectionResult]
    inference_time_ms: float
    model_version: str


class ModelWatcher:
    """Watches model file for changes and triggers reload."""

    def __init__(
        self,
        model_path: Path,
        on_change: callable,
        interval: float = 5.0,
    ):
        self.model_path = model_path
        self.on_change = on_change
        self.interval = interval
        self._last_mtime: Optional[float] = None
        self._running = False
        self._thread: Optional[threading.Thread] = None

    def start(self) -> None:
        """Start watching for model changes."""
        if self._running:
            return

        self._running = True
        self._last_mtime = self._get_mtime()
        self._thread = threading.Thread(target=self._watch_loop, daemon=True)
        self._thread.start()
        logger.info(f"Started watching model file: {self.model_path}")

    def stop(self) -> None:
        """Stop watching."""
        self._running = False
        if self._thread:
            self._thread.join(timeout=self.interval + 1)
        logger.info("Stopped model watcher")

    def _get_mtime(self) -> Optional[float]:
        """Get model file modification time."""
        try:
            return self.model_path.stat().st_mtime
        except FileNotFoundError:
            return None

    def _watch_loop(self) -> None:
        """Main watch loop."""
        while self._running:
            time.sleep(self.interval)
            current_mtime = self._get_mtime()

            if current_mtime is None:
                continue

            if self._last_mtime is None or current_mtime > self._last_mtime:
                logger.info("Model file changed, triggering reload...")
                self._last_mtime = current_mtime
                try:
                    self.on_change()
                except Exception as e:
                    logger.error(f"Error during model reload: {e}")


class Detector:
    """YOLOv8 detector with hot reload support."""

    def __init__(self, config: ModelConfig):
        self.config = config
        self._model: Optional[YOLO] = None
        self._model_lock = threading.RLock()
        self._model_version: str = "unknown"
        self._class_mapping: dict[int, tuple[str, str]] = {}  # class_id -> (sku_id, name)
        self._watcher: Optional[ModelWatcher] = None

    @property
    def model_loaded(self) -> bool:
        """Check if model is loaded."""
        with self._model_lock:
            return self._model is not None

    @property
    def model_version(self) -> str:
        """Get current model version."""
        return self._model_version

    @property
    def class_names(self) -> list[str]:
        """Get class names from model."""
        with self._model_lock:
            if self._model is None:
                return []
            return list(self._model.names.values())

    def load_model(self) -> bool:
        """Load or reload the model."""
        model_path = self.config.model_path

        if not model_path.exists():
            logger.warning(f"Model file not found: {model_path}")
            return False

        try:
            logger.info(f"Loading model from {model_path}...")
            new_model = YOLO(str(model_path))

            # Warm up the model
            dummy_input = np.zeros((self.config.input_size, self.config.input_size, 3), dtype=np.uint8)
            new_model.predict(dummy_input, verbose=False)

            with self._model_lock:
                self._model = new_model
                self._model_version = self._compute_version(model_path)

            logger.info(f"Model loaded successfully. Version: {self._model_version}")
            logger.info(f"Classes: {self.class_names}")
            return True

        except Exception as e:
            logger.error(f"Failed to load model: {e}")
            return False

    def _compute_version(self, model_path: Path) -> str:
        """Compute version string from model file."""
        mtime = model_path.stat().st_mtime
        return f"v{int(mtime)}"

    def start_hot_reload(self) -> None:
        """Start watching for model changes."""
        if self._watcher is not None:
            return

        self._watcher = ModelWatcher(
            model_path=self.config.model_path,
            on_change=self.load_model,
            interval=self.config.watch_interval,
        )
        self._watcher.start()

    def stop_hot_reload(self) -> None:
        """Stop watching for model changes."""
        if self._watcher:
            self._watcher.stop()
            self._watcher = None

    def update_class_mapping(self, mapping: dict[int, tuple[str, str]]) -> None:
        """Update class ID to SKU mapping.

        Args:
            mapping: Dict of class_id -> (sku_id, class_name)
        """
        self._class_mapping = mapping
        logger.info(f"Updated class mapping with {len(mapping)} classes")

    def detect(
        self,
        image_bytes: bytes,
        confidence_threshold: Optional[float] = None,
        iou_threshold: Optional[float] = None,
    ) -> InferenceResult:
        """Run detection on image.

        Args:
            image_bytes: JPEG/PNG image bytes
            confidence_threshold: Minimum confidence threshold
            iou_threshold: NMS IoU threshold

        Returns:
            InferenceResult with detections and metadata

        Raises:
            RuntimeError: If model is not loaded
        """
        with self._model_lock:
            if self._model is None:
                raise RuntimeError("Model not loaded")

            # Use defaults if not specified
            conf = confidence_threshold or self.config.default_confidence
            iou = iou_threshold or self.config.default_iou

            # Decode image
            image = Image.open(io.BytesIO(image_bytes))
            if image.mode != "RGB":
                image = image.convert("RGB")

            # Get original dimensions for normalization
            orig_width, orig_height = image.size

            # Run inference
            start_time = time.perf_counter()
            results = self._model.predict(
                image,
                conf=conf,
                iou=iou,
                imgsz=self.config.input_size,
                verbose=False,
            )
            inference_time = (time.perf_counter() - start_time) * 1000

            # Parse results
            detections = []
            if results and len(results) > 0:
                result = results[0]
                if result.boxes is not None:
                    boxes = result.boxes
                    for i in range(len(boxes)):
                        class_id = int(boxes.cls[i].item())
                        confidence = float(boxes.conf[i].item())

                        # Get bbox in xyxy format and normalize
                        x1, y1, x2, y2 = boxes.xyxy[i].tolist()
                        x1_norm = x1 / orig_width
                        y1_norm = y1 / orig_height
                        x2_norm = x2 / orig_width
                        y2_norm = y2 / orig_height

                        # Get class name from mapping or model
                        if class_id in self._class_mapping:
                            _, class_name = self._class_mapping[class_id]
                        else:
                            class_name = self._model.names.get(class_id, f"class_{class_id}")

                        detections.append(
                            DetectionResult(
                                class_id=class_id,
                                class_name=class_name,
                                confidence=confidence,
                                x1=x1_norm,
                                y1=y1_norm,
                                x2=x2_norm,
                                y2=y2_norm,
                            )
                        )

            return InferenceResult(
                detections=detections,
                inference_time_ms=inference_time,
                model_version=self._model_version,
            )

    def get_model_info(self) -> dict:
        """Get model metadata."""
        with self._model_lock:
            if self._model is None:
                return {
                    "version": "none",
                    "architecture": "none",
                    "class_names": [],
                    "input_size": self.config.input_size,
                }

            return {
                "version": self._model_version,
                "architecture": "yolov8",
                "class_names": list(self._model.names.values()),
                "input_size": self.config.input_size,
            }
