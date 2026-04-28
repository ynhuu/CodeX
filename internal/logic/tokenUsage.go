package logic

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"web/internal/svc"
	"web/internal/types"

	"go.uber.org/zap"
)

type TokenUsageLogic struct {
	svcCtx *svc.ServiceContext
}

func NewTokenUsageLogic(svcCtx *svc.ServiceContext) *TokenUsageLogic {
	return &TokenUsageLogic{
		svcCtx: svcCtx,
	}
}

func (l *TokenUsageLogic) ObserveResponseCompleted(resp *http.Response, session *types.CodexSession, sk string) {
	if resp.Body == nil {
		return
	}

	originBody := resp.Body
	pr, pw := io.Pipe()
	resp.Body = pr

	go l.streamAndObserveResponseCompleted(originBody, pw, session.AccountID, sk)
}

func (l *TokenUsageLogic) streamAndObserveResponseCompleted(originBody io.ReadCloser, pw *io.PipeWriter, accountID, sk string) {
	defer originBody.Close()

	scanner := bufio.NewScanner(originBody)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var eventName string
	var dataLines []string

	for scanner.Scan() {
		line := scanner.Text()
		if _, err := io.WriteString(pw, line+"\n"); err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		switch {
		case strings.HasPrefix(line, "event:"):
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			dataLines = dataLines[:0]
		case strings.HasPrefix(line, "data:"):
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		case line == "":
			if eventName == "response.completed" {
				l.recordCompletedTokenUsage(accountID, sk, []byte(strings.Join(dataLines, "\n")))
			}
			eventName = ""
			dataLines = dataLines[:0]
		}
	}

	if err := scanner.Err(); err != nil {
		_ = pw.CloseWithError(err)
		return
	}
	_ = pw.Close()
}

func (l *TokenUsageLogic) recordCompletedTokenUsage(accountID, sk string, data []byte) {
	if len(data) == 0 {
		return
	}

	var event types.TokenUsageCompletedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		zap.S().Warnw("parse response.completed failed", "accountId", accountID, "err", err)
		return
	}

	usage := event.Response.Usage
	if usage.TotalTokens == 0 && usage.InputTokens == 0 && usage.OutputTokens == 0 {
		return
	}

	if err := l.svcCtx.TokenUsage.Add(sk, usage); err != nil {
		zap.S().Warnw("save token usage failed", "accountId", accountID, "err", err)
	}
}
