package svc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"web/internal/types"
)

type TokenUsage struct {
	mu   sync.RWMutex
	dir  string
	data map[string]types.TokenUsage
}

func NewTokenUsage(dir string) *TokenUsage {
	t := &TokenUsage{
		dir:  dir,
		data: make(map[string]types.TokenUsage),
	}
	t.loadAll()
	return t
}

func (t *TokenUsage) Add(sk string, usage types.TokenUsageItem) error {
	if sk == "" {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	record := normalizeTokenUsageDate(t.data[sk], now)

	record.Today.InputTokens += usage.InputTokens
	record.Today.OutputTokens += usage.OutputTokens
	record.Today.TotalTokens += usage.TotalTokens
	record.Total.InputTokens += usage.InputTokens
	record.Total.OutputTokens += usage.OutputTokens
	record.Total.TotalTokens += usage.TotalTokens
	record.UpdatedAt = now.Unix()
	t.data[sk] = record

	return nil
}

func (t *TokenUsage) Usage(sk string) (types.TokenUsage, error) {
	if sk == "" {
		return types.TokenUsage{}, nil
	}

	t.mu.RLock()
	usage, ok := t.data[sk]
	if !ok || usage.Date == time.Now().Format(time.DateOnly) {
		t.mu.RUnlock()
		return usage, nil
	}
	t.mu.RUnlock()

	t.mu.Lock()
	defer t.mu.Unlock()

	usage = normalizeTokenUsageDate(t.data[sk], time.Now())
	t.data[sk] = usage
	return usage, nil
}

func normalizeTokenUsageDate(usage types.TokenUsage, now time.Time) types.TokenUsage {
	today := now.Format(time.DateOnly)
	if usage.Date == today {
		return usage
	}

	yesterday := now.AddDate(0, 0, -1).Format(time.DateOnly)
	if usage.Date == yesterday {
		usage.Yesterday = usage.Today
	} else {
		usage.Yesterday = types.TokenUsageItem{}
	}
	usage.Today = types.TokenUsageItem{}
	usage.Date = today
	return usage
}

func (t *TokenUsage) Flush() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for sk, usage := range t.data {
		if err := t.save(sk, usage); err != nil {
			return err
		}
	}
	return nil
}

func (t *TokenUsage) load(sk string) (types.TokenUsage, error) {
	data, err := os.ReadFile(t.path(sk))
	if err != nil {
		return types.TokenUsage{}, err
	}

	var usage types.TokenUsage
	if err := json.Unmarshal(data, &usage); err != nil {
		return types.TokenUsage{}, err
	}
	if usage.Date == "" && usage.Total.TotalTokens == 0 && usage.Total.InputTokens == 0 && usage.Total.OutputTokens == 0 {
		var legacy types.TokenUsageItem
		if err := json.Unmarshal(data, &legacy); err == nil {
			usage.Date = time.Now().Format(time.DateOnly)
			usage.Today = legacy
			usage.Total = legacy
		}
	}
	return usage, nil
}

func (t *TokenUsage) loadAll() {
	entries, err := os.ReadDir(t.dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		sk := strings.TrimSuffix(entry.Name(), ".json")
		usage, err := t.load(sk)
		if err != nil {
			continue
		}
		t.data[sk] = usage
	}
}

func (t *TokenUsage) save(sk string, usage types.TokenUsage) error {
	if err := os.MkdirAll(t.dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(usage, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(t.path(sk), data, 0o644)
}

func (t *TokenUsage) path(sk string) string {
	return filepath.Join(t.dir, sk+".json")
}
