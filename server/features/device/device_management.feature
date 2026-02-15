@api @device
Feature: Device Management
  As a device administrator
  I want to register and manage vending machine devices
  So that they can interact with the system

  Background:
    Given the API server is running
    And the database is clean

  @smoke
  Scenario: Register a new device
    When I register a device with the following details:
      | machine_id | name              | location        |
      | DEVICE-001 | Vending Machine 1 | Building A      |
    Then the response status should be 201
    And the response should contain field "id"
    And the response should contain field "machine_id" with value "DEVICE-001"
    And the response should contain field "message" with value "device registered"

  Scenario: Register device with same machine ID returns existing device
    Given a device exists with machine ID "DEVICE-001"
    When I register a device with the following details:
      | machine_id | name              | location        |
      | DEVICE-001 | Vending Machine 1 | Building A      |
    Then the response status should be 200
    And the response should contain field "message" with value "device already registered"

  Scenario: Device can retrieve active SKUs for ML model
    Given the following SKUs exist:
      | code      | name        | price_cents | weight_grams | weight_tolerance |
      | APPLE-001 | Fuji Apple  | 250         | 150          | 10               |
      | APPLE-002 | Gala Apple  | 230         | 140          | 10               |
    When I send a GET request to "/api/v1/device/skus"
    Then the response status should be 200
    And the response should contain 2 SKUs
    And each SKU should have fields "code,name,weight_grams,weight_tolerance"

  @validation
  Scenario: Reject device registration with empty machine ID
    When I register a device with the following details:
      | machine_id | name              | location        |
      |            | Vending Machine 1 | Building A      |
    Then the response status should be 400
