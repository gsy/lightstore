"""Tests for detector module."""

import io
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest
from PIL import Image

from app.config import ModelConfig
from app.detector import Detector, DetectionResult, InferenceResult


@pytest.fixture
def model_config(tmp_path: Path) -> ModelConfig:
    """Create test model config."""
    return ModelConfig(
        model_dir=tmp_path / "models",
        model_name="test.pt",
        watch_interval=1.0,
        default_confidence=0.5,
        default_iou=0.45,
        input_size=640,
    )


@pytest.fixture
def sample_image_bytes() -> bytes:
    """Create sample test image."""
    img = Image.new("RGB", (640, 480), color="red")
    buffer = io.BytesIO()
    img.save(buffer, format="JPEG")
    return buffer.getvalue()


class TestDetector:
    """Tests for Detector class."""

    def test_init(self, model_config: ModelConfig) -> None:
        """Test detector initialization."""
        detector = Detector(model_config)
        assert detector.model_loaded is False
        assert detector.model_version == "unknown"

    def test_model_not_found(self, model_config: ModelConfig) -> None:
        """Test loading non-existent model."""
        detector = Detector(model_config)
        result = detector.load_model()
        assert result is False
        assert detector.model_loaded is False

    @patch("app.detector.YOLO")
    def test_load_model_success(
        self,
        mock_yolo_class: MagicMock,
        model_config: ModelConfig,
    ) -> None:
        """Test successful model loading."""
        # Create model file
        model_config.model_dir.mkdir(parents=True)
        model_path = model_config.model_path
        model_path.write_text("mock model")

        # Setup mock
        mock_model = MagicMock()
        mock_model.names = {0: "cola", 1: "sprite"}
        mock_model.predict.return_value = []
        mock_yolo_class.return_value = mock_model

        detector = Detector(model_config)
        result = detector.load_model()

        assert result is True
        assert detector.model_loaded is True
        assert "cola" in detector.class_names

    def test_detect_without_model(
        self,
        model_config: ModelConfig,
        sample_image_bytes: bytes,
    ) -> None:
        """Test detection fails without loaded model."""
        detector = Detector(model_config)

        with pytest.raises(RuntimeError, match="Model not loaded"):
            detector.detect(sample_image_bytes)

    @patch("app.detector.YOLO")
    def test_detect_success(
        self,
        mock_yolo_class: MagicMock,
        model_config: ModelConfig,
        sample_image_bytes: bytes,
    ) -> None:
        """Test successful detection."""
        import torch

        # Create model file
        model_config.model_dir.mkdir(parents=True)
        model_config.model_path.write_text("mock model")

        # Setup mock model
        mock_model = MagicMock()
        mock_model.names = {0: "cola", 1: "sprite"}

        # Mock detection result
        mock_boxes = MagicMock()
        mock_boxes.cls = torch.tensor([0])
        mock_boxes.conf = torch.tensor([0.95])
        mock_boxes.xyxy = torch.tensor([[100, 100, 200, 200]])
        mock_boxes.__len__ = lambda self: 1

        mock_result = MagicMock()
        mock_result.boxes = mock_boxes

        mock_model.predict.return_value = [mock_result]
        mock_yolo_class.return_value = mock_model

        detector = Detector(model_config)
        detector.load_model()

        result = detector.detect(sample_image_bytes)

        assert isinstance(result, InferenceResult)
        assert len(result.detections) == 1
        assert result.detections[0].class_name == "cola"
        assert result.detections[0].confidence == pytest.approx(0.95)

    def test_update_class_mapping(self, model_config: ModelConfig) -> None:
        """Test updating class mapping."""
        detector = Detector(model_config)

        mapping = {
            0: ("sku-001", "Coca-Cola"),
            1: ("sku-002", "Sprite"),
        }
        detector.update_class_mapping(mapping)

        assert detector._class_mapping == mapping
