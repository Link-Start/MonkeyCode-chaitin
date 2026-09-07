package agentconfig

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chaitin/MonkeyCode/monkeyai/backend/internal/identity"
	"github.com/chaitin/MonkeyCode/monkeyai/backend/internal/model"
	"github.com/chaitin/MonkeyCode/monkeyai/backend/internal/setting"
	"github.com/go-chi/chi/v5"
)

type settingReaderStub struct {
	config setting.Config
}

func TestConfigEndpointSupportsETagAndHasNoSSEEndpoint(t *testing.T) {
	service := NewService(
		settingReaderStub{config: setting.Config{Settings: map[string]json.RawMessage{}}},
		modelReaderStub{},
		"https://monkeyai.example.com",
	)
	router := chi.NewRouter()
	service.RegisterAgent(router)

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/config", nil))
	if first.Code != http.StatusOK || first.Header().Get("ETag") == "" {
		t.Fatalf("status = %d, etag = %q", first.Code, first.Header().Get("ETag"))
	}

	second := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/config", nil)
	request.Header.Set("If-None-Match", first.Header().Get("ETag"))
	router.ServeHTTP(second, request)
	if second.Code != http.StatusNotModified {
		t.Fatalf("status = %d", second.Code)
	}

	events := httptest.NewRecorder()
	router.ServeHTTP(events, httptest.NewRequest(http.MethodGet, "/config/events", nil))
	if events.Code != http.StatusNotFound {
		t.Fatalf("events status = %d", events.Code)
	}
}

func (s settingReaderStub) AgentConfig(context.Context) (setting.Config, error) { return s.config, nil }

type modelReaderStub struct {
	models []model.AgentModel
}

func (s modelReaderStub) AgentModels(context.Context, string, bool) ([]model.AgentModel, error) {
	return s.models, nil
}

func TestSnapshotIsStableAndContainsNoUpstreamCredential(t *testing.T) {
	updatedAt := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	service := NewService(
		settingReaderStub{config: setting.Config{
			UpdatedAt: updatedAt,
			Settings: map[string]json.RawMessage{
				"branding": json.RawMessage(`{"product_name":"MonkeyAI"}`),
			},
		}},
		modelReaderStub{models: []model.AgentModel{{
			ID: "model-1", Model: "model-1", DisplayName: "GPT-5",
			Protocol: model.ProtocolOpenAIResponses, UpdatedAt: updatedAt,
		}}},
		"https://monkeyai.example.com/",
	)
	first, err := service.Snapshot(t.Context(), identity.User{ID: "user-1"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Snapshot(t.Context(), identity.User{ID: "user-1"})
	if err != nil {
		t.Fatal(err)
	}
	if first.Version != second.Version || first.Version == "" {
		t.Fatalf("versions = %q, %q", first.Version, second.Version)
	}
	if first.ModelGateway.BaseURL != "https://monkeyai.example.com/v1" || first.ModelGateway.Authentication != "api_key" {
		t.Fatalf("gateway = %#v", first.ModelGateway)
	}
	encoded, _ := json.Marshal(first)
	if string(encoded) == "" || strings.Contains(string(encoded), "api.openai.com") || strings.Contains(string(encoded), "upstream-secret") {
		t.Fatalf("snapshot = %s", encoded)
	}
}
