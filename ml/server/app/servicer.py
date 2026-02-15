"""gRPC service implementation."""

import logging
import time
import uuid
from typing import TYPE_CHECKING

import grpc

from .detector import Detector

if TYPE_CHECKING:
    from .generated import detection_pb2, detection_pb2_grpc

logger = logging.getLogger(__name__)


class DetectionServicer:
    """gRPC Detection Service implementation."""

    def __init__(self, detector: Detector, start_time: float):
        self.detector = detector
        self.start_time = start_time

    def Detect(
        self,
        request: "detection_pb2.DetectRequest",
        context: grpc.ServicerContext,
    ) -> "detection_pb2.DetectResponse":
        """Handle detection request."""
        from .generated import detection_pb2

        request_id = str(uuid.uuid4())[:8]
        logger.info(
            f"[{request_id}] Detection request from device: {request.device_id}, "
            f"image size: {len(request.image)} bytes"
        )

        if not self.detector.model_loaded:
            context.set_code(grpc.StatusCode.UNAVAILABLE)
            context.set_details("Model not loaded")
            return detection_pb2.DetectResponse()

        try:
            # Get thresholds (use 0 to indicate default)
            conf_threshold = request.confidence_threshold if request.confidence_threshold > 0 else None
            iou_threshold = request.iou_threshold if request.iou_threshold > 0 else None

            result = self.detector.detect(
                image_bytes=request.image,
                confidence_threshold=conf_threshold,
                iou_threshold=iou_threshold,
            )

            # Build response
            detections = []
            for det in result.detections:
                # Get SKU ID from mapping if available
                sku_id = ""
                if det.class_id in self.detector._class_mapping:
                    sku_id = self.detector._class_mapping[det.class_id][0]

                detections.append(
                    detection_pb2.Detection(
                        class_name=det.class_name,
                        sku_id=sku_id,
                        class_id=det.class_id,
                        confidence=det.confidence,
                        bbox=detection_pb2.BoundingBox(
                            x1=det.x1,
                            y1=det.y1,
                            x2=det.x2,
                            y2=det.y2,
                        ),
                    )
                )

            logger.info(
                f"[{request_id}] Found {len(detections)} detections "
                f"in {result.inference_time_ms:.1f}ms"
            )

            return detection_pb2.DetectResponse(
                detections=detections,
                model_version=result.model_version,
                inference_time_ms=result.inference_time_ms,
                request_id=request_id,
            )

        except Exception as e:
            logger.error(f"[{request_id}] Detection failed: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return detection_pb2.DetectResponse()

    def HealthCheck(
        self,
        request: "detection_pb2.Empty",
        context: grpc.ServicerContext,
    ) -> "detection_pb2.HealthResponse":
        """Handle health check request."""
        from .generated import detection_pb2

        uptime = int(time.time() - self.start_time)
        model_loaded = self.detector.model_loaded

        if model_loaded:
            status = "healthy"
            healthy = True
        else:
            status = "model_not_loaded"
            healthy = False

        return detection_pb2.HealthResponse(
            healthy=healthy,
            status=status,
            model_loaded=model_loaded,
            uptime_seconds=uptime,
        )

    def GetModelInfo(
        self,
        request: "detection_pb2.Empty",
        context: grpc.ServicerContext,
    ) -> "detection_pb2.ModelInfo":
        """Handle model info request."""
        from .generated import detection_pb2

        info = self.detector.get_model_info()

        return detection_pb2.ModelInfo(
            version=info["version"],
            architecture=info["architecture"],
            class_names=info["class_names"],
            input_width=info["input_size"],
            input_height=info["input_size"],
            trained_at="",  # TODO: Store in model metadata
            mAP50=0.0,  # TODO: Store in model metadata
            mAP50_95=0.0,  # TODO: Store in model metadata
        )

    def SyncClasses(
        self,
        request: "detection_pb2.SyncClassesRequest",
        context: grpc.ServicerContext,
    ) -> "detection_pb2.SyncClassesResponse":
        """Handle class sync request from catalog service."""
        from .generated import detection_pb2

        mapping = {}
        for cls in request.classes:
            mapping[cls.class_id] = (cls.sku_id, cls.class_name)

        self.detector.update_class_mapping(mapping)

        logger.info(f"Synced {len(mapping)} classes from catalog")

        return detection_pb2.SyncClassesResponse(
            success=True,
            class_count=len(mapping),
        )
