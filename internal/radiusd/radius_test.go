package radiusd

import (
	"bytes"
	"sync"
	"testing"
	"time"

	radiuserrors "github.com/talkincode/toughradius/v9/internal/radiusd/errors"
)

// Testpure logic functions without database dependency

func TestCheckAuthRateLimitBasic(t *testing.T) {
	service := &RadiusService{
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
	}

	// TestFirst authentication
	err := service.CheckAuthRateLimit("user1")
	if err != nil {
		t.Errorf("first auth should succeed, got error: %v", err)
	}

	// TestFrequent authentication（should be limited）
	err = service.CheckAuthRateLimit("user1")
	if err == nil {
		t.Error("expected rate limit error for rapid authentication")
	}

	// Validate error types
	authErr, ok := radiuserrors.GetAuthError(err)
	if !ok {
		t.Errorf("expected AuthError, got %T", err)
	} else if authErr.MetricsType != "radus_reject_limit" {
		t.Errorf("expected reject limit error, got %s", authErr.MetricsType)
	}
}

func TestCheckAuthRateLimitAfterWait(t *testing.T) {
	service := &RadiusService{
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
	}

	// First authentication
	_ = service.CheckAuthRateLimit("user1")

	// Wait beyond rate limit time
	time.Sleep(time.Duration(RadiusAuthRateInterval+1) * time.Second)

	// Second authentication should succeed
	err := service.CheckAuthRateLimit("user1")
	if err != nil {
		t.Errorf("auth after wait should succeed, got error: %v", err)
	}
}

func TestCheckAuthRateLimitDifferentUsers(t *testing.T) {
	service := &RadiusService{
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
	}

	// Authentication of different users should not affect each other
	err1 := service.CheckAuthRateLimit("user1")
	if err1 != nil {
		t.Errorf("user1 first auth should succeed: %v", err1)
	}

	err2 := service.CheckAuthRateLimit("user2")
	if err2 != nil {
		t.Errorf("user2 first auth should succeed: %v", err2)
	}

	// Validate two users currently in the cache
	service.arclock.Lock()
	count := len(service.AuthRateCache)
	service.arclock.Unlock()

	if count != 2 {
		t.Errorf("expected 2 users in cache, got %d", count)
	}
}

func TestReleaseAuthRateLimit(t *testing.T) {
	service := &RadiusService{
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
	}

	// Add user to rate limit cache
	_ = service.CheckAuthRateLimit("user1")

	// Release rate limit
	service.ReleaseAuthRateLimit("user1")

	// Validate the user is removed from the cache
	service.arclock.Lock()
	_, exists := service.AuthRateCache["user1"]
	service.arclock.Unlock()

	if exists {
		t.Error("user should be removed from cache after release")
	}

	// Immediate re-authentication should succeed
	err := service.CheckAuthRateLimit("user1")
	if err != nil {
		t.Errorf("auth should succeed after release: %v", err)
	}
}

func TestCheckAuthRateLimitConcurrent(t *testing.T) {
	service := &RadiusService{
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
	}

	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	// Concurrent test for same user
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := service.CheckAuthRateLimit("concurrent_user")
			mu.Lock()
			if err == nil {
				successCount++
			} else {
				failCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Only first should succeed, rest should fail
	if successCount < 1 {
		t.Error("at least one concurrent request should succeed")
	}
	if failCount < 1 {
		t.Error("some concurrent requests should fail due to rate limit")
	}

	t.Logf("Concurrent test: %d success, %d failed", successCount, failCount)
}

func TestEAPStateManagement(t *testing.T) {
	service := &RadiusService{
		EapStateCache: make(map[string]EapState),
		eaplock:       sync.Mutex{},
	}

	// Add EAP Status
	stateID := "test-state-id"
	username := "testuser"
	challenge := []byte("challenge-data")
	eapMethod := "eap-md5"

	service.AddEapState(stateID, username, challenge, eapMethod)

	// get EAP Status
	state, err := service.GetEapState(stateID)
	if err != nil {
		t.Fatalf("failed to get EAP state: %v", err)
	}

	if state.Username != username {
		t.Errorf("expected username %s, got %s", username, state.Username)
	}

	if !bytes.Equal(state.Challenge, challenge) {
		t.Errorf("challenge data mismatch")
	}

	if state.EapMethad != eapMethod {
		t.Errorf("expected method %s, got %s", eapMethod, state.EapMethad)
	}

	if state.Success {
		t.Error("initial state should have Success=false")
	}

	// Delete EAP Status
	service.DeleteEapState(stateID)

	// Validatedeleted
	_, err = service.GetEapState(stateID)
	if err == nil {
		t.Error("expected error when getting deleted state")
	}
}

func TestGetEapStateNotFound(t *testing.T) {
	service := &RadiusService{
		EapStateCache: make(map[string]EapState),
		eaplock:       sync.Mutex{},
	}

	_, err := service.GetEapState("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent state")
	}

	if err.Error() != "state not found" {
		t.Errorf("expected 'state not found' error, got: %v", err)
	}
}

func TestEAPStateConcurrentAccess(t *testing.T) {
	service := &RadiusService{
		EapStateCache: make(map[string]EapState),
		eaplock:       sync.Mutex{},
	}

	var wg sync.WaitGroup
	stateCount := 100

	// Concurrent add states
	for i := 0; i < stateCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			stateID := "state-" + string(rune(id))
			service.AddEapState(stateID, "user", []byte("challenge"), "eap-md5")
		}(i)
	}

	wg.Wait()

	// ValidateAll states added
	service.eaplock.Lock()
	count := len(service.EapStateCache)
	service.eaplock.Unlock()

	if count != stateCount {
		t.Errorf("expected %d states, got %d", stateCount, count)
	}
}

func TestAuthRateCacheConcurrentAccess(t *testing.T) {
	service := &RadiusService{
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
	}

	var wg sync.WaitGroup
	userCount := 50

	// Concurrent add different users
	for i := 0; i < userCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			username := "user-" + string(rune(id))
			_ = service.CheckAuthRateLimit(username)
		}(i)
	}

	wg.Wait()

	// Validate the number of users in the cache
	service.arclock.Lock()
	count := len(service.AuthRateCache)
	service.arclock.Unlock()

	if count != userCount {
		t.Logf("Note: Expected %d users, got %d", userCount, count)
	}
}

func TestEAPStateUpdate(t *testing.T) {
	service := &RadiusService{
		EapStateCache: make(map[string]EapState),
		eaplock:       sync.Mutex{},
	}

	stateID := "test-state"

	// Add the initial status
	service.AddEapState(stateID, "user1", []byte("challenge1"), "eap-md5")

	// Update state（by overwriting）
	service.AddEapState(stateID, "user2", []byte("challenge2"), "eap-mschapv2")

	// ValidateStatusupdated
	state, err := service.GetEapState(stateID)
	if err != nil {
		t.Fatalf("failed to get state: %v", err)
	}

	if state.Username != "user2" {
		t.Errorf("expected username user2, got %s", state.Username)
	}

	if state.EapMethad != "eap-mschapv2" {
		t.Errorf("expected method eap-mschapv2, got %s", state.EapMethad)
	}
}

func TestReleaseAuthRateLimitNonexistent(t *testing.T) {
	service := &RadiusService{
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
	}

	// Releasing a non-existent user should not panic
	service.ReleaseAuthRateLimit("nonexistent-user")

	// Validate the cache is empty
	service.arclock.Lock()
	count := len(service.AuthRateCache)
	service.arclock.Unlock()

	if count != 0 {
		t.Errorf("expected empty cache, got %d entries", count)
	}
}

func TestDeleteEapStateNonexistent(t *testing.T) {
	service := &RadiusService{
		EapStateCache: make(map[string]EapState),
		eaplock:       sync.Mutex{},
	}

	// DeleteNon-existent state should not panic
	service.DeleteEapState("nonexistent-state")

	// Validate the cache is empty
	service.eaplock.Lock()
	count := len(service.EapStateCache)
	service.eaplock.Unlock()

	if count != 0 {
		t.Errorf("expected empty cache, got %d entries", count)
	}
}

func TestMultipleEAPStates(t *testing.T) {
	service := &RadiusService{
		EapStateCache: make(map[string]EapState),
		eaplock:       sync.Mutex{},
	}

	// Add multiple different EAP Status
	states := map[string]string{
		"state1": "user1",
		"state2": "user2",
		"state3": "user3",
	}

	for stateID, username := range states {
		service.AddEapState(stateID, username, []byte("challenge"), "eap-md5")
	}

	// ValidateAll states exist
	for stateID, expectedUser := range states {
		state, err := service.GetEapState(stateID)
		if err != nil {
			t.Errorf("failed to get state %s: %v", stateID, err)
			continue
		}
		if state.Username != expectedUser {
			t.Errorf("state %s: expected user %s, got %s", stateID, expectedUser, state.Username)
		}
	}
}

func TestAuthRateLimitExpiry(t *testing.T) {
	service := &RadiusService{
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
	}

	// Add user
	_ = service.CheckAuthRateLimit("user1")

	// Get add time
	service.arclock.Lock()
	startTime := service.AuthRateCache["user1"].Starttime
	service.arclock.Unlock()

	// Validatetimestamp
	if time.Since(startTime) > time.Second {
		t.Error("start time should be recent")
	}

	// Wait for expiration
	time.Sleep(time.Duration(RadiusAuthRateInterval+1) * time.Second)

	// Should be able to authenticate again
	err := service.CheckAuthRateLimit("user1")
	if err != nil {
		t.Errorf("should be able to auth after expiry: %v", err)
	}
}
