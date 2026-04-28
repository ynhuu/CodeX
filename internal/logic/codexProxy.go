package logic

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"web/internal/middleware"
	"web/internal/svc"

	"go.uber.org/zap"
)

type CodexProxyLogic struct {
	svcCtx *svc.ServiceContext
}

func NewCodexProxyLogic(svcCtx *svc.ServiceContext) *CodexProxyLogic {
	return &CodexProxyLogic{
		svcCtx: svcCtx,
	}
}

func (l *CodexProxyLogic) ReverseProxy(rw http.ResponseWriter, req *http.Request, planType string) error {
	session, err := l.svcCtx.Session.Next(planType)
	if err != nil {
		return err
	}

	accept := req.Header.Get("Accept")
	acceptE := req.Header.Get("Accept-Encoding")
	ua := req.Header.Get("User-Agent")
	cType := req.Header.Get("Content-Type")
	sid := req.Header.Get("X-Session-Affinity")
	sk := middleware.CurrentSK(req)

	req.Header = make(http.Header)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", accept)
	req.Header.Set("Session-Id", sid)
	req.Header.Set("Content-Type", cType)
	req.Header.Set("X-Session-Affinity", sid)
	req.Header.Set("Accept-Encoding", acceptE)
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)

	target, _ := url.Parse("https://chatgpt.com/backend-api/codex/responses")
	proxy := httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			req.Out.URL.Scheme = target.Scheme
			req.Out.URL.Host = target.Host
			req.Out.URL.Path = target.Path
			req.Out.Host = target.Host
		},
		ModifyResponse: func(resp *http.Response) error {
			if resp.StatusCode == http.StatusTooManyRequests {
				resetAt, err := NewCodexAuthLogic(l.svcCtx).GetResetAt(session)
				if err != nil {
					zap.S().Errorw("GetResetAt", "accountId", session.AccountID, "err", err)
					return nil
				}

				session.ResetAt = resetAt
				if err := l.svcCtx.Session.Insert(*session); err != nil {
					zap.S().Errorw("Session update failed", "accountId", session.AccountID, "err", err)
				}
			}

			NewTokenUsageLogic(l.svcCtx).ObserveResponseCompleted(resp, session, sk)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(rw, req)

	return nil
}
