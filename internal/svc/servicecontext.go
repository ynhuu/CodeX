package svc

import (
	"context"
	"net/http"
	"time"
	"web/internal/config"
	"web/internal/middleware"

	"go.uber.org/zap"
)

type ServiceContext struct {
	Ctx        context.Context
	cancel     func()
	Config     config.Config
	Session    *Session
	TokenUsage *TokenUsage
	APIKey     middleware.APIKey
	HttpClient *http.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx, cancel := context.WithCancel(context.Background())

	svcCtx := &ServiceContext{
		Ctx:        ctx,
		cancel:     cancel,
		Config:     c,
		Session:    NewSession(ctx, c.SessionDir),
		TokenUsage: NewTokenUsage(c.TokenUsageDir),
		APIKey:     middleware.NewAPIKey(c.SK),
		HttpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}

	svcCtx.Session.LoadSession()
	go svcCtx.Session.Refresher()

	return svcCtx
}

func (s *ServiceContext) Close() {
	s.cancel()
	if err := s.TokenUsage.Flush(); err != nil {
		zap.S().Warnw("flush token usage failed", "err", err)
	}
	_ = zap.L().Sync()
}
