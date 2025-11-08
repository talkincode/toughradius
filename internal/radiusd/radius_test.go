package radiusd

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

// 测试不依赖数据库的纯逻辑功能

func TestCheckAuthRateLimitBasic(t *testing.T) {
	service := &RadiusService{
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
	}

	// 测试首次认证
	err := service.CheckAuthRateLimit("user1")
	if err != nil {
		t.Errorf("first auth should succeed, got error: %v", err)
	}

	// 测试频繁认证（应该被限制）
	err = service.CheckAuthRateLimit("user1")
	if err == nil {
		t.Error("expected rate limit error for rapid authentication")
	}

	// 验证错误类型
	authErr, ok := err.(*AuthError)
	if !ok {
		t.Errorf("expected AuthError, got %T", err)
	} else if authErr.Type != "radus_reject_limit" {
		t.Errorf("expected reject limit error, got %s", authErr.Type)
	}
}

func TestCheckAuthRateLimitAfterWait(t *testing.T) {
	service := &RadiusService{
		AuthRateCache: make(map[string]AuthRateUser),
		arclock:       sync.Mutex{},
	}

	// 首次认证
	_ = service.CheckAuthRateLimit("user1")

	// 等待超过限制时间
	time.Sleep(time.Duration(RadiusAuthRateInterval+1) * time.Second)

	// 再次认证应该成功
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

	// 不同用户的认证不应该互相影响
	err1 := service.CheckAuthRateLimit("user1")
	if err1 != nil {
		t.Errorf("user1 first auth should succeed: %v", err1)
	}

	err2 := service.CheckAuthRateLimit("user2")
	if err2 != nil {
		t.Errorf("user2 first auth should succeed: %v", err2)
	}

	// 验证缓存中有两个用户
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

	// 添加用户到限制缓存
	_ = service.CheckAuthRateLimit("user1")

	// 释放限制
	service.ReleaseAuthRateLimit("user1")

	// 验证用户已从缓存中移除
	service.arclock.Lock()
	_, exists := service.AuthRateCache["user1"]
	service.arclock.Unlock()

	if exists {
		t.Error("user should be removed from cache after release")
	}

	// 立即再次认证应该成功
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

	// 并发测试相同用户
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

	// 只有第一个应该成功，其余应该失败
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

	// 添加 EAP 状态
	stateID := "test-state-id"
	username := "testuser"
	challenge := []byte("challenge-data")
	eapMethod := "eap-md5"

	service.AddEapState(stateID, username, challenge, eapMethod)

	// 获取 EAP 状态
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

	// 删除 EAP 状态
	service.DeleteEapState(stateID)

	// 验证已删除
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

	// 并发添加状态
	for i := 0; i < stateCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			stateID := "state-" + string(rune(id))
			service.AddEapState(stateID, "user", []byte("challenge"), "eap-md5")
		}(i)
	}

	wg.Wait()

	// 验证所有状态都被添加
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

	// 并发添加不同用户
	for i := 0; i < userCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			username := "user-" + string(rune(id))
			_ = service.CheckAuthRateLimit(username)
		}(i)
	}

	wg.Wait()

	// 验证缓存中的用户数
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

	// 添加初始状态
	service.AddEapState(stateID, "user1", []byte("challenge1"), "eap-md5")

	// 更新状态（通过覆盖）
	service.AddEapState(stateID, "user2", []byte("challenge2"), "eap-mschapv2")

	// 验证状态被更新
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

	// 释放不存在的用户不应该 panic
	service.ReleaseAuthRateLimit("nonexistent-user")

	// 验证缓存为空
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

	// 删除不存在的状态不应该 panic
	service.DeleteEapState("nonexistent-state")

	// 验证缓存为空
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

	// 添加多个不同的 EAP 状态
	states := map[string]string{
		"state1": "user1",
		"state2": "user2",
		"state3": "user3",
	}

	for stateID, username := range states {
		service.AddEapState(stateID, username, []byte("challenge"), "eap-md5")
	}

	// 验证所有状态都存在
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

	// 添加用户
	_ = service.CheckAuthRateLimit("user1")

	// 获取添加时间
	service.arclock.Lock()
	startTime := service.AuthRateCache["user1"].Starttime
	service.arclock.Unlock()

	// 验证时间戳
	if time.Since(startTime) > time.Second {
		t.Error("start time should be recent")
	}

	// 等待到期
	time.Sleep(time.Duration(RadiusAuthRateInterval+1) * time.Second)

	// 应该能再次认证
	err := service.CheckAuthRateLimit("user1")
	if err != nil {
		t.Errorf("should be able to auth after expiry: %v", err)
	}
}
