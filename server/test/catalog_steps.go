package test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Catalog-specific step definitions

func iCreateSKUWithDetails(table *godog.Table) error {
	if len(table.Rows) < 2 {
		return fmt.Errorf("table must have at least 2 rows (header + data)")
	}

	row := table.Rows[1]
	sku := map[string]interface{}{
		"code":         getCellValue(table, row, "code"),
		"name":         getCellValue(table, row, "name"),
		"price_cents":  parseCellInt(table, row, "price_cents"),
		"weight_grams": parseCellFloat(table, row, "weight_grams"),
		"currency":     "USD",
	}

	// Optional fields
	if tolerance := getCellValue(table, row, "weight_tolerance"); tolerance != "" {
		sku["weight_tolerance"] = parseCellFloat(table, row, "weight_tolerance")
	}

	err := testContext.SendRequest("POST", "/api/v1/skus", sku)
	if err != nil {
		return err
	}

	// Store created SKU ID if successful
	if testContext.LastResponse.StatusCode == 201 {
		response, _ := testContext.GetResponseJSON()
		if id, ok := response["id"].(string); ok {
			code := getCellValue(table, row, "code")
			testContext.CreatedSKUs[code] = id
		}
	}

	return nil
}

func iCreateSKUWithFieldSetTo(field, value string) error {
	sku := map[string]interface{}{
		"code":         "TEST-001",
		"name":         "Test Product",
		"price_cents":  100,
		"weight_grams": 100.0,
		"currency":     "USD",
	}

	// Override the specified field
	switch field {
	case "code":
		sku["code"] = value
	case "name":
		sku["name"] = value
	case "price_cents":
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			sku["price_cents"] = v
		} else {
			sku["price_cents"] = value // Keep as string to trigger validation error
		}
	case "weight_grams":
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			sku["weight_grams"] = v
		} else {
			sku["weight_grams"] = value
		}
	}

	return testContext.SendRequest("POST", "/api/v1/skus", sku)
}

func aSKUExistsWithCode(code string) error {
	sku := map[string]interface{}{
		"code":         code,
		"name":         "Test Product",
		"price_cents":  100,
		"weight_grams": 100.0,
		"currency":     "USD",
	}

	err := testContext.SendRequest("POST", "/api/v1/skus", sku)
	if err != nil {
		return err
	}

	if testContext.LastResponse.StatusCode != 201 {
		return fmt.Errorf("failed to create SKU: status %d", testContext.LastResponse.StatusCode)
	}

	// Store created SKU ID
	response, _ := testContext.GetResponseJSON()
	if id, ok := response["id"].(string); ok {
		testContext.CreatedSKUs[code] = id
	}

	return nil
}

func theFollowingSKUsExist(table *godog.Table) error {
	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header
		}

		sku := map[string]interface{}{
			"code":         getCellValue(table, row, "code"),
			"name":         getCellValue(table, row, "name"),
			"price_cents":  parseCellInt(table, row, "price_cents"),
			"weight_grams": parseCellFloat(table, row, "weight_grams"),
			"currency":     "USD",
		}

		// Optional fields
		if tolerance := getCellValue(table, row, "weight_tolerance"); tolerance != "" {
			sku["weight_tolerance"] = parseCellFloat(table, row, "weight_tolerance")
		}

		err := testContext.SendRequest("POST", "/api/v1/skus", sku)
		if err != nil {
			return err
		}

		if testContext.LastResponse.StatusCode != 201 {
			return fmt.Errorf("failed to create SKU %s: status %d",
				getCellValue(table, row, "code"), testContext.LastResponse.StatusCode)
		}

		// Store created SKU ID
		response, _ := testContext.GetResponseJSON()
		if id, ok := response["id"].(string); ok {
			code := getCellValue(table, row, "code")
			testContext.CreatedSKUs[code] = id
		}
	}

	return nil
}

func theResponseShouldContainSKUs(count int) error {
	response, err := testContext.GetResponseJSON()
	if err != nil {
		return err
	}

	skus, ok := response["skus"].([]interface{})
	if !ok {
		return fmt.Errorf("skus field not found or not an array")
	}

	if len(skus) != count {
		return fmt.Errorf("expected %d SKUs, got %d", count, len(skus))
	}

	return nil
}

func eachSKUShouldHaveFields(fields string) error {
	response, err := testContext.GetResponseJSON()
	if err != nil {
		return err
	}

	skus, ok := response["skus"].([]interface{})
	if !ok {
		return fmt.Errorf("skus field not found or not an array")
	}

	requiredFields := strings.Split(fields, ",")
	for i, sku := range skus {
		skuMap, ok := sku.(map[string]interface{})
		if !ok {
			return fmt.Errorf("SKU at index %d is not an object", i)
		}

		for _, field := range requiredFields {
			field = strings.TrimSpace(field)
			if _, exists := skuMap[field]; !exists {
				return fmt.Errorf("SKU at index %d missing field %s", i, field)
			}
		}
	}

	return nil
}

// Helper functions

func getCellValue(table *godog.Table, row *godog.TableRow, columnName string) string {
	for i, cell := range table.Rows[0].Cells {
		if cell.Value == columnName && i < len(row.Cells) {
			return row.Cells[i].Value
		}
	}
	return ""
}

func parseCellInt(table *godog.Table, row *godog.TableRow, columnName string) int64 {
	value := getCellValue(table, row, columnName)
	if value == "" {
		return 0
	}
	v, _ := strconv.ParseInt(value, 10, 64)
	return v
}

func parseCellFloat(table *godog.Table, row *godog.TableRow, columnName string) float64 {
	value := getCellValue(table, row, columnName)
	if value == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(value, 64)
	return v
}
