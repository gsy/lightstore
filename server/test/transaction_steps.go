package test

import (
	"fmt"

	"github.com/cucumber/godog"
)

// Transaction-specific step definitions

func iStartSessionOnDevice(machineID string) error {
	session := map[string]interface{}{
		"machine_id": machineID,
	}

	err := testContext.SendRequest("POST", "/api/v1/session/start", session)
	if err != nil {
		return err
	}

	// Store created session ID if successful
	if testContext.LastResponse.StatusCode == 201 {
		response, _ := testContext.GetResponseJSON()
		if sessionID, ok := response["session_id"].(string); ok {
			testContext.CreatedSessions["current"] = sessionID
		}
	}

	return nil
}

func anActiveSessionExistsOnDevice(machineID string) error {
	// Ensure device exists
	if _, exists := testContext.CreatedDevices[machineID]; !exists {
		if err := aDeviceExistsWithMachineID(machineID); err != nil {
			return err
		}
	}

	// Start session
	return iStartSessionOnDevice(machineID)
}

func anActiveSessionWithItemsExistsOnDevice(machineID string) error {
	// Create active session
	if err := anActiveSessionExistsOnDevice(machineID); err != nil {
		return err
	}

	// Add some items
	sessionID := testContext.CreatedSessions["current"]
	deviceID := testContext.CreatedDevices[machineID]

	detection := map[string]interface{}{
		"device_id":  deviceID,
		"session_id": sessionID,
		"items": []map[string]interface{}{
			{
				"sku":        "APPLE-001",
				"confidence": 0.95,
			},
			{
				"sku":        "APPLE-002",
				"confidence": 0.92,
			},
		},
	}

	err := testContext.SendRequest("POST", "/api/v1/device/detection", detection)
	if err != nil {
		return err
	}

	if testContext.LastResponse.StatusCode != 200 {
		return fmt.Errorf("failed to add items to session: status %d, body: %s",
			testContext.LastResponse.StatusCode, string(testContext.LastBody))
	}

	return nil
}

func aCompletedSessionExistsOnDevice(machineID string) error {
	// Create session with items
	if err := anActiveSessionWithItemsExistsOnDevice(machineID); err != nil {
		return err
	}

	// Confirm the session
	sessionID := testContext.CreatedSessions["current"]
	confirm := map[string]interface{}{
		"payment_ref": "TEST-PAY-123",
	}

	err := testContext.SendRequest("POST", fmt.Sprintf("/api/v1/session/%s/confirm", sessionID), confirm)
	if err != nil {
		return err
	}

	if testContext.LastResponse.StatusCode != 200 {
		return fmt.Errorf("failed to confirm session: status %d", testContext.LastResponse.StatusCode)
	}

	return nil
}

func iSubmitDetectionsToSession(table *godog.Table) error {
	sessionID := testContext.CreatedSessions["current"]
	if sessionID == "" {
		return fmt.Errorf("no active session found")
	}

	// Find device ID for this session
	var deviceID string
	for _, id := range testContext.CreatedDevices {
		deviceID = id
		break
	}

	var items []map[string]interface{}
	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header
		}

		item := map[string]interface{}{
			"sku":        getCellValue(table, row, "sku"),
			"confidence": parseCellFloat(table, row, "confidence"),
		}
		items = append(items, item)
	}

	detection := map[string]interface{}{
		"device_id":  deviceID,
		"session_id": sessionID,
		"items":      items,
	}

	return testContext.SendRequest("POST", "/api/v1/device/detection", detection)
}

func iSubmitDetectionsToSessionID(sessionID string) error {
	// Find device ID
	var deviceID string
	for _, id := range testContext.CreatedDevices {
		deviceID = id
		break
	}

	detection := map[string]interface{}{
		"device_id":  deviceID,
		"session_id": sessionID,
		"items": []map[string]interface{}{
			{
				"sku":        "APPLE-001",
				"confidence": 0.95,
			},
		},
	}

	return testContext.SendRequest("POST", "/api/v1/device/detection", detection)
}

func iConfirmSessionWithPaymentRef(paymentRef string) error {
	sessionID := testContext.CreatedSessions["current"]
	if sessionID == "" {
		return fmt.Errorf("no active session found")
	}

	confirm := map[string]interface{}{
		"payment_ref": paymentRef,
	}

	return testContext.SendRequest("POST", fmt.Sprintf("/api/v1/session/%s/confirm", sessionID), confirm)
}

func iCancelSessionWithReason(reason string) error {
	sessionID := testContext.CreatedSessions["current"]
	if sessionID == "" {
		return fmt.Errorf("no active session found")
	}

	cancel := map[string]interface{}{
		"reason": reason,
	}

	return testContext.SendRequest("POST", fmt.Sprintf("/api/v1/session/%s/cancel", sessionID), cancel)
}

func theResponseShouldContainItems(count int) error {
	response, err := testContext.GetResponseJSON()
	if err != nil {
		return err
	}

	items, ok := response["items"].([]interface{})
	if !ok {
		return fmt.Errorf("items field not found or not an array")
	}

	if len(items) != count {
		return fmt.Errorf("expected %d items, got %d", count, len(items))
	}

	return nil
}

func theTotalShouldBeCents(cents int) error {
	response, err := testContext.GetResponseJSON()
	if err != nil {
		return err
	}

	totalCents, ok := response["total_cents"].(float64)
	if !ok {
		return fmt.Errorf("total_cents field not found or not a number")
	}

	if int(totalCents) != cents {
		return fmt.Errorf("expected total %d cents, got %d cents", cents, int(totalCents))
	}

	return nil
}

func theResponseShouldContainItems() error {
	response, err := testContext.GetResponseJSON()
	if err != nil {
		return err
	}

	items, ok := response["items"].([]interface{})
	if !ok {
		return fmt.Errorf("items field not found or not an array")
	}

	if len(items) == 0 {
		return fmt.Errorf("expected items in response, got none")
	}

	return nil
}
