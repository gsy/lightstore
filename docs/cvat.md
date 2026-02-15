# CVAT Deployment Guide with NVIDIA GPU Support

## Prerequisites Setup

### 1. Install NVIDIA Drivers & CUDA Toolkit

```bash
# Check if NVIDIA driver is installed
nvidia-smi

# If not installed (Ubuntu/Debian):
sudo apt update
sudo apt install nvidia-driver-525  # or latest version
sudo reboot

# Verify after reboot
nvidia-smi  # Should show your RTX 2080
```

### 2. Install NVIDIA Container Toolkit (for Docker GPU access)

```bash
# Add NVIDIA package repository
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | sudo apt-key add -
curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.list | \
    sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list

# Install nvidia-container-toolkit
sudo apt update
sudo apt install -y nvidia-container-toolkit

# Configure Docker to use NVIDIA runtime
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker

# Test GPU access in container
docker run --rm --gpus all nvidia/cuda:11.8.0-base-ubuntu22.04 nvidia-smi
```

## CVAT Installation with GPU Support

### 3. Clone and Configure CVAT

```bash
# Clone CVAT repository
cd ~/
git clone https://github.com/cvat-ai/cvat
cd cvat

# Create environment configuration
cat > .env <<EOF
# GPU Configuration
CVAT_NUCLIO_AUTO_GPU=1
CVAT_NUCLIO_NUM_GPU=1

# Database
POSTGRES_HOST=cvat_db
POSTGRES_DB=cvat
POSTGRES_USER=root
POSTGRES_PASSWORD=cvat_postgresql_password

# Redis
REDIS_HOST=cvat_redis

# Ports
CVAT_HOST=localhost
CVAT_PORT=8080
EOF
```

### 4. Modify docker-compose.yml for GPU Support

```bash
# Backup original
cp docker-compose.yml docker-compose.yml.backup

# Edit docker-compose.yml to add GPU support to serverless workers
```

Add this to the `cvat_worker_auto_annotation` service in `docker-compose.yml`:

```yaml
  cvat_worker_auto_annotation:
    container_name: cvat_worker_auto_annotation
    image: cvat/server:${CVAT_VERSION:-dev}
    restart: always
    depends_on:
      - cvat_redis
      - cvat_db
    environment:
      CVAT_REDIS_HOST: cvat_redis
      CVAT_POSTGRES_HOST: cvat_db
      CVAT_SERVERLESS: 1
    command: run worker.auto_annotation --loglevel=INFO
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

### 5. Deploy CVAT with Serverless (Nuclio) for Auto-Annotation

```bash
# Deploy CVAT with serverless functions (includes auto-annotation)
docker compose -f docker-compose.yml -f components/serverless/docker-compose.serverless.yml up -d

# Wait for all services to start (2-3 minutes)
docker compose ps

# Check logs
docker compose logs -f
```

### 6. Deploy GPU-Accelerated Auto-Annotation Functions

CVAT uses Nuclio serverless functions for ML models. Deploy GPU-enabled functions:

```bash
# Navigate to serverless functions directory
cd serverless/

# Deploy YOLOv5 with GPU support (good for object detection)
nuctl deploy --project-name cvat \
  --path ./pytorch/ultralytics/yolov5/nuclio \
  --platform local \
  --base-image ultralytics/yolov5:latest \
  --triggers '{"myHttpTrigger": {"maxWorkers": 2}}' \
  --resource-limit nvidia.com/gpu=1

# Alternative: Use CVAT's built-in deployment script
cd ~/cvat/serverless
./deploy_cpu.sh serverless/pytorch/facebookresearch/detectron2/retinanet_r101  # For CPU
./deploy_gpu.sh serverless/pytorch/ultralytics/yolov5  # For GPU (if script exists)
```

**Note**: CVAT's serverless functions need to be configured individually. Check available functions:

```bash
cd ~/cvat/serverless
ls -la pytorch/  # See available model integrations
```

## Manual GPU Function Deployment

If automatic deployment doesn't work, manually configure Nuclio:

### 7. Install Nuclio CLI

```bash
# Download nuclio CLI
curl -s https://api.github.com/repos/nuclio/nuclio/releases/latest \
  | grep -i "browser_download_url.*nuctl.*$(uname)" \
  | cut -d : -f 2,3 \
  | tr -d \" \
  | wget -O nuctl -qi -
chmod +x nuctl
sudo mv nuctl /usr/local/bin/
```

### 8. Deploy Custom GPU-Enabled Function

Create a YOLOv5 function with GPU support:

```bash
cd ~/cvat/serverless/pytorch/ultralytics/yolov5/nuclio

# Create function config with GPU
cat > function-gpu.yaml <<EOF
apiVersion: nuclio.io/v1
kind: NuclioFunction
metadata:
  name: pth-yolov5
  namespace: nuclio
spec:
  description: YOLOv5 object detection with GPU
  runtime: python:3.8
  handler: main:handler
  env:
    - name: PYTHONPATH
      value: /opt/nuclio
  triggers:
    http:
      maxWorkers: 2
      kind: http
  resources:
    limits:
      nvidia.com/gpu: "1"
  build:
    baseImage: ultralytics/yolov5:latest
    commands:
      - pip install pillow
EOF

# Deploy
nuctl deploy --project-name cvat --path . --file function-gpu.yaml --platform local
```

## CVAT Usage with Pre-Labeling

### 9. Access CVAT Web Interface

```bash
# Create superuser account
docker exec -it cvat_server bash -c "python manage.py createsuperuser"
# Enter username, email, password

# Open browser
open http://localhost:8080
```

### 10. Configure Auto-Annotation in CVAT

1. **Login** to CVAT at `http://localhost:8080`

2. **Create a Project**:
   - Click "Projects" → "Create new project"
   - Name: "Vending Machine Beverages"
   - Add labels (your SKU IDs):
     ```
     coca-cola-500ml
     pepsi-330ml
     water-mineral-500ml
     energy-drink-redbull
     ... (add all your SKUs)
     ```

3. **Create a Task**:
   - Click "Tasks" → "Create new task"
   - Name: "Batch 1 - Unlabeled Images"
   - Select project: "Vending Machine Beverages"
   - Upload images (drag & drop or select files)
   - Click "Submit"

4. **Run Auto-Annotation**:
   - Open the task
   - Click "Actions" → "Automatic Annotation"
   - Select model: "YOLOv5" (or available GPU model)
   - Configure:
     - Mapping: Map model classes to your labels
     - Threshold: 0.5 (adjust for confidence)
     - Cleanup: Enable to remove low-confidence boxes
   - Click "Annotate"
   - Wait for processing (watch GPU usage with `nvidia-smi`)

5. **Review & Edit**:
   - Pre-labeled boxes appear automatically
   - Edit, add, or delete boxes as needed
   - Use keyboard shortcuts:
     - `N` - next frame/image
     - `Shift+N` - create new box
     - `Delete` - remove selected box
     - `F` - finish drawing

6. **Export Annotations**:
   - Click task → "Actions" → "Export annotations"
   - Choose format:
     - **YOLO 1.1** - for training YOLO models
     - **COCO 1.0** - for TensorFlow/PyTorch
     - **CVAT for images 1.1** - CVAT native format

## Monitoring GPU Usage

```bash
# Real-time GPU monitoring
watch -n 1 nvidia-smi

# Check which container is using GPU
docker ps --format "table {{.Names}}\t{{.Status}}"

# Check CVAT worker logs
docker compose logs -f cvat_worker_auto_annotation

# Check Nuclio function logs (if using serverless)
docker logs -f $(docker ps -q -f name=nuclio)
```

## Troubleshooting

### GPU Not Detected in Container

```bash
# Verify Docker can access GPU
docker run --rm --gpus all nvidia/cuda:11.8.0-base-ubuntu22.04 nvidia-smi

# If fails, restart Docker with NVIDIA runtime
sudo systemctl restart docker

# Check nvidia-container-runtime is configured
cat /etc/docker/daemon.json  # Should show nvidia runtime
```

### Auto-Annotation Models Not Available

```bash
# Check if serverless component is running
docker compose -f docker-compose.yml -f components/serverless/docker-compose.serverless.yml ps

# Manually deploy a model
cd ~/cvat/serverless
# List available models
find . -name "nuclio" -type d

# Deploy specific model
cd pytorch/ultralytics/yolov5/nuclio
nuctl deploy --project-name cvat --path . --platform local
```

### Performance Optimization

```yaml
# Edit docker-compose.yml - increase worker resources
cvat_worker_auto_annotation:
  environment:
    NUMPROCS: 2  # Parallel workers
  deploy:
    resources:
      reservations:
        devices:
          - capabilities: [gpu]
            count: 1  # Use 1 GPU (you have RTX 2080)
```

## Quick Start Script

Save this as `start-cvat-gpu.sh`:

```bash
#!/bin/bash

echo "🚀 Starting CVAT with GPU support..."

# Check GPU
if ! nvidia-smi &> /dev/null; then
    echo "❌ NVIDIA GPU not detected"
    exit 1
fi

# Check Docker GPU access
if ! docker run --rm --gpus all nvidia/cuda:11.8.0-base-ubuntu22.04 nvidia-smi &> /dev/null; then
    echo "❌ Docker cannot access GPU"
    exit 1
fi

cd ~/cvat

# Start CVAT with serverless
docker compose -f docker-compose.yml -f components/serverless/docker-compose.serverless.yml up -d

echo "⏳ Waiting for services to start..."
sleep 30

echo "✅ CVAT running at http://localhost:8080"
echo "📊 Monitor GPU: watch -n 1 nvidia-smi"
```

```bash
chmod +x start-cvat-gpu.sh
./start-cvat-gpu.sh
```

## Next Steps

1. **Prepare your beverage images** from ESP32-CAM captures
2. **Create SKU label list** in CVAT matching your catalog
3. **Upload batch of images** (start with 100-200)
4. **Run GPU pre-labeling** with YOLOv5/Detectron2
5. **Manual review** - correct boxes, add missing items
6. **Export annotations** in YOLO format for model training
7. **Train custom model** on labeled dataset
8. **Deploy to ESP32** as TFLite model
