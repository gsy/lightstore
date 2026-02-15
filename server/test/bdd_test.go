package test

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vending-machine/server/internal/platform/postgres"
	"github.com/vending-machine/server/test/support"
)

var (
	testContext *support.TestContext
	dbPool      *pgxpool.Pool
)

func TestFeatures(t *testing.T) {
	// Setup database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://vending:vending@localhost:5432/vending_test?sslmode=disable"
	}

	var err error
	dbPool, err = support.ConnectTestDB(databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer dbPool.Close()

	// Run migrations
	if err := postgres.RunMigrations(dbPool); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	testContext = support.NewTestContext()
	testContext.DBPool = dbPool

	// Lifecycle hooks
	ctx.Before(beforeScenario)
	ctx.After(afterScenario)

	// Common steps
	ctx.Step(`^the API server is running$`, theAPIServerIsRunning)
	ctx.Step(`^the database is clean$`, theDatabaseIsClean)
	ctx.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)"$`, iSendRequestTo)
	ctx.Step(`^the response status should be (\d+)$`, theResponseStatusShouldBe)
	ctx.Step(`^the response should contain field "([^"]*)"$`, theResponseShouldContainField)
	ctx.Step(`^the response should contain field "([^"]*)" with value "([^"]*)"$`, theResponseShouldContainFieldWithValue)
	ctx.Step(`^the response field "([^"]*)" should be "([^"]*)"$`, theResponseFieldShouldBe)
	ctx.Step(`^the response should contain error "([^"]*)"$`, theResponseShouldContainError)

	// Catalog steps
	ctx.Step(`^I create a SKU with the following details:$`, iCreateSKUWithDetails)
	ctx.Step(`^I create a SKU with ([^ ]+) set to (.+)$`, iCreateSKUWithFieldSetTo)
	ctx.Step(`^a SKU exists with code "([^"]*)"$`, aSKUExistsWithCode)
	ctx.Step(`^the following SKUs exist:$`, theFollowingSKUsExist)
	ctx.Step(`^the response should contain (\d+) SKUs$`, theResponseShouldContainSKUs)
	ctx.Step(`^each SKU should have fields "([^"]*)"$`, eachSKUShouldHaveFields)

	// Device steps
	ctx.Step(`^I register a device with the following details:$`, iRegisterDeviceWithDetails)
	ctx.Step(`^a device exists with machine ID "([^"]*)"$`, aDeviceExistsWithMachineID)

	// Transaction steps
	ctx.Step(`^I start a session on device "([^"]*)"$`, iStartSessionOnDevice)
	ctx.Step(`^an active session exists on device "([^"]*)"$`, anActiveSessionExistsOnDevice)
	ctx.Step(`^an active session with items exists on device "([^"]*)"$`, anActiveSessionWithItemsExistsOnDevice)
	ctx.Step(`^a completed session exists on device "([^"]*)"$`, aCompletedSessionExistsOnDevice)
	ctx.Step(`^I submit the following detections to the session:$`, iSubmitDetectionsToSession)
	ctx.Step(`^I submit detections to session "([^"]*)"$`, iSubmitDetectionsToSessionID)
	ctx.Step(`^I confirm the session with payment reference "([^"]*)"$`, iConfirmSessionWithPaymentRef)
	ctx.Step(`^I cancel the session with reason "([^"]*)"$`, iCancelSessionWithReason)
	ctx.Step(`^the response should contain (\d+) items$`, theResponseShouldContainItems)
	ctx.Step(`^the total should be (\d+) cents$`, theTotalShouldBeCents)
	ctx.Step(`^the response should contain items$`, theResponseShouldContainItems)
}

func beforeScenario(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	testContext.Reset()
	return ctx, nil
}

func afterScenario(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	if testContext.Server != nil {
		testContext.Server.Close()
	}
	return ctx, nil
}
