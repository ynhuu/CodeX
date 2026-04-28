package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"web/internal/types"

	"go.uber.org/zap"
)

type Session struct {
	ctx      context.Context
	path     string
	mu       sync.RWMutex
	idx      int
	planIdx  map[string]int
	sessions []types.CodexSession
}

func NewSession(ctx context.Context, path string) *Session {
	return &Session{ctx: ctx, path: path, planIdx: make(map[string]int)}
}

func (s *Session) LoadSession() {
	if s.path == "" {
		zap.S().Warn("SessionDir not configured, skip loading sessions")
		return
	}

	entries, err := os.ReadDir(s.path)
	if err != nil {
		zap.S().Errorw("failed to read SessionDir", "dir", s.path, "err", err)
		return
	}

	var sessions []types.CodexSession
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(s.path, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			zap.S().Warnw("failed to read session file", "file", path, "err", err)
			continue
		}

		var codex types.CodexSession
		if err := json.Unmarshal(data, &codex); err != nil {
			zap.S().Warnw("failed to parse session file", "file", path, "err", err)
			continue
		}

		sessions = append(sessions, codex)
	}
	s.mu.Lock()
	s.sessions = sessions
	s.mu.Unlock()
}

func (s *Session) Next(planType string) (*types.CodexSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := len(s.sessions)
	if n == 0 {
		return nil, fmt.Errorf("no sessions available")
	}

	now := time.Now()
	if planType == "" {
		start := s.idx % n
		s.idx++
		for i := range n {
			entry := s.sessions[(start+i)%n]
			if entry.ResetAt == 0 || now.After(time.Unix(entry.ResetAt, 0)) {
				return &entry, nil
			}
		}
		return nil, fmt.Errorf("all sessions are currently disabled")
	}

	matched := make([]types.CodexSession, 0, len(s.sessions))
	for _, entry := range s.sessions {
		if strings.Contains(entry.PlanType, planType) {
			matched = append(matched, entry)
		}
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("no sessions available for plan type %q", planType)
	}

	start := s.planIdx[planType] % len(matched)
	s.planIdx[planType]++
	for i := range len(matched) {
		entry := matched[(start+i)%len(matched)]
		if entry.ResetAt == 0 || now.After(time.Unix(entry.ResetAt, 0)) {
			return &entry, nil
		}
	}

	return nil, fmt.Errorf("all sessions are currently disabled")
}

//func (s *Session) Disable(accountID string) {
//	s.mu.Lock()
//	var updated *types.CodexSession
//
//	for i := range s.sessions {
//		if s.sessions[i].AccountID == accountID {
//			s.sessions[i].ResetAt = time.Now().Insert(5 * time.Hour).Unix()
//			snapshot := s.sessions[i]
//			updated = &snapshot
//			zap.S().Warnw("session disabled due to quota exhausted",
//				"accountId", accountID,
//				"disabledUntil", time.Unix(s.sessions[i].ResetAt, 0).Format(time.RFC3339),
//			)
//			break
//		}
//	}
//	s.mu.Unlock()
//
//	if updated != nil {
//		if err := s.SaveSession(updated); err != nil {
//			zap.S().Warnw("save disabled session failed", "accountId", updated.AccountID, "err", err)
//		}
//	}
//}

func (s *Session) Insert(codex types.CodexSession) error {
	if err := s.SaveSession(&codex); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.sessions {
		if s.sessions[i].AccountID == codex.AccountID {
			s.sessions[i] = codex
			return nil
		}
	}
	s.sessions = append(s.sessions, codex)
	return nil
}

//func (s *Session) Delete(accountID string) error {
//	path := filepath.Join(s.path, accountID+".json")
//	if err := os.Remove(filepath.Clean(path)); err != nil && !os.IsNotExist(err) {
//		return err
//	}
//
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	for i := range s.sessions {
//		if s.sessions[i].AccountID == accountID {
//			s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
//			break
//		}
//	}
//	return nil
//}

func (s *Session) SaveSession(session *types.CodexSession) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	if s.path != "" {
		if err := os.MkdirAll(s.path, 0o755); err != nil {
			return err
		}
	}
	path := filepath.Join(s.path, session.AccountID+".json")
	return os.WriteFile(path, data, 0o644)
}

func (s *Session) Refresher() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			snapshot := append([]types.CodexSession(nil), s.sessions...)
			s.mu.RUnlock()
			for _, item := range snapshot {
				expireAt := time.Unix(item.Expires, 0)
				if !time.Now().Add(time.Hour * 24).After(expireAt) {
					continue
				}
				zap.S().Infow("session expiring, refreshing", "accountId", item.AccountID)
				if err := s.refreshSession(item.AccountID); err != nil {
					zap.S().Warnw("refresh session failed", "accountId", item.AccountID, "err", err)
				}
			}
		}
	}
}

func (s *Session) refreshSession(accountID string) error {
	s.mu.RLock()
	var session types.CodexSession
	found := false
	for i := range s.sessions {
		if s.sessions[i].AccountID == accountID {
			session = s.sessions[i]
			found = true
			break
		}
	}
	s.mu.RUnlock()
	if !found {
		return fmt.Errorf("session not found: %s", accountID)
	}

	payload := strings.NewReader(fmt.Sprintf(`grant_type=refresh_token&client_id=app_EMoamEEZ73f0CkXaXp7hrann&refresh_token=%s`, session.RefreshToken))
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://auth.openai.com/oauth/token", payload)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh token response status: %d", resp.StatusCode)
	}

	var result types.CodexSession
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if result.AccessToken == "" {
		return fmt.Errorf("refresh response missing access token")
	}

	s.mu.Lock()
	var updated types.CodexSession
	found = false
	for i := range s.sessions {
		if s.sessions[i].AccountID != accountID {
			continue
		}
		s.sessions[i].AccessToken = result.AccessToken
		if result.RefreshToken != "" {
			s.sessions[i].RefreshToken = result.RefreshToken
		}
		s.sessions[i].Expires = time.Now().Add(time.Duration(result.Expires) * time.Second).Unix()
		updated = s.sessions[i]
		found = true
		break
	}
	s.mu.Unlock()
	if !found {
		return fmt.Errorf("session removed before update: %s", accountID)
	}

	zap.S().Infow("session refreshed", "accountId", updated.AccountID)

	if err := s.SaveSession(&updated); err != nil {
		zap.S().Warnw("save session failed", "accountId", updated.AccountID, "err", err)
	}

	return nil
}
