"""Configuration management for ML server."""

import os
from dataclasses import dataclass, field
from pathlib import Path


@dataclass
class ServerConfig:
    """gRPC server configuration."""

    host: str = "0.0.0.0"
    port: int = 50051
    max_workers: int = 10

    @property
    def address(self) -> str:
        return f"{self.host}:{self.port}"


@dataclass
class ModelConfig:
    """Model configuration."""

    model_dir: Path = field(default_factory=lambda: Path("models"))
    model_name: str = "best.pt"
    watch_interval: float = 5.0  # seconds between model file checks

    # Inference defaults
    default_confidence: float = 0.5
    default_iou: float = 0.45
    input_size: int = 640

    @property
    def model_path(self) -> Path:
        return self.model_dir / self.model_name


@dataclass
class CatalogConfig:
    """Catalog service connection for class sync."""

    grpc_address: str = "localhost:8081"
    sync_interval: float = 300.0  # seconds between class syncs


@dataclass
class Config:
    """Main configuration container."""

    server: ServerConfig = field(default_factory=ServerConfig)
    model: ModelConfig = field(default_factory=ModelConfig)
    catalog: CatalogConfig = field(default_factory=CatalogConfig)

    debug: bool = False
    log_level: str = "INFO"

    @classmethod
    def from_env(cls) -> "Config":
        """Load configuration from environment variables."""
        return cls(
            server=ServerConfig(
                host=os.getenv("GRPC_HOST", "0.0.0.0"),
                port=int(os.getenv("GRPC_PORT", "50051")),
                max_workers=int(os.getenv("GRPC_MAX_WORKERS", "10")),
            ),
            model=ModelConfig(
                model_dir=Path(os.getenv("MODEL_DIR", "models")),
                model_name=os.getenv("MODEL_NAME", "best.pt"),
                watch_interval=float(os.getenv("MODEL_WATCH_INTERVAL", "5.0")),
                default_confidence=float(os.getenv("DEFAULT_CONFIDENCE", "0.5")),
                default_iou=float(os.getenv("DEFAULT_IOU", "0.45")),
                input_size=int(os.getenv("INPUT_SIZE", "640")),
            ),
            catalog=CatalogConfig(
                grpc_address=os.getenv("CATALOG_GRPC_ADDRESS", "localhost:8081"),
                sync_interval=float(os.getenv("CLASS_SYNC_INTERVAL", "300.0")),
            ),
            debug=os.getenv("DEBUG", "false").lower() == "true",
            log_level=os.getenv("LOG_LEVEL", "INFO"),
        )
