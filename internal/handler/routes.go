package handler

import (
	"net/http"
	"web/internal/svc"

	core "github.com/ynhuu/gin-core"
)

func RegisterHandlers(app *core.Server, svcCtx *svc.ServiceContext) {
	app.GET("/codex", Home())

	codex := app.Group("/codex", svcCtx.APIKey.Auth())
	codex.POST("/responses", CodexResponses(svcCtx))
	codex.GET("/usage", CodexUsage(svcCtx))
	codex.Match([]string{http.MethodGet, http.MethodPost}, "/cli", CodexDeviceAuth(svcCtx))
}
