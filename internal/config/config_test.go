// config_test.go
package config

import (
	"encoding/json"
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"testing" // Import the testing package

	"github.com/stretchr/testify/assert" // Using testify for clearer assertions
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary config file
func createTempConfigFile(t *testing.T, dir string, content Config) string {
	t.Helper() // Marks this as a test helper function

	filePath := filepath.Join(dir, configFileName) // Use the same constant
	data, err := json.MarshalIndent(content, "", "  ")
	require.NoError(t, err, "Failed to marshal config for test setup") // Use require to stop if setup fails

	err = os.WriteFile(filePath, data, 0644) // Write the file
	require.NoError(t, err, "Failed to write temp config file")

	return filePath
}

func TestRead(t *testing.T) {
	// --- Test Setup ---
	// We need the real home directory path to know where `Read` *will* look.
	// This highlights the difficulty in testing the unmodified function perfectly.
	currentUser, err := user.Current()
	require.NoError(t, err, "Failed to get current user for test setup")
	expectedPath := filepath.Join(currentUser.HomeDir, configFileName)

	// Backup existing config if it exists, and ensure cleanup
	var backupData []byte
	var originalExists bool
	if _, err := os.Stat(expectedPath); err == nil {
		originalExists = true
		backupData, err = os.ReadFile(expectedPath)
		require.NoError(t, err, "Failed to backup existing config file")
		// Defer restoration *before* potential deletion/overwrite
		defer func() {
			t.Logf("Restoring original config file at %s", expectedPath)
			err := os.WriteFile(expectedPath, backupData, 0644)
			assert.NoError(t, err, "Failed to restore original config file")
		}()
	} else {
		// Ensure the file is removed if it didn't exist originally
		defer func() {
			if !originalExists {
				t.Logf("Removing test config file from real home dir: %s", expectedPath)
				os.Remove(expectedPath) // Attempt removal, ignore error if it's already gone
			}
		}()
	}
	// --- End Test Setup ---

	t.Run("Success", func(t *testing.T) {
		// Arrange: Create a valid config file in the *actual* location `Read` expects
		t.Logf("Creating temporary test file at: %s", expectedPath)
		expectedConfig := Config{
			DBURL:           "postgres://user:pass@host:5432/db",
			CurrentUserName: "testuser",
		}
		data, err := json.MarshalIndent(expectedConfig, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(expectedPath, data, 0644)
		require.NoError(t, err)

		// Act
		cfg, err := Read()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedConfig, cfg)
		// Cleanup for this subtest (remove the file immediately) - the defer handles final cleanup
		os.Remove(expectedPath)
	})

	t.Run("File Not Found", func(t *testing.T) {
		// Arrange: Ensure the file does NOT exist where Read expects it
		os.Remove(expectedPath) // Attempt removal first

		// Act
		_, err := Read()

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, os.ErrNotExist), "Expected os.ErrNotExist, got: %v", err)
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		// Arrange: Create a file with invalid JSON where Read expects it
		t.Logf("Creating malformed test file at: %s", expectedPath)
		malformedData := []byte(`{"db_url": "missing_quote, }`)
		err := os.WriteFile(expectedPath, malformedData, 0644)
		require.NoError(t, err)

		// Act
		_, err = Read()

		// Assert
		assert.Error(t, err)
		var syntaxError *json.SyntaxError
		assert.ErrorAs(t, err, &syntaxError, "Expected a json.SyntaxError")

		// Cleanup for this subtest
		os.Remove(expectedPath)
	})
}

func TestSetUser_CurrentImplementation(t *testing.T) {
	// This test demonstrates the bugs in the current SetUser implementation

	// Arrange
	cfg := &Config{DBURL: "initial_db", CurrentUserName: "initial_user"}
	tempDir := t.TempDir() // Create a temporary directory

	// Create a dummy file in the *temp* directory (which SetUser won't find)
	dummyHomePath := createTempConfigFile(t, tempDir, *cfg)
	t.Logf("Created dummy config in temp dir: %s", dummyHomePath)

	// Create a *readable* file in the *current working directory*
	// because that's where the buggy SetUser tries to open.
	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetFilePath := filepath.Join(cwd, configFileName)
	t.Logf("Creating readable file where SetUser incorrectly looks: %s", targetFilePath)
	// Create an empty readable file
	file, err := os.Create(targetFilePath)
	require.NoError(t, err)
	file.Close() // Close immediately after creation

	// Ensure cleanup of the file in the CWD
	defer func() {
		t.Logf("Removing test file from CWD: %s", targetFilePath)
		os.Remove(targetFilePath)
	}()

	// Act
	err = cfg.SetUser("new_user")

	// Assert
	assert.Error(t, err, "Expected an error because SetUser opens read-only and tries to write")
	// The specific error might vary slightly by OS (e.g., bad file descriptor, permission denied on write)
	// but it *should* be an error related to the write operation failing.
	t.Logf("SetUser returned error (expected): %v", err)

	// Also assert that the original config struct *was* modified in memory
	assert.Equal(t, "new_user", cfg.CurrentUserName)

	// We could optionally read `targetFilePath` to confirm it wasn't written to,
	// but the error assertion is the primary goal here.
}

/*
// Example of how you might test a *corrected* SetUser
// THIS REQUIRES MODIFYING THE ORIGINAL SetUser FUNCTION
func TestSetUser_Corrected(t *testing.T) {
	// --- THIS TEST ASSUMES SetUser IS CORRECTED ---
	// 1. It calculates the path using user.Current() + filepath.Join
	// 2. It opens the file using os.OpenFile with write/create/truncate flags

	// Arrange
	tempHomeDir := t.TempDir()
	t.Setenv("HOME", tempHomeDir) // Try to influence user.Current() - **Platform Dependent!**
	// On Windows, you might need to set USERPROFILE

	// Verify the env var worked (optional, might fail in some test runners/OSs)
	currentUser, err := user.Current()
	if err != nil || currentUser.HomeDir != tempHomeDir {
		 t.Skipf("Skipping test: Could not reliably override home directory. user.Current().HomeDir = %s", currentUser.HomeDir)
	}


	initialConfig := Config{DBURL: "initial_db", CurrentUserName: "initial_user"}
	configPath := createTempConfigFile(t, tempHomeDir, initialConfig)
	t.Logf("Created initial config at mocked home: %s", configPath)

	cfgToModify := initialConfig // Work on a copy

	// Act
	err = cfgToModify.SetUser("new_test_user") // Call the *corrected* SetUser

	// Assert
	require.NoError(t, err, "Corrected SetUser failed") // Expect no error now

	// Verify the file content was updated
	updatedData, err := os.ReadFile(configPath)
	require.NoError(t, err, "Failed to read back updated config file")

	var updatedConfig Config
	err = json.Unmarshal(updatedData, &updatedConfig)
	require.NoError(t, err, "Failed to unmarshal updated config file")

	assert.Equal(t, "new_test_user", updatedConfig.CurrentUserName)
	assert.Equal(t, initialConfig.DBURL, updatedConfig.DBURL, "DBURL should not have changed")

}
*/
