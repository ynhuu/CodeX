package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"web/internal/logic"
	"web/internal/middleware"
	"web/internal/svc"
	"web/internal/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CodexResponses(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var openai types.OpenAIResponse
		if err := c.ShouldBind(&openai); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		sid := c.GetHeader("X-Session-Affinity")
		codex := openai.ToCodex(sid)
		body, err := codex.ToReader()
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to serialize request")
			return
		}

		c.Request.Body = io.NopCloser(body)
		c.Request.ContentLength = body.Size()

		planType := svcCtx.Config.PlanType(codex.Model)
		if err := logic.NewCodexProxyLogic(svcCtx).ReverseProxy(c.Writer, c.Request, planType); err != nil {
			c.String(http.StatusServiceUnavailable, err.Error())
		}
	}
}

func CodexDeviceAuth(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet:
			deviceAuthId, code, err := logic.NewCodexAuthLogic(svcCtx).GetUserCode()
			if err != nil {
				zap.S().Error("CodexDeviceAuth", zap.Error(err))
				c.Status(http.StatusInternalServerError)
				return
			}

			c.Data(http.StatusOK, "text/html; charset=utf-8", fmt.Appendf(nil, types.CodexDeviceOAthTmpl, code, deviceAuthId))
		case http.MethodPost:
			deviceAuthID := c.PostForm("deviceAuthId")
			code := c.PostForm("code")
			if deviceAuthID != "" && code != "" {
				logic.NewCodexAuthLogic(svcCtx).GetDeviceToken(c.Request.Context(), deviceAuthID, code)
				return
			}
			c.Status(http.StatusBadRequest)
		}
	}
}

func CodexUsage(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		usage, err := svcCtx.TokenUsage.Usage(middleware.CurrentSK(c.Request))
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		if strings.Contains(c.GetHeader("Accept"), "text/html") {
			updatedAt := "never"
			if usage.UpdatedAt > 0 {
				updatedAt = time.Unix(usage.UpdatedAt, 0).Format(time.RFC3339)
			}
			usageJSON, err := json.Marshal(usage)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", fmt.Appendf(nil, types.TokenUsageTmpl,
				updatedAt,
				usageJSON,
			))
			return
		}

		c.JSON(http.StatusOK, usage)
	}
}
