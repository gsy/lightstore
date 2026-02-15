@api @transaction
Feature: Shopping Session
  As a customer
  I want to complete a shopping session
  So that I can purchase items from the vending machine

  Background:
    Given the API server is running
    And the database is clean
    And a device exists with machine ID "DEVICE-001"
    And the following SKUs exist:
      | code      | name        | price_cents | weight_grams | weight_tolerance |
      | APPLE-001 | Fuji Apple  | 250         | 150          | 10               |
      | APPLE-002 | Gala Apple  | 230         | 140          | 10               |
      | BANANA-01 | Banana      | 180         | 120          | 8                |

  @smoke
  Scenario: Start a new shopping session
    When I start a session on device "DEVICE-001"
    Then the response status should be 201
    And the response should contain field "session_id"
    And the response should contain field "device_id"
    And the response should contain field "expires_at"
    And the response should contain field "message" with value "session started, place items on scale"

  Scenario: Submit item detections to session
    Given an active session exists on device "DEVICE-001"
    When I submit the following detections to the session:
      | sku       | confidence |
      | APPLE-001 | 0.95       |
      | APPLE-002 | 0.92       |
    Then the response status should be 200
    And the response should contain field "session_id"
    And the response should contain 2 items
    And the total should be 480 cents

  Scenario: Get session details
    Given an active session with items exists on device "DEVICE-001"
    When I send a GET request to "/api/v1/session/{session_id}"
    Then the response status should be 200
    And the response field "session.status" should be "active"
    And the response should contain items

  Scenario: Confirm session and complete purchase
    Given an active session with items exists on device "DEVICE-001"
    When I confirm the session with payment reference "PAY-12345"
    Then the response status should be 200
    And the response field "status" should be "completed"
    And the response should contain field "message" with value "purchase confirmed"

  Scenario: Cancel an active session
    Given an active session exists on device "DEVICE-001"
    When I cancel the session with reason "customer changed mind"
    Then the response status should be 200
    And the response field "status" should be "cancelled"
    And the response should contain field "message" with value "session cancelled"

  @error-handling
  Scenario: Cannot start session on non-existent device
    When I start a session on device "NONEXISTENT"
    Then the response status should be 404
    And the response should contain error "device not found"

  @error-handling
  Scenario: Cannot submit detection to non-existent session
    When I submit detections to session "NONEXISTENT-SESSION"
    Then the response status should be 404
    And the response should contain error "session not found"

  @error-handling
  Scenario: Cannot confirm session without items
    Given an active session exists on device "DEVICE-001"
    When I confirm the session with payment reference "PAY-12345"
    Then the response status should be 422
    And the response should contain error "no items detected"

  @error-handling
  Scenario: Cannot submit detection to completed session
    Given a completed session exists on device "DEVICE-001"
    When I submit the following detections to the session:
      | sku       | confidence |
      | APPLE-001 | 0.95       |
    Then the response status should be 422
    And the response should contain error "session not active"

  @error-handling
  Scenario: Cannot cancel completed session
    Given a completed session exists on device "DEVICE-001"
    When I cancel the session with reason "test"
    Then the response status should be 422
    And the response should contain error "session already completed"
