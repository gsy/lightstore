package test

import (
	"fmt"

	"github.com/cucumber/godog"
)

// Device-specific step definitions

func iRegisterDeviceWithDetails(table *godog.Table) error {
	if len(table.Rows) < 2 {
		return fmt.Errorf("table must have at least 2 rows (header + data)")
	}

	row := table.Rows[1]
	device := map[string]interface{}{
		"machine_id": getCellValue(table, row, "machine_id"),
		"name":       getCellValue(table, row, "name"),
		"location":   getCellValue(table, row, "location"),
	}

	err := testContext.SendRequest("POST", "/api/v1/device/register", device)
	if err != nil {
		return err
	}

	// Store created device ID if successful
	if testContext.LastResponse.StatusCode == 201 || testContext.LastResponse.StatusCode == 200 {
		response, _ := testContext.GetResponseJSON()
		if id, ok := response["id"].(string); ok {
			machineID := getCellValue(table, row, "machine_id")
			testContext.CreatedDevices[machineID] = id
		}
	}

	return nil
}

func aDeviceExistsWithMachineID(machineID string) error {
	device := map[string]interface{}{
		"machine_id": machineID,
		"name":       "Test Device",
		"location":   "Test Location",
	}

	err := testContext.SendRequest("POST", "/api/v1/device/register", device)
	if err != nil {
		return err
	}

	if testContext.LastResponse.StatusCode != 201 && testContext.LastResponse.StatusCode != 200 {
		return fmt.Errorf("failed to create device: status %d", testContext.LastResponse.StatusCode)
	}

	// Store created device ID
	response, _ := testContext.GetResponseJSON()
	if id, ok := response["id"].(string); ok {
		testContext.CreatedDevices[machineID] = id
	}

	return nil
}
