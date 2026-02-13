# Hardware Setup Guide

## Components Required

### Main Components
| Component | Quantity | Notes |
|-----------|----------|-------|
| ESP32-S3-CAM | 1 | With OV2640 camera |
| HX711 Load Cell Amplifier | 1 | 24-bit ADC |
| Load Cell (5kg) | 1 | 0.1g precision, strain gauge type |
| USB-C cable | 1 | For programming and power |

### Supporting Components
| Component | Quantity | Notes |
|-----------|----------|-------|
| Jumper wires | ~10 | Female-to-female recommended |
| Breadboard | 1 | For prototyping |
| Platform/plate | 1 | For scale surface |
| Mounting bracket | 1 | For angled camera mount |

---

## Wiring Diagram

```
                    ESP32-S3-CAM
                   ┌─────────────┐
                   │             │
                   │    CAM      │ (built-in OV2640)
                   │             │
                   │         3V3 │──────────────┐
                   │         GND │────────────┐ │
                   │             │            │ │
                   │       GPIO1 │────┐       │ │
                   │       GPIO2 │──┐ │       │ │
                   │             │  │ │       │ │
                   └─────────────┘  │ │       │ │
                                    │ │       │ │
                      HX711         │ │       │ │
                   ┌─────────────┐  │ │       │ │
              ┌────│ E+      VCC │──│─│───────│─┘
              │ ┌──│ E-      GND │──│─│───────┘
              │ │  │ A-      DT  │──┘ │ (Data)
              │ │  │ A+      SCK │────┘ (Clock)
              │ │  └─────────────┘
              │ │
              │ │   Load Cell (5kg)
              │ │  ┌─────────────┐
              │ └──│ White (E-)  │
              └────│ Black (E+)  │
                   │ Green (A-)  │──┐
                   │ Red (A+)    │──│─┐
                   └─────────────┘  │ │
                                    │ │
                      HX711         │ │
                   ┌─────────────┐  │ │
                   │ E+          │  │ │
                   │ E-          │  │ │
                   │ A-  ────────│──┘ │
                   │ A+  ────────│────┘
                   └─────────────┘
```

---

## Pin Connections

### ESP32-S3-CAM → HX711

| ESP32-S3-CAM | HX711 | Description |
|--------------|-------|-------------|
| 3V3 | VCC | Power (3.3V) |
| GND | GND | Ground |
| GPIO1 | DT | Data (configurable) |
| GPIO2 | SCK | Clock (configurable) |

**Note**: GPIO pins can be changed in firmware. Avoid GPIO pins used by camera.

### HX711 → Load Cell

| HX711 | Load Cell Wire | Description |
|-------|----------------|-------------|
| E+ | Red | Excitation + |
| E- | Black | Excitation - |
| A+ | White | Signal + (Channel A) |
| A- | Green | Signal - (Channel A) |

**Note**: Wire colors may vary by manufacturer. Check your load cell datasheet.

---

## ESP32-S3-CAM GPIO Reference

### Available GPIOs (not used by camera)

| GPIO | Recommended Use |
|------|-----------------|
| GPIO1 | HX711 DT (Data) |
| GPIO2 | HX711 SCK (Clock) |
| GPIO3 | Reserved (future use) |
| GPIO46 | Reserved (future use) |

### GPIOs Used by Camera (DO NOT USE)

These pins are occupied by the OV2640 camera module:
- GPIO4-5 (I2C for camera control)
- GPIO6-11 (typically flash/PSRAM)
- GPIO39-42 (camera data)

Check your specific ESP32-S3-CAM board pinout.

---

## Assembly Steps

### Step 1: Load Cell Setup

1. Mount load cell on stable base
2. Attach platform/plate to load cell (where items will be placed)
3. Ensure load cell is level

```
     ┌─────────────────┐
     │    Platform     │  ← Items placed here
     └────────┬────────┘
              │
     ┌────────┴────────┐
     │    Load Cell    │
     └────────┬────────┘
              │
     ┌────────┴────────┐
     │      Base       │
     └─────────────────┘
```

### Step 2: Wire HX711 to Load Cell

1. Connect load cell wires to HX711 (E+, E-, A+, A-)
2. Double-check wire colors against your load cell datasheet

### Step 3: Wire HX711 to ESP32-S3-CAM

1. Connect VCC → 3V3
2. Connect GND → GND
3. Connect DT → GPIO1
4. Connect SCK → GPIO2

### Step 4: Mount Camera

1. Position ESP32-S3-CAM at angle facing the scale platform
2. Recommended angle: 30-45° from horizontal
3. Distance: 20-30cm from platform center (adjust based on field of view)

```
    ESP32-S3-CAM
         │\
         │ \  30-45°
         │  \
         │   \
         │    ↘
    ─────┴─────────────
       Scale Platform
```

### Step 5: Power Up

1. Connect ESP32-S3-CAM via USB-C
2. Verify power LED on ESP32-S3-CAM
3. Verify power LED on HX711 (if present)

---

## Initial Testing

### Test 1: Camera

```cpp
// Simple camera test - check serial output for JPEG data
#include "esp_camera.h"

void setup() {
  Serial.begin(115200);

  camera_config_t config;
  // ... camera config (see firmware)

  esp_err_t err = esp_camera_init(&config);
  if (err != ESP_OK) {
    Serial.printf("Camera init failed: 0x%x", err);
    return;
  }
  Serial.println("Camera OK");
}
```

### Test 2: HX711 / Scale

```cpp
// Simple scale test
#include "HX711.h"

HX711 scale;
const int DT_PIN = 1;
const int SCK_PIN = 2;

void setup() {
  Serial.begin(115200);
  scale.begin(DT_PIN, SCK_PIN);
  Serial.println("Scale initialized");
}

void loop() {
  if (scale.is_ready()) {
    long reading = scale.read();
    Serial.printf("Raw reading: %ld\n", reading);
  }
  delay(500);
}
```

---

## Troubleshooting

| Issue | Possible Cause | Solution |
|-------|---------------|----------|
| Camera not detected | Wrong pins / bad connection | Check camera ribbon cable |
| HX711 not responding | Wrong GPIO pins | Verify DT/SCK connections |
| Scale readings unstable | Power supply noise | Use capacitor on VCC, stable power |
| Weight always zero | Load cell wiring wrong | Check E+/E-/A+/A- connections |
| WiFi not connecting | Antenna issue | Ensure antenna not covered |

---

## Next Steps

1. [ ] Assemble hardware
2. [ ] Flash test firmware
3. [ ] Verify camera captures images
4. [ ] Verify scale reads weight changes
5. [ ] Calibrate scale with known weights
6. [ ] Proceed to data collection phase
