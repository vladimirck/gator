package main // Or the actual package name where fetchFeed is defined

import (
	"context"
	"testing"
	"time"
	// Add other necessary imports if fetchFeed depends on them
)

// --- Function Signature (for reference) ---
// Assume fetchFeed is defined in this package or imported correctly.
// func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error)

// --- Test Function ---

// TestFetchFeed tests the fetchFeed function with specific URLs.
func TestFetchFeed(t *testing.T) {
	// Define test cases using a slice of structs
	testCases := []struct {
		name    string // Name for the subtest
		url     string // URL to fetch
		wantErr bool   // Set to true if an error is expected for this URL
	}{
		{
			name:    "WagsLane Feed",
			url:     "https://www.wagslane.dev/index.xml",
			wantErr: false, // Expect success for this valid feed URL
		},
		{
			name:    "Multiverso Feed",
			url:     "https://multiverso.do/index.xml",
			wantErr: false, // Expect success for this valid feed URL
		},
		// --- Optional: Add more test cases ---
		// Example: Test case for an invalid or non-existent URL
		// {
		// 	name:    "Invalid URL Format",
		// 	url:     "://invalid-url", // Malformed URL
		// 	wantErr: true,           // Expect an error
		// },
		// {
		// 	name:    "Non-Existent Feed",
		// 	url:     "https://example.com/thisdoesnotexist.xml",
		// 	wantErr: true, // Expect an error (e.g., 404 Not Found)
		// },
	}

	// Iterate over the defined test cases
	for _, tc := range testCases {
		// Run each test case as a distinct subtest
		// This makes test output clearer, especially with failures.
		t.Run(tc.name, func(t *testing.T) {
			// Create a context with a timeout.
			// This prevents tests from hanging indefinitely if a network request stalls.
			// Adjust the timeout (e.g., 30 seconds) as appropriate for your use case.
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			// Ensure the context's cancel function is called when the subtest finishes.
			// This releases resources associated with the context.
			defer cancel()

			// Call the function you want to test
			feed, err := fetchFeed(ctx, tc.url)

			// --- Assertions ---

			// Check if an error occurred when one wasn't expected,
			// or if no error occurred when one *was* expected.
			if (err != nil) != tc.wantErr {
				t.Errorf("fetchFeed() error = %v, wantErr %v", err, tc.wantErr)
				// If the error expectation is wrong, no point in checking the feed object.
				return
			}

			// If no error was expected (wantErr is false), perform further checks.
			if !tc.wantErr {
				// Check that the returned feed object is not nil for successful fetches.
				if feed == nil {
					t.Errorf("fetchFeed() returned a nil feed for URL '%s', but expected a non-nil feed", tc.url)
				}
				// Optional: Add more specific assertions here if needed.
				// For example, you could check if certain fields in the feed struct are populated:
				// if feed != nil && feed.Title == "" {
				// 	t.Logf("Warning: fetchFeed() returned feed with empty title for URL '%s'", tc.url)
				// }
			}

			// Optional: If an error *was* expected (wantErr is true), you might want
			// to inspect the error type or message more closely.
			// if tc.wantErr && err != nil {
			// 	// Example: Check if the error is a specific type
			// 	// if !errors.Is(err, expectedErrorType) {
			// 	// 	t.Errorf("fetchFeed() returned error type %T, want error type %T", err, expectedErrorType)
			// 	// }
			// }
		})
	}
}

// Note: You need to have the actual `fetchFeed` function defined or imported
// in the same package for this test to run. You also need the definition
// of the `RSSFeed` struct.
