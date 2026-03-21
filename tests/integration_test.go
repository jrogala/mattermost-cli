package tests

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/cucumber/godog"
)

var (
	globalEnv *TestEnvironment
	fresh     = flag.Bool("fresh", false, "force fresh containers (stop and recreate)")
)

func TestMain(m *testing.M) {
	flag.Parse()

	// Disable Ryuk — we handle cleanup ourselves via idle killer
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	// -fresh: kill existing containers before starting
	if *fresh {
		log.Println("Fresh mode: stopping existing containers...")
		_ = exec.Command("docker", "stop", mmName, pgName).Run()
		_ = exec.Command("docker", "rm", mmName, pgName).Run()
		_ = exec.Command("docker", "network", "rm", networkName).Run()
	}

	// Isolate config dir so login tests don't touch real config
	tmpDir, err := os.MkdirTemp("", "mm-cli-test-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	ctx := context.Background()
	globalEnv, err = NewTestEnvironment(ctx)
	if err != nil {
		log.Fatalf("Failed to start test environment: %v", err)
	}

	code := m.Run()

	// Don't terminate containers — schedule idle shutdown instead
	globalEnv.ScheduleIdleShutdown()

	os.Exit(code)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero exit status from godog")
	}
}
