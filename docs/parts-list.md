# Parts List

## Essential Components

| Component | Specification | Qty | Est. Price (USD) | Notes |
|-----------|---------------|-----|------------------|-------|
| ESP32-S3-CAM | ESP32-S3 + OV2640 camera | 1 | $15-25 | With 8MB PSRAM recommended |
| HX711 Module | 24-bit ADC breakout | 1 | $2-5 | Green or purple PCB common |
| Load Cell | 5kg strain gauge | 1 | $5-10 | 4-wire, beam type |
| USB-C Cable | Data + power capable | 1 | $5 | For programming |

**Subtotal**: ~$27-45

---

## Recommended Boards

### ESP32-S3-CAM Options

| Board | PSRAM | Price | Link |
|-------|-------|-------|------|
| Freenove ESP32-S3-WROOM CAM | 8MB | ~$20 | Amazon, AliExpress |
| AI-Thinker ESP32-S3-CAM | 8MB | ~$15 | AliExpress |
| Seeed Studio XIAO ESP32S3 Sense | 8MB | ~$15 | Seeed Studio |

**Recommended**: Freenove ESP32-S3-WROOM - good documentation, reliable.

### Load Cell Options

| Type | Capacity | Precision | Notes |
|------|----------|-----------|-------|
| Beam load cell (bar shape) | 5kg | 0.1g capable | Most common, easy to mount |
| Platform load cell (4x sensors) | 5kg | Higher accuracy | Requires 4 cells + combinator |

**Recommended**: Single beam load cell for prototype. Upgrade to platform type for production.

---

## Prototyping Supplies

| Component | Qty | Est. Price | Notes |
|-----------|-----|------------|-------|
| Breadboard (830 point) | 1 | $5 | For initial wiring |
| Jumper wires (F-F) | 20 | $3 | Female-to-female |
| Jumper wires (M-F) | 20 | $3 | Male-to-female |
| Micro USB to USB-C adapter | 1 | $3 | If your computer lacks USB-C |

**Subtotal**: ~$14

---

## Scale Platform Materials

| Component | Qty | Est. Price | Notes |
|-----------|-----|------------|-------|
| Acrylic plate (15x15cm, 3mm) | 1 | $3-5 | Top platform for items |
| Acrylic plate (15x15cm, 5mm) | 1 | $3-5 | Base plate |
| M4 screws + spacers | 4 sets | $3 | Mount load cell |
| Rubber feet | 4 | $2 | Anti-slip for base |

**Subtotal**: ~$11-15

---

## Camera Mounting

| Component | Qty | Est. Price | Notes |
|-----------|-----|------------|-------|
| Small tripod or gooseneck mount | 1 | $5-10 | Adjustable angle |
| 3D printed bracket | 1 | $0-5 | Optional, custom design |
| Zip ties | 5 | $1 | Quick mounting |

**Subtotal**: ~$6-16

---

## LED Lighting (External, Separate Power)

| Component | Qty | Est. Price | Notes |
|-----------|-----|------------|-------|
| LED strip (white, 30cm) | 1 | $5 | 5V USB powered (own plug) |
| Diffuser panel | 1 | $3 | Optional, reduces glare |

**Subtotal**: ~$8

**Note**: LED strip uses its own USB power. No extra power supply needed for ESP32 system.

---

## Tools Required

| Tool | Notes |
|------|-------|
| Soldering iron | For load cell wires (optional if pre-wired) |
| Multimeter | Verify connections |
| Screwdriver set | Assembly |
| Wire strippers | If soldering needed |

---

## Total Estimated Cost

| Category | Cost (USD) |
|----------|------------|
| Essential components | $27-45 |
| Prototyping supplies | $14 |
| Scale platform | $11-15 |
| Camera mounting | $6-16 |
| LED lighting | $8 |
| **Total** | **$66-98** |

---

## Where to Buy

### Fast Shipping (US)
- **Amazon** - ESP32-S3-CAM, HX711, load cells
- **Adafruit** - Quality components, good docs
- **SparkFun** - Premium components
- **Digi-Key** - Electronics components

### Budget (Longer Shipping)
- **AliExpress** - Cheapest, 2-4 weeks shipping
- **Banggood** - Similar to AliExpress

### Specialty
- **Seeed Studio** - XIAO series boards
- **Mouser** - Industrial grade components

---

## Quick Start Kit (Minimum Order)

For fastest prototyping, order these first:

| Item | Source | Price |
|------|--------|-------|
| Freenove ESP32-S3-WROOM CAM Board | Amazon | $20 |
| HX711 + 5kg Load Cell Kit | Amazon | $10 |
| Jumper Wire Kit | Amazon | $6 |

**Minimum total**: ~$36

This gets you started with hardware testing while you source the remaining parts.

---

## Sample Amazon Search Terms

```
"ESP32-S3 CAM PSRAM"
"HX711 load cell 5kg kit"
"breadboard jumper wire kit"
"5V LED strip USB white"
```

---

## Power Requirements

```
USB (5V)
   │
   └──► ESP32-S3-CAM (200-300mA)
            │
            ├──► Camera OV2640 (via 3.3V regulator, ~100-150mA)
            │
            └──► HX711 (via 3.3V pin, ~1.5mA)

Total: ~350-450mA (within USB 2.0 500mA limit)
```

**No extra power supply required** - single USB powers ESP32 + camera + HX711.

LED strip has its own USB plug (separate circuit, always-on).

---

## Notes

1. **Buy spares**: Get 2x HX711 modules - they're cheap and occasionally defective
2. **Load cell wiring**: Some kits include pre-soldered connectors, saves time
3. **USB-C cable**: Ensure it's a data cable, not charge-only
4. **PSRAM**: Essential for camera buffering - verify 8MB PSRAM on ESP32-S3 board
