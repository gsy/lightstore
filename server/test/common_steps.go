package test

import (
	"fmt"
	"strings"

	"github.com/vending-machine/server/test/support"
)

// Common step definitions used across all features

func theAPIServerIsRunning() error {
	testContext.Server = support.StartTestServer(testContext.DBPool)
	return nil
}

func theDatabaseIsClean() error {
	return testContext.CleanDatabase()
}

func iSendRequestTo(method, path string) error {
	// Replace placeholders in path
	path = replacePlaceholders(path)
	return testContext.SendRequest(method, path, nil)
}

func theResponseStatusShouldBe(expectedStatus int) error {
	if testContext.LastResponse.StatusCode != expectedStatus {
		return fmt.Errorf("expected status %d, got %d. Body: %s",
			expectedStatus, testContext.LastResponse.StatusCode, string(testContext.LastBody))
	}
	return nil
}

func theResponseShouldContainField(field string) error {
	response, err := testContext.GetResponseJSON()
	if err != nil {
		return err
	}

	if _, exists := response[field]; !exists {
		return fmt.Errorf("field %s not found in response: %v", field, response)
	}

	return nil
}

func theResponseShouldContainFieldWithValue(field, expectedValue string) error {
	response, err := testContext.GetResponseJSON()
	if err != nil {
		return err
	}

	actualValue, exists := response[field]
	if !exists {
		return fmt.Errorf("field %s not found in response", field)
	}

	if fmt.Sprint(actualValue) != expectedValue {
		return fmt.Errorf("field %s: expected %s, got %v", field, expectedValue, actualValue)
	}

	return nil
}

func theResponseFieldShouldBe(field, expectedValue string) error {
	value, err := testContext.GetNestedField(field)
	if err != nil {
		return err
	}

	actualValue := fmt.Sprint(value)
	if actualValue != expectedValue {
		return fmt.Errorf("field %s: expected %s, got %s", field, expectedValue, actualValue)
	}

	return nil
}

func theResponseShouldContainError(expectedError string) error {
	response, err := testContext.GetResponseJSON()
	if err != nil {
		return err
	}

	errorMsg, exists := response["error"]
	if !exists {
		return fmt.Errorf("no error field in response: %v", response)
	}

	errorStr := fmt.Sprint(errorMsg)
	if !strings.Contains(strings.ToLower(errorStr), strings.ToLower(expectedError)) {
		return fmt.Errorf("expected error containing %q, got %q", expectedError, errorStr)
	}

	return nil
}

// replacePlaceholders replaces {placeholder} with actual values from test context
func replacePlaceholders(path string) string {
	// Replace {sku_id} with the last created SKU ID
	if strings.Contains(path, "{sku_id}") {
		for _, id := range testContext.CreatedSKUs {
			path = strings.Replace(path, "{sku_id}", id, 1)
			break
		}
	}

	// Replace {device_id} with the last created device ID
	if strings.Contains(path, "{device_id}") {
		for _, id := range testContext.CreatedDevices {
			path = strings.Replace(path, "{device_id}", id, 1)
			break
		}
	}

	// Replace {session_id} with the last created session ID
	if strings.Contains(path, "{session_id}") {
		for _, id := range testContext.CreatedSessions {
			path = strings.Replace(path, "{session_id}", id, 1)
			break
		}
	}

	return path
}
