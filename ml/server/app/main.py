"""ML Detection Server entry point."""

import logging
import signal
import sys
import time
from concurrent import futures

import grpc

from .config import Config
from .detector import Detector
from .servicer import DetectionServicer


def setup_logging(level: str) -> None:
    """Configure logging."""
    logging.basicConfig(
        level=getattr(logging, level.upper()),
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        handlers=[logging.StreamHandler(sys.stdout)],
    )


def serve() -> None:
    """Start the gRPC server."""
    config = Config.from_env()
    setup_logging(config.log_level)
    logger = logging.getLogger(__name__)

    logger.info("Starting ML Detection Server...")
    logger.info(f"Configuration: {config}")

    # Initialize detector
    detector = Detector(config.model)

    # Try to load model (may not exist yet)
    if config.model.model_path.exists():
        detector.load_model()
    else:
        logger.warning(
            f"Model not found at {config.model.model_path}. "
            "Server will start but detection unavailable until model is deployed."
        )

    # Start hot reload watcher
    detector.start_hot_reload()

    # Create gRPC server
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=config.server.max_workers),
        options=[
            ("grpc.max_receive_message_length", 50 * 1024 * 1024),  # 50MB max
            ("grpc.max_send_message_length", 50 * 1024 * 1024),
        ],
    )

    # Import generated code and register servicer
    from .generated import detection_pb2_grpc

    servicer = DetectionServicer(detector, start_time=time.time())
    detection_pb2_grpc.add_DetectionServiceServicer_to_server(servicer, server)

    # Start server
    server.add_insecure_port(config.server.address)
    server.start()
    logger.info(f"Server started on {config.server.address}")

    # Graceful shutdown handling
    shutdown_event = False

    def handle_shutdown(signum, frame):
        nonlocal shutdown_event
        if shutdown_event:
            return
        shutdown_event = True
        logger.info("Shutdown signal received, stopping server...")
        detector.stop_hot_reload()
        server.stop(grace=5)

    signal.signal(signal.SIGTERM, handle_shutdown)
    signal.signal(signal.SIGINT, handle_shutdown)

    # Wait for termination
    server.wait_for_termination()
    logger.info("Server stopped")


def main() -> None:
    """Main entry point."""
    serve()


if __name__ == "__main__":
    main()
