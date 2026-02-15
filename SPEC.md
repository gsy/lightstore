# Beverage Vending Machine Recognition System

## Overview

ESP32-S3 based smart vending machine system that recognizes beverages using computer vision and weight verification. Supports 20-30 beverage items with hybrid recognition (on-device ML with cloud fallback).

---

## Hardware Components

| Component | Model/Spec | Notes |
|-----------|------------|-------|
| MCU | ESP32-S3-CAM | Dual-core, 8MB PSRAM, built-in camera |
| Camera | OV2640 (typical) | Included with ESP32-S3-CAM, angled mount |
| Load Cell | 5kg capacity | 0.1g precision required |
| ADC | HX711 | 24-bit ADC for load cell |
| WiFi | Built-in | For backend communication |
| LED Lighting | External | Not controlled by ESP32, separate power |

**Notes**:
- No on-device display - all user feedback via mobile app
- LED lighting always-on or on separate circuit (not ESP32-controlled)
- Camera mounted at angle for better label visibility

---

## System Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Mobile App │◄───►│   Backend   │◄───►│ ESP32-S3-CAM│
│  (Customer) │     │   Server    │     │  (Device)   │
└─────────────┘     └─────────────┘     └─────────────┘
      │                   │                    │
      │ • Scan QR code    │ • Auth & session   │ • Camera capture
      │ • View items      │ • Cloud inference  │ • Weight sensing
      │ • Payment         │ • Payment process  │ • On-device ML
      │ • Receipt         │ • Inventory mgmt   │ • Data upload
      └───────────────────┴────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                 ESP32-S3-CAM Device                     │
│  ┌─────────┐  ┌─────────────┐  ┌──────────────────┐    │
│  │ Camera  │  │ TFLite Micro│  │ Weight Sensor    │    │
│  │ OV2640  │─►│ Object Det. │◄─│ HX711 + LoadCell │    │
│  └─────────┘  └──────┬──────┘  └──────────────────┘    │
│                      │                                  │
│              ┌───────▼───────┐                         │
│              │ Detection     │                         │
│              │ Results       │                         │
│              └───────┬───────┘                         │
│                      │                                  │
│         ┌────────────▼────────────┐                    │
│         │ All items confident?    │                    │
│         └────────────┬────────────┘                    │
│              Yes     │     No                          │
│               │      │      │                          │
│               ▼      │      ▼                          │
│         ┌─────────┐  │  ┌─────────────┐               │
│         │ Send to │  │  │ Upload img  │               │
│         │ Backend │  │  │ to Backend  │               │
│         │ (JSON)  │  │  │ for Cloud ML│               │
│         └─────────┘  │  └─────────────┘               │
└─────────────────────────────────────────────────────────┘
```

---

## Recognition Strategy

### Multi-Item Object Detection

Since multiple beverages can be placed simultaneously, we need **object detection** (not just classification):

| Approach | Model | Pros | Cons |
|----------|-------|------|------|
| YOLO-based | YOLOv5-nano, YOLOv8-nano | Fast, accurate | May be too large for ESP32 |
| SSD MobileNet | SSD MobileNetV2 | Good for embedded | Moderate accuracy |
| FOMO | Edge Impulse FOMO | Designed for MCU | Less precise bounding boxes |

**Recommended**: Start with **FOMO** or **SSD MobileNetV1** for on-device, full **YOLOv8** on cloud.

**Model Constraints**:
- Max detections per image: 10 items
- Classes: 20-30 beverage types
- Input resolution: 320x320 or 240x240 (TBD based on performance)

### Primary: On-Device Inference
- **Model**: TensorFlow Lite Micro (quantized INT8)
- **Type**: Object detection (multi-item)
- **Input**: Camera image (e.g., 320x320 or smaller)
- **Output**: List of detected items with bounding boxes + confidence
- **Weight verification**: Sum of detected item weights ≈ measured total weight (±tolerance)

### Fallback: Cloud Recognition
- **Trigger**:
  - Any item confidence < 80%
  - Weight mismatch > tolerance
  - Unknown object detected
- **Method**: Upload JPEG image via WiFi to self-hosted backend
- **Cloud model**: YOLOv8 or similar (more accurate, no size constraints)

### Weight Cross-Validation

```
Detected items: [Coke 330ml, Sprite 330ml]
Expected weights: [350g, 345g]
Expected total: 695g (±5g tolerance)
Measured weight: 698g
Status: MATCH ✓
```

If weight mismatch detected → trigger cloud recognition for verification.

---

## Beverage Database

| Field | Description |
|-------|-------------|
| ID | Unique identifier |
| Name | Beverage name |
| Weight | Expected weight (g) with tolerance |
| Price | Unit price |
| Image | Reference image(s) for training |
| Barcode | Optional backup identification |

**Capacity**: 20-30 items

---

## Scale Specifications

| Parameter | Value |
|-----------|-------|
| Max capacity | 5 kg |
| Resolution | 0.1 g |
| Tolerance | ±0.5 g (TBD) |
| Load cell type | Strain gauge |
| ADC | HX711 (24-bit) |

---

## Functional Requirements

### ESP32 Device
1. [ ] Capture image when weight change detected
2. [ ] Measure total weight of items
3. [ ] Run on-device object detection (multi-item)
4. [ ] Verify results using weight cross-check
5. [ ] Send results to backend (JSON: items + confidence + weight)
6. [ ] Upload image to backend if confidence low
7. [ ] Communicate status via WiFi API

### Backend Server
1. [ ] Device registration and management
2. [ ] Session management (QR code → session)
3. [ ] Receive detection results from device
4. [ ] Run cloud ML inference when needed
5. [ ] Beverage database management
6. [ ] Price calculation
7. [ ] Payment processing integration
8. [ ] Transaction logging

### Mobile App (Provided by client)
1. [ ] QR code scanning
2. [ ] Display detected items and prices
3. [ ] Payment flow
4. [ ] Receipt/history

---

## Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| Recognition time (on-device) | < 2 seconds |
| Recognition accuracy | > 95% |
| Weight measurement time | < 500 ms |
| Cloud fallback time | < 5 seconds |
| Power | 5V USB or battery TBD |

---

## User Flow

```
1. Customer approaches vending machine
           │
           ▼
2. Scans QR code on machine with mobile app
           │
           ▼
3. App connects to backend, starts session
           │
           ▼
4. Customer places beverage(s) on scale
           │
           ▼
5. ESP32 detects weight change → triggers recognition
           │
           ▼
6. On-device ML detects items
           │
           ├─── High confidence + weight match ───►  Results sent to backend
           │                                                    │
           └─── Low confidence / mismatch ───► Image uploaded   │
                                                to backend      │
                                                    │           │
                                                    ▼           │
                                              Cloud inference   │
                                                    │           │
                                                    └───────────┤
                                                                ▼
7. Backend updates app with detected items & prices
           │
           ▼
8. Customer confirms and pays in app
           │
           ▼
9. Transaction complete, customer takes items
           │
           ▼
10. [If wrong] Customer requests refund via app
```

### Error Handling Strategy
- Always return **highest confidence result** even if below threshold
- Weight mismatch triggers cloud verification but still returns best guess
- Customer can dispute via app → refund processed
- Disputed transactions logged for model improvement

---

## Resolved Design Decisions

| Decision | Choice | Notes |
|----------|--------|-------|
| QR code | Static machine ID | Simple, app fetches session from backend |
| Lighting | LED (not ESP32-controlled) | Always-on or separate power circuit |
| Camera angle | Angled view | Better for reading labels |
| Max items | 10 items | Design constraint for detection model |
| Error handling | Use highest confidence result | Refund via app if incorrect |

---

## Additional Design Decisions

| Decision | Choice |
|----------|--------|
| LED power | Always-on when machine powered |
| Beverage orientation | Any direction (model must handle rotation variance) |

---

## API Design

### Device → Backend

**POST /api/device/detection**
```json
{
  "device_id": "VM-001",
  "session_id": "abc123",
  "timestamp": "2024-01-15T10:30:00Z",
  "items": [
    {"id": "coke-330", "confidence": 0.95, "bbox": [x, y, w, h]},
    {"id": "sprite-330", "confidence": 0.88, "bbox": [x, y, w, h]}
  ],
  "total_weight": 698.5,
  "image_required": false
}
```

**POST /api/device/image** (fallback)
```json
{
  "device_id": "VM-001",
  "session_id": "abc123",
  "image": "<base64 encoded JPEG>",
  "total_weight": 698.5
}
```

### Backend → Device

**Response**
```json
{
  "status": "confirmed",
  "items": [
    {"id": "coke-330", "name": "Coca-Cola 330ml", "price": 2.50},
    {"id": "sprite-330", "name": "Sprite 330ml", "price": 2.50}
  ],
  "total": 5.00
}
```

---

## Data Collection Strategy

### Training Image Collection
1. **Setup**: Mount ESP32-S3-CAM in final position with proper lighting
2. **Capture variations**:
   - Each beverage individually
   - Multiple beverages together (2-5 items)
   - Different orientations (upright, tilted, label facing different directions)
   - Different lighting conditions
3. **Target**: ~100-200 images per beverage, ~500+ multi-item combinations
4. **Labeling tool**: LabelImg, CVAT, or Roboflow

### Weight Database
- Weigh each beverage type multiple times
- Record: mean weight, min, max, standard deviation
- Account for temperature variations (cold vs room temp)

---

## Development Phases

### Phase 1: Hardware Prototype ← CURRENT
- [x] ESP32-S3-CAM setup and camera test (firmware ready)
- [x] HX711 + load cell wiring guide (docs ready)
- [x] Basic image capture + weight reading (firmware ready)
- [ ] Calibration with known weights
- [ ] WiFi connectivity to backend

### Phase 2: Data Collection
- [ ] Build data collection firmware
- [ ] Capture training images for all beverages
- [ ] Label images with bounding boxes
- [ ] Record weight data for each beverage

### Phase 3: ML Model Development
- [ ] Train object detection model (FOMO/SSD for device)
- [ ] Train YOLOv8 model for cloud
- [ ] Convert device model to TFLite INT8
- [ ] Benchmark on ESP32-S3

### Phase 4: Backend Development ← CURRENT
- [x] Device API endpoints (register, detection, image upload)
- [x] Session management (start, get, confirm, cancel)
- [x] Beverage CRUD endpoints
- [x] Transaction & refund endpoints
- [x] Database schema & migrations
- [x] Docker & Kubernetes deployment configs
- [ ] Cloud inference service (ML)

### Phase 5: Integration & Testing
- [ ] End-to-end device + backend flow
- [ ] Weight validation logic
- [ ] Cloud fallback mechanism
- [ ] App integration testing

### Phase 6: Deployment
- [ ] Production firmware
- [ ] Enclosure design
- [ ] Field testing
- [ ] Documentation

---

## Environment

- **Location**: Indoor only
- **Temperature**: Room temperature (controlled environment)
- **Lighting**: Supplementary LED lighting (always-on, separate circuit)

---

## Notes

_Add discussion notes here_

