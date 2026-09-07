package apikey

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type storeStub struct {
	keys    []Key
	hash    string
	userID  string
	authErr error
}

func (s *storeStub) Create(_ context.Context, key Key, hash string) (Key, error) {
	key.ID = "key-1"
	key.CreatedAt = time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	s.keys = append(s.keys, key)
	s.hash = hash
	return key, nil
}

func (s *storeStub) ListByUser(context.Context, string) ([]Key, error) { return s.keys, nil }
func (s *storeStub) List(context.Context, string) ([]Key, error)       { return s.keys, nil }
func (s *storeStub) Revoke(context.Context, string, string) error      { return nil }
func (s *storeStub) Authenticate(context.Context, string, string) (string, error) {
	return s.userID, s.authErr
}

func TestCreateReturnsSecretOnce(t *testing.T) {
	store := &storeStub{}
	service := NewService(store)
	service.now = func() time.Time { return time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC) }

	created, err := service.Create(t.Context(), "user-1", CreateInput{Name: " MacBook ", Scopes: []string{ScopeModelInvoke}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(created.APIKey, "mk_") || created.Prefix != created.APIKey[:keyPrefixLength] {
		t.Fatalf("key = %#v", created)
	}
	if created.Name != "MacBook" || created.UserID != "user-1" {
		t.Fatalf("metadata = %#v", created.Key)
	}
	if store.hash == "" || strings.Contains(store.hash, created.APIKey) {
		t.Fatalf("hash = %q", store.hash)
	}
	if created.ExpiresAt.Sub(service.now()) != 90*24*time.Hour {
		t.Fatalf("expires_at = %s", created.ExpiresAt)
	}
}

func TestCreateValidatesInput(t *testing.T) {
	service := NewService(&storeStub{})
	tests := []CreateInput{
		{Name: ""},
		{Name: "agent", Scopes: []string{"unknown"}},
		{Name: "agent", ExpiresInDays: 366},
	}
	for _, input := range tests {
		if _, err := service.Create(t.Context(), "user-1", input); err == nil {
			t.Fatalf("input %#v should fail", input)
		}
	}
}

func TestAuthenticateHidesStoreErrors(t *testing.T) {
	store := &storeStub{userID: "user-1"}
	service := NewService(store)
	userID, err := service.Authenticate(t.Context(), "mk_secret", ScopeModelInvoke)
	if err != nil || userID != "user-1" {
		t.Fatalf("authenticate = %q, %v", userID, err)
	}
	store.authErr = errors.New("database unavailable")
	if _, err := service.Authenticate(t.Context(), "mk_secret", ScopeModelInvoke); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("error = %v", err)
	}
}
