/**
 * Beverage Vending Machine Recognition System
 * ESP32-S3-CAM + HX711 Load Cell
 *
 * Phase 1: Hardware Test Firmware
 */

#include <Arduino.h>
#include "esp_camera.h"
#include "HX711.h"

// =============================================================================
// Pin Definitions
// =============================================================================

// HX711 Load Cell Amplifier
#define HX711_DT_PIN   1
#define HX711_SCK_PIN  2

// Camera pins for ESP32-S3-CAM (adjust for your specific board)
#define PWDN_GPIO_NUM    -1
#define RESET_GPIO_NUM   -1
#define XCLK_GPIO_NUM    10
#define SIOD_GPIO_NUM    40
#define SIOC_GPIO_NUM    39

#define Y9_GPIO_NUM      48
#define Y8_GPIO_NUM      11
#define Y7_GPIO_NUM      12
#define Y6_GPIO_NUM      14
#define Y5_GPIO_NUM      16
#define Y4_GPIO_NUM      18
#define Y3_GPIO_NUM      17
#define Y2_GPIO_NUM      15
#define VSYNC_GPIO_NUM   38
#define HREF_GPIO_NUM    47
#define PCLK_GPIO_NUM    13

// =============================================================================
// Global Objects
// =============================================================================

HX711 scale;

// Scale calibration factor (adjust after calibration)
float calibrationFactor = 420.0;  // Example value, needs calibration

// Weight threshold to trigger capture (grams)
const float WEIGHT_THRESHOLD = 50.0;

// Previous weight for change detection
float previousWeight = 0.0;

// =============================================================================
// Camera Setup
// =============================================================================

bool initCamera() {
    camera_config_t config;
    config.ledc_channel = LEDC_CHANNEL_0;
    config.ledc_timer = LEDC_TIMER_0;
    config.pin_d0 = Y2_GPIO_NUM;
    config.pin_d1 = Y3_GPIO_NUM;
    config.pin_d2 = Y4_GPIO_NUM;
    config.pin_d3 = Y5_GPIO_NUM;
    config.pin_d4 = Y6_GPIO_NUM;
    config.pin_d5 = Y7_GPIO_NUM;
    config.pin_d6 = Y8_GPIO_NUM;
    config.pin_d7 = Y9_GPIO_NUM;
    config.pin_xclk = XCLK_GPIO_NUM;
    config.pin_pclk = PCLK_GPIO_NUM;
    config.pin_vsync = VSYNC_GPIO_NUM;
    config.pin_href = HREF_GPIO_NUM;
    config.pin_sccb_sda = SIOD_GPIO_NUM;
    config.pin_sccb_scl = SIOC_GPIO_NUM;
    config.pin_pwdn = PWDN_GPIO_NUM;
    config.pin_reset = RESET_GPIO_NUM;
    config.xclk_freq_hz = 20000000;
    config.frame_size = FRAMESIZE_QVGA;  // 320x240 for testing
    config.pixel_format = PIXFORMAT_JPEG;
    config.grab_mode = CAMERA_GRAB_WHEN_EMPTY;
    config.fb_location = CAMERA_FB_IN_PSRAM;
    config.jpeg_quality = 12;
    config.fb_count = 2;

    // Higher resolution if PSRAM available
    if (psramFound()) {
        config.jpeg_quality = 10;
        config.fb_count = 2;
        config.grab_mode = CAMERA_GRAB_LATEST;
        Serial.println("PSRAM found, using higher quality settings");
    } else {
        config.frame_size = FRAMESIZE_QVGA;
        config.fb_location = CAMERA_FB_IN_DRAM;
        config.fb_count = 1;
        Serial.println("No PSRAM, using lower quality settings");
    }

    esp_err_t err = esp_camera_init(&config);
    if (err != ESP_OK) {
        Serial.printf("Camera init failed with error 0x%x\n", err);
        return false;
    }

    Serial.println("Camera initialized successfully");
    return true;
}

// =============================================================================
// Scale Setup
// =============================================================================

bool initScale() {
    scale.begin(HX711_DT_PIN, HX711_SCK_PIN);

    if (!scale.is_ready()) {
        Serial.println("HX711 not found!");
        return false;
    }

    scale.set_scale(calibrationFactor);
    scale.tare();  // Reset to zero

    Serial.println("Scale initialized and tared");
    return true;
}

// =============================================================================
// Weight Reading
// =============================================================================

float readWeight() {
    if (!scale.is_ready()) {
        return -1.0;
    }

    // Average of 5 readings for stability
    float weight = scale.get_units(5);
    return weight;
}

// =============================================================================
// Image Capture
// =============================================================================

camera_fb_t* captureImage() {
    camera_fb_t* fb = esp_camera_fb_get();
    if (!fb) {
        Serial.println("Camera capture failed");
        return nullptr;
    }

    Serial.printf("Captured image: %dx%d, %d bytes\n",
                  fb->width, fb->height, fb->len);
    return fb;
}

void releaseImage(camera_fb_t* fb) {
    if (fb) {
        esp_camera_fb_return(fb);
    }
}

// =============================================================================
// Main Setup
// =============================================================================

void setup() {
    Serial.begin(115200);
    delay(1000);

    Serial.println("\n========================================");
    Serial.println("Beverage Recognition System - Phase 1");
    Serial.println("Hardware Test Firmware");
    Serial.println("========================================\n");

    // Initialize camera
    Serial.println("[1/2] Initializing camera...");
    if (!initCamera()) {
        Serial.println("ERROR: Camera initialization failed!");
        Serial.println("Check camera ribbon cable and pin definitions.");
    }

    // Initialize scale
    Serial.println("\n[2/2] Initializing scale...");
    if (!initScale()) {
        Serial.println("ERROR: Scale initialization failed!");
        Serial.println("Check HX711 wiring: DT->GPIO1, SCK->GPIO2");
    }

    Serial.println("\n========================================");
    Serial.println("Setup complete. Monitoring weight...");
    Serial.println("Place items on scale to trigger capture.");
    Serial.println("========================================\n");
}

// =============================================================================
// Main Loop
// =============================================================================

void loop() {
    // Read current weight
    float currentWeight = readWeight();

    if (currentWeight >= 0) {
        Serial.printf("Weight: %.1f g\n", currentWeight);

        // Check if significant weight change (item placed)
        float weightChange = currentWeight - previousWeight;

        if (weightChange > WEIGHT_THRESHOLD) {
            Serial.println("\n*** Weight increase detected! ***");
            Serial.printf("Change: +%.1f g\n", weightChange);

            // Capture image
            Serial.println("Capturing image...");
            camera_fb_t* fb = captureImage();

            if (fb) {
                // In Phase 1, just log the capture
                // Later phases will do inference here
                Serial.println("Image captured successfully");
                Serial.printf("Ready for recognition (TODO: ML inference)\n");

                releaseImage(fb);
            }

            Serial.println();
        }

        previousWeight = currentWeight;
    } else {
        Serial.println("Scale read error");
    }

    delay(500);  // Read every 500ms
}
