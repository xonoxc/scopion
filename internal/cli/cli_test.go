package cli

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestExecute(t *testing.T) {
	// Test that Execute doesn't panic for valid commands
	// We'll test the commands exist and can be executed without panicking
	t.Run("commands are defined", func(t *testing.T) {
		// Just test that the commands are properly initialized
		if rootCmd.Commands() == nil {
			t.Error("Root command should have subcommands")
		}

		commands := rootCmd.Commands()
		foundStart := false
		foundVersion := false

		for _, cmd := range commands {
			if cmd.Use == "start" {
				foundStart = true
			}
			if cmd.Use == "version" {
				foundVersion = true
			}
		}

		if !foundStart {
			t.Error("Start command should be defined")
		}
		if !foundVersion {
			t.Error("Version command should be defined")
		}
	})
}

func TestStartServerSignature(t *testing.T) {
	// Test that startServer function accepts the correct parameters
	// We can't easily test the full server startup in unit tests due to dependencies,
	// but we can test that the function signature is correct and basic logic works

	t.Run("function signature", func(t *testing.T) {
		// Test with immediate timeout to avoid actual server startup
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// This will fail due to missing migrations, but we just want to test it doesn't panic
		err := startServer(ctx, "9999", false)
		if err == nil {
			t.Error("Expected error due to missing migrations directory")
		}
	})

	t.Run("demo parameter logic", func(t *testing.T) {
		// We can't test the actual demo logic without setting up the full environment,
		// but we can verify the function accepts both true and false for demo parameter
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Test with demo=false
		err1 := startServer(ctx, "9998", false)
		if err1 == nil {
			t.Error("Expected error due to missing migrations directory")
		}

		// Test with demo=true
		ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel2()

		err2 := startServer(ctx2, "9997", true)
		if err2 == nil {
			t.Error("Expected error due to missing migrations directory")
		}
	})
}

func TestScorpionArt(t *testing.T) {
	// Test that scorpionArt contains expected content
	if !strings.Contains(scorpionArt, "Scopion") {
		t.Error("scorpionArt should contain 'Scopion'")
	}

	if !strings.Contains(scorpionArt, "observability") {
		t.Error("scorpionArt should contain 'observability'")
	}

	if !strings.Contains(scorpionArt, "___") {
		t.Error("scorpionArt should contain ASCII art characters")
	}
}

func TestRootCommand(t *testing.T) {
	// Test that root command has expected properties
	if rootCmd.Use != "scopion" {
		t.Errorf("Expected root command use to be 'scopion', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short != "Scopion - Single-binary observability" {
		t.Errorf("Expected root command short to be 'Scopion - Single-binary observability', got '%s'", rootCmd.Short)
	}

	if !strings.Contains(rootCmd.Long, "Scopion") {
		t.Error("Root command long description should contain 'Scopion'")
	}
}

func TestStartCommand(t *testing.T) {
	// Test that start command has expected properties
	if startCmd.Use != "start" {
		t.Errorf("Expected start command use to be 'start', got '%s'", startCmd.Use)
	}

	if startCmd.Short != "Start the Scopion server" {
		t.Errorf("Expected start command short to be 'Start the Scopion server', got '%s'", startCmd.Short)
	}
}

func TestVersionCommand(t *testing.T) {
	// Test that version command has expected properties
	if versionCmd.Use != "version" {
		t.Errorf("Expected version command use to be 'version', got '%s'", versionCmd.Use)
	}

	if versionCmd.Short != "Print the version number" {
		t.Errorf("Expected version command short to be 'Print the version number', got '%s'", versionCmd.Short)
	}
}

func TestCommandFlags(t *testing.T) {
	// Test that start command has the expected flags
	portFlag := startCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("Expected 'port' flag to be defined")
	}

	demoFlag := startCmd.Flags().Lookup("demo")
	if demoFlag == nil {
		t.Error("Expected 'demo' flag to be defined")
	}
}
