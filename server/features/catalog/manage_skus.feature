@api @catalog
Feature: Manage SKUs
  As a catalog manager
  I want to manage product SKUs
  So that I can maintain the product inventory

  Background:
    Given the API server is running
    And the database is clean

  @smoke
  Scenario: Create a new SKU
    When I create a SKU with the following details:
      | code        | name          | price_cents | weight_grams |
      | APPLE-001   | Fuji Apple    | 250         | 150          |
    Then the response status should be 201
    And the response should contain field "id"
    And the response should contain field "message" with value "SKU created"

  Scenario: List all SKUs
    Given the following SKUs exist:
      | code      | name        | price_cents | weight_grams |
      | APPLE-001 | Fuji Apple  | 250         | 150          |
      | APPLE-002 | Gala Apple  | 230         | 140          |
      | APPLE-003 | Pink Lady   | 280         | 160          |
    When I send a GET request to "/api/v1/skus"
    Then the response status should be 200
    And the response should contain 3 SKUs

  Scenario: Get SKU by ID
    Given a SKU exists with code "APPLE-001"
    When I send a GET request to "/api/v1/skus/{sku_id}"
    Then the response status should be 200
    And the response field "code" should be "APPLE-001"
    And the response field "active" should be "true"

  Scenario: List only active SKUs
    Given the following SKUs exist:
      | code      | name        | price_cents | weight_grams |
      | APPLE-001 | Fuji Apple  | 250         | 150          |
      | APPLE-002 | Gala Apple  | 230         | 140          |
    When I send a GET request to "/api/v1/skus/active"
    Then the response status should be 200
    And the response should contain 2 SKUs

  @error-handling
  Scenario: Reject duplicate SKU code
    Given a SKU exists with code "APPLE-001"
    When I create a SKU with the following details:
      | code        | name          | price_cents | weight_grams |
      | APPLE-001   | Another Apple | 300         | 160          |
    Then the response status should be 409
    And the response should contain error "duplicate SKU code"

  @validation
  Scenario Outline: Validate SKU fields
    When I create a SKU with <field> set to <value>
    Then the response status should be <status>

    Examples:
      | field        | value | status |
      | code         | ""    | 400    |
      | name         | ""    | 400    |
      | price_cents  | -1    | 422    |
      | weight_grams | 0     | 422    |
