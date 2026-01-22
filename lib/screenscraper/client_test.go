package screenscraper

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// testdataPath returns the absolute path to the testdata directory.
func testdataPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get test file path")
	}
	return filepath.Join(filepath.Dir(filename), "testdata")
}

// loadFixture reads a JSON fixture file from testdata/.
func loadFixture(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(testdataPath(t), filename))
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", filename, err)
	}
	return data
}

// endpointFixtures maps API endpoints to their fixture files.
var endpointFixtures = map[string]string{
	// Game endpoints
	"/jeuInfos.php":     "game_info.json",
	"/jeuRecherche.php": "search_games.json",

	// User/Status endpoints
	"/ssuserInfos.php":  "user_info.json",
	"/ssinfraInfos.php": "infra_info.json",

	// Reference data lists
	"/systemesListe.php":        "list_systems.json",
	"/genresListe.php":          "list_genres.json",
	"/languesListe.php":         "list_languages.json",
	"/regionsListe.php":         "list_regions.json",
	"/classificationsListe.php": "list_classifications.json",
	"/famillesListe.php":        "list_families.json",
	"/mediasJeuListe.php":       "list_game_media_types.json",
	"/mediasSystemeListe.php":   "list_system_media_types.json",
	"/nbJoueursListe.php":       "list_player_counts.json",
	"/userlevelsListe.php":      "list_user_levels.json",
	"/romTypesListe.php":        "list_rom_types.json",
	"/supportTypesListe.php":    "list_support_types.json",
	"/infosJeuListe.php":        "list_game_info_types.json",
	"/infosRomListe.php":        "list_rom_info_types.json",
}

// newMockServer creates an httptest.Server that routes requests to fixtures.
func newMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for error simulation query param
		if errCode := r.URL.Query().Get("error"); errCode != "" {
			code := 404
			if _, err := fmt.Sscanf(errCode, "%d", &code); err == nil {
				w.WriteHeader(code)
				return
			}
		}

		if fixture, ok := endpointFixtures[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(loadFixture(t, fixture))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})
	return httptest.NewServer(handler)
}

// newTestClient creates a ScreenscraperClient pointing to the given server URL.
func newTestClient(t *testing.T, serverURL string) *ScreenscraperClient {
	t.Helper()
	inner, err := NewClientWithResponses(
		serverURL,
		WithRequestEditorFn(credentialEditor("testdev", "testpass", "testsoft", "testuser", "testuserpass")),
	)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}
	return &ScreenscraperClient{ClientWithResponses: inner}
}

func TestGetGameInfo_Success(t *testing.T) {
	server := newMockServer(t)
	defer server.Close()

	client := newTestClient(t, server.URL)
	ctx := context.Background()

	resp, err := client.GetGameInfoWithResponse(ctx, &GetGameInfoParams{
		GameID: "2138",
	})
	if err != nil {
		t.Fatalf("GetGameInfoWithResponse() error = %v", err)
	}

	if !IsSuccess(resp) {
		t.Errorf("Expected success response, got status %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected JSON200 to be populated")
	}

	if resp.JSON200.Header.Success != "true" {
		t.Errorf("Expected Header.Success = 'true', got %v", resp.JSON200.Header.Success)
	}

	if resp.JSON200.Response.Game.Id == "" {
		t.Fatal("Expected Response.Game to be populated")
	}

	game := resp.JSON200.Response.Game
	if game.Id != "2138" {
		t.Errorf("Expected Game.Id = '2138', got %v", game.Id)
	}
}

func TestSearchGames_Success(t *testing.T) {
	server := newMockServer(t)
	defer server.Close()

	client := newTestClient(t, server.URL)
	ctx := context.Background()

	resp, err := client.SearchGamesWithResponse(ctx, &SearchGamesParams{
		SearchQuery: "zelda",
	})
	if err != nil {
		t.Fatalf("SearchGamesWithResponse() error = %v", err)
	}

	if !IsSuccess(resp) {
		t.Errorf("Expected success response, got status %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected JSON200 to be populated")
	}

	if resp.JSON200.Header.Success != "true" {
		t.Errorf("Expected Header.Success = 'true', got %v", resp.JSON200.Header.Success)
	}

	games := resp.JSON200.Response.Games
	if len(games) == 0 {
		t.Error("Expected at least one game in search results")
	}

	// Verify first game has expected structure
	if games[0].Id == "" {
		t.Error("Expected first game to have an Id")
	}
}

// mockResponse implements the Response interface for testing response helpers.
type mockResponse struct {
	statusCode int
}

func (m mockResponse) StatusCode() int {
	return m.statusCode
}

func TestResponseHelpers_StatusCodes(t *testing.T) {
	tests := []struct {
		name              string
		statusCode        int
		wantSuccess       bool
		wantNotFound      bool
		wantQuotaExceeded bool
		wantRateLimited   bool
		wantServerBusy    bool
		wantInvalidCreds  bool
		wantAPILocked     bool
		wantBlacklisted   bool
	}{
		{
			name:        "200 OK",
			statusCode:  200,
			wantSuccess: true,
		},
		{
			name:        "201 Created",
			statusCode:  201,
			wantSuccess: true,
		},
		{
			name:         "404 Not Found",
			statusCode:   404,
			wantNotFound: true,
		},
		{
			name:            "429 Rate Limited",
			statusCode:      429,
			wantRateLimited: true,
		},
		{
			name:              "430 Quota Exceeded",
			statusCode:        430,
			wantQuotaExceeded: true,
		},
		{
			name:              "431 Quota KO Exceeded",
			statusCode:        431,
			wantQuotaExceeded: true,
		},
		{
			name:           "401 Server Busy",
			statusCode:     401,
			wantServerBusy: true,
		},
		{
			name:             "403 Invalid Credentials",
			statusCode:       403,
			wantInvalidCreds: true,
		},
		{
			name:          "423 API Locked",
			statusCode:    423,
			wantAPILocked: true,
		},
		{
			name:            "426 Blacklisted",
			statusCode:      426,
			wantBlacklisted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := mockResponse{statusCode: tt.statusCode}

			if got := IsSuccess(r); got != tt.wantSuccess {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.wantSuccess)
			}
			if got := IsNotFound(r); got != tt.wantNotFound {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.wantNotFound)
			}
			if got := IsQuotaExceeded(r); got != tt.wantQuotaExceeded {
				t.Errorf("IsQuotaExceeded() = %v, want %v", got, tt.wantQuotaExceeded)
			}
			if got := IsRateLimited(r); got != tt.wantRateLimited {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.wantRateLimited)
			}
			if got := IsServerBusy(r); got != tt.wantServerBusy {
				t.Errorf("IsServerBusy() = %v, want %v", got, tt.wantServerBusy)
			}
			if got := IsInvalidCredentials(r); got != tt.wantInvalidCreds {
				t.Errorf("IsInvalidCredentials() = %v, want %v", got, tt.wantInvalidCreds)
			}
			if got := IsAPILocked(r); got != tt.wantAPILocked {
				t.Errorf("IsAPILocked() = %v, want %v", got, tt.wantAPILocked)
			}
			if got := IsBlacklisted(r); got != tt.wantBlacklisted {
				t.Errorf("IsBlacklisted() = %v, want %v", got, tt.wantBlacklisted)
			}
		})
	}
}

func TestGetUserInfo_Success(t *testing.T) {
	server := newMockServer(t)
	defer server.Close()

	client := newTestClient(t, server.URL)
	ctx := context.Background()

	resp, err := client.GetUserInfoWithResponse(ctx)
	if err != nil {
		t.Fatalf("GetUserInfoWithResponse() error = %v", err)
	}

	if !IsSuccess(resp) {
		t.Errorf("Expected success response, got status %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected JSON200 to be populated")
	}

	if resp.JSON200.Header.Success != "true" {
		t.Errorf("Expected Header.Success = 'true', got %v", resp.JSON200.Header.Success)
	}
}

func TestGetInfraInfo_Success(t *testing.T) {
	server := newMockServer(t)
	defer server.Close()

	client := newTestClient(t, server.URL)
	ctx := context.Background()

	resp, err := client.GetInfraInfoWithResponse(ctx)
	if err != nil {
		t.Fatalf("GetInfraInfoWithResponse() error = %v", err)
	}

	if !IsSuccess(resp) {
		t.Errorf("Expected success response, got status %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		t.Fatal("Expected JSON200 to be populated")
	}

	if resp.JSON200.Header.Success != "true" {
		t.Errorf("Expected Header.Success = 'true', got %v", resp.JSON200.Header.Success)
	}
}

func TestListEndpoints_Success(t *testing.T) {
	server := newMockServer(t)
	defer server.Close()

	client := newTestClient(t, server.URL)
	ctx := context.Background()

	tests := []struct {
		name     string
		callFunc func() (Response, error)
	}{
		{
			name: "ListSystems",
			callFunc: func() (Response, error) {
				return client.ListSystemsWithResponse(ctx)
			},
		},
		{
			name: "ListGenres",
			callFunc: func() (Response, error) {
				return client.ListGenresWithResponse(ctx)
			},
		},
		{
			name: "ListLanguages",
			callFunc: func() (Response, error) {
				return client.ListLanguagesWithResponse(ctx)
			},
		},
		{
			name: "ListRegions",
			callFunc: func() (Response, error) {
				return client.ListRegionsWithResponse(ctx)
			},
		},
		{
			name: "ListClassifications",
			callFunc: func() (Response, error) {
				return client.ListClassificationsWithResponse(ctx)
			},
		},
		{
			name: "ListFamilies",
			callFunc: func() (Response, error) {
				return client.ListFamiliesWithResponse(ctx)
			},
		},
		{
			name: "ListGameMediaTypes",
			callFunc: func() (Response, error) {
				return client.ListGameMediaTypesWithResponse(ctx)
			},
		},
		{
			name: "ListSystemMediaTypes",
			callFunc: func() (Response, error) {
				return client.ListSystemMediaTypesWithResponse(ctx)
			},
		},
		{
			name: "ListPlayerCounts",
			callFunc: func() (Response, error) {
				return client.ListPlayerCountsWithResponse(ctx)
			},
		},
		{
			name: "ListUserLevels",
			callFunc: func() (Response, error) {
				return client.ListUserLevelsWithResponse(ctx)
			},
		},
		{
			name: "ListRomTypes",
			callFunc: func() (Response, error) {
				return client.ListRomTypesWithResponse(ctx)
			},
		},
		{
			name: "ListSupportTypes",
			callFunc: func() (Response, error) {
				return client.ListSupportTypesWithResponse(ctx)
			},
		},
		{
			name: "ListGameInfoTypes",
			callFunc: func() (Response, error) {
				return client.ListGameInfoTypesWithResponse(ctx)
			},
		},
		{
			name: "ListRomInfoTypes",
			callFunc: func() (Response, error) {
				return client.ListRomInfoTypesWithResponse(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.callFunc()
			if err != nil {
				t.Fatalf("%s() error = %v", tt.name, err)
			}

			if !IsSuccess(resp) {
				t.Errorf("Expected success response, got status %d", resp.StatusCode())
			}
		})
	}
}

func TestGetGameInfo_NotFound(t *testing.T) {
	// Create a mock server that returns 404 for this test
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := newTestClient(t, server.URL)
	ctx := context.Background()

	resp, err := client.GetGameInfoWithResponse(ctx, &GetGameInfoParams{
		GameID: "nonexistent",
	})
	if err != nil {
		t.Fatalf("GetGameInfoWithResponse() error = %v", err)
	}

	if IsSuccess(resp) {
		t.Error("Expected non-success response for 404")
	}

	if !IsNotFound(resp) {
		t.Errorf("Expected IsNotFound() = true, got status %d", resp.StatusCode())
	}

	if resp.JSON200 != nil {
		t.Error("Expected JSON200 to be nil for 404 response")
	}
}
