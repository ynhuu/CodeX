package logic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"web/internal/svc"
	"web/internal/types"

	"go.uber.org/zap"
)

type CodexAuth struct {
	svcCtx *svc.ServiceContext
}

func NewCodexAuthLogic(svcCtx *svc.ServiceContext) *CodexAuth {
	return &CodexAuth{
		svcCtx: svcCtx,
	}
}

func (l *CodexAuth) GetUserCode() (string, string, error) {
	payload := strings.NewReader(`{"client_id": "app_EMoamEEZ73f0CkXaXp7hrann"}`)
	resp, err := l.svcCtx.HttpClient.Post("https://auth.openai.com/api/accounts/deviceauth/usercode", "application/json", payload)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("get user code status=%d body=%s", resp.StatusCode, string(body))
	}

	var userCode struct {
		DeviceAuthId string `json:"device_auth_id"`
		Code         string `json:"user_code"`
	}

	if err := json.Unmarshal(body, &userCode); err != nil {
		return "", "", fmt.Errorf("decode user code failed: %w, body=%s", err, string(body))
	}

	return userCode.DeviceAuthId, userCode.Code, nil
}

func (l *CodexAuth) GetDeviceToken(ctx context.Context, deviceAuthId, code string) {
	payload := fmt.Sprintf(`{"device_auth_id": "%s", "user_code": "%s"}`, deviceAuthId, code)
	ctxTimeOut, cancel := context.WithTimeout(l.svcCtx.Ctx, 5*time.Minute)
	defer cancel()

	tickLoop := time.NewTicker(5 * time.Second)
	defer tickLoop.Stop()

	defer zap.S().Infow("Exit GetDeviceToken.", "code", code)

	for {
		select {
		case <-ctx.Done():
			return
		case <-l.svcCtx.Ctx.Done():
			return
		case <-ctxTimeOut.Done():
			return
		case <-tickLoop.C:
			resp, err := l.svcCtx.HttpClient.Post("https://auth.openai.com/api/accounts/deviceauth/token", "application/json", strings.NewReader(payload))
			if err != nil {
				zap.S().Errorw("GetDeviceToken", "err", err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				_ = resp.Body.Close()
				continue
			}

			var token struct {
				Status            string `json:"status"`
				AuthorizationCode string `json:"authorization_code"`
				CodeVerifier      string `json:"code_verifier"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&token); err != nil || token.Status != "success" {
				zap.S().Errorw("GetDeviceToken", "err", err)
				_ = resp.Body.Close()
				continue
			}
			_ = resp.Body.Close()

			l.getOAthToken(token.AuthorizationCode, token.CodeVerifier)
			return
		}
	}
}

func (l *CodexAuth) getOAthToken(authorizationCode, codeVerifier string) {
	payload := strings.NewReader(url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {authorizationCode},
		"redirect_uri":  {"https://auth.openai.com/deviceauth/callback"},
		"client_id":     {"app_EMoamEEZ73f0CkXaXp7hrann"},
		"code_verifier": {codeVerifier},
	}.Encode())
	resp, err := l.svcCtx.HttpClient.Post("https://auth.openai.com/oauth/token", "application/x-www-form-urlencoded", payload)
	if err != nil {
		zap.S().Errorw("getOAthToken", "err", err)
		return
	}
	defer resp.Body.Close()

	var session types.CodexSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		zap.S().Errorw("getOAthToken", "err", err)
		return
	}

	session.Expires = time.Now().Add(time.Duration(session.Expires) * time.Second).Unix()
	parts := strings.Split(session.AccessToken, ".")
	data, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var idToken struct {
		OpenAIAuth struct {
			PlanType string `json:"chatgpt_plan_type"`
		} `json:"https://api.openai.com/auth"`
		OpenAIProfile struct {
			Email string `json:"email"`
		} `json:"https://api.openai.com/profile"`
	}

	_ = json.Unmarshal(data, &idToken)
	session.AccountID = idToken.OpenAIProfile.Email
	session.PlanType = idToken.OpenAIAuth.PlanType

	if err := l.svcCtx.Session.Insert(session); err != nil {
		zap.S().Errorw("getOAthToken err", err.Error())
		return
	}

	zap.S().Infow("getOAthToken success", "AccountID", session.AccountID)
}

func (l *CodexAuth) GetResetAt(session *types.CodexSession) (int64, error) {
	req, err := http.NewRequest(http.MethodGet, "https://chatgpt.com/backend-api/wham/usage", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	resp, err := l.svcCtx.HttpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("usage response status: %d", resp.StatusCode)
	}

	var usage types.CodexUsage
	if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
		return 0, err
	}

	zap.S().Infow("GetUsage success", "Usage", usage)
	return usage.RateLimit.PrimaryWindow.ResetAt, nil
}
