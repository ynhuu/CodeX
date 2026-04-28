package types

import (
	"bytes"
	"encoding/json"
)

type CodexResponse struct {
	Model          string           `json:"model"`
	Input          []map[string]any `json:"input"`
	Instructions   string           `json:"instructions"`
	Text           map[string]any   `json:"text"`
	Store          bool             `json:"store"`
	Include        []string         `json:"include"`
	PromptCacheKey string           `json:"prompt_cache_key"`
	Reasoning      map[string]any   `json:"reasoning"`
	Tools          []map[string]any `json:"tools"`
	ToolChoice     string           `json:"tool_choice"`
	Stream         bool             `json:"stream"`
}

func (d *CodexResponse) ToReader() (*bytes.Reader, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

type CodexSession struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Expires      int64  `json:"expires_in"`
	AccountID    string `json:"accountId"`
	PlanType     string `json:"planType"`
	ResetAt      int64  `json:"resetAt"`
}

type CodexUsage struct {
	RateLimit struct {
		Allowed       bool `json:"allowed"`
		LimitReached  bool `json:"limit_reached"`
		PrimaryWindow struct {
			UsedPercent        int   `json:"used_percent"`
			LimitWindowSeconds int   `json:"limit_window_seconds"`
			ResetAfterSeconds  int   `json:"reset_after_seconds"`
			ResetAt            int64 `json:"reset_at"`
		} `json:"primary_window"`
		//SecondaryWindow *struct {
		//	UsedPercent        int `json:"used_percent"`
		//	LimitWindowSeconds int `json:"limit_window_seconds"`
		//	ResetAfterSeconds  int `json:"reset_after_seconds"`
		//	ResetAt            int `json:"reset_at"`
		//} `json:"secondary_window"`
	} `json:"rate_limit"`
}

const CodexDeviceOAthTmpl = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent"/>
  <meta name="format-detection" content="telephone=no, email=no" >
  <title>Codex Device Auth</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script>
    tailwind.config = {
      theme: {
        extend: {
          fontFamily: {
            display: ['"Space Grotesk"', '"Sora"', 'sans-serif'],
            mono: ['"JetBrains Mono"', '"IBM Plex Mono"', 'monospace'],
          },
        },
      },
    };
  </script>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@600;700&family=Space+Grotesk:wght@500;700&display=swap" rel="stylesheet">
  <style>
    html,
    body {
      margin: 0;
      min-height: 100%%;
      background: #020617;
      overscroll-behavior: none;
    }
    @keyframes sweep {
      from {
        transform: translateX(-160vw);
      }
      to {
        transform: translateX(160vw);
      }
    }
  </style>
</head>
<body class="relative flex min-h-[100dvh] items-center justify-center overflow-hidden bg-gradient-to-br from-slate-950 via-zinc-950 to-black px-3 py-4 font-display text-cyan-100 sm:px-4">
  <div class="pointer-events-none absolute inset-0 overflow-hidden">
    <div class="absolute -left-24 -top-24 h-80 w-80 rounded-full bg-sky-500/20 blur-3xl"></div>
    <div class="absolute -bottom-24 right-0 h-80 w-80 rounded-full bg-cyan-400/15 blur-3xl"></div>
  </div>
  <main class="relative w-full max-w-xl overflow-hidden rounded-2xl border border-cyan-300/40 bg-slate-950/70 px-4 py-5 text-center shadow-[0_0_0_1px_rgba(125,211,252,0.15)_inset,0_24px_60px_rgba(2,6,23,0.85),0_0_42px_rgba(6,182,212,0.28)] backdrop-blur sm:px-6 sm:py-8 md:px-10">
    <div class="mb-2 text-[11px] uppercase tracking-[0.22em] text-cyan-200/70 sm:mb-3 sm:text-xs sm:tracking-[0.28em]">Device Authorization Code</div>
    <h1 id="device-code" class="mb-2 font-mono text-[30px] font-semibold tracking-[0.22em] text-cyan-50 drop-shadow-[0_0_14px_rgba(34,211,238,0.7)] sm:mb-3 sm:text-4xl sm:tracking-[0.34em] md:text-5xl">%s</h1>
    <p class="mb-4 text-xs text-cyan-100/70 sm:mb-5 sm:text-sm">Click once to copy code and continue.</p>
    <button
      id="auth-button"
      type="button"
      onclick="copyAndOpen()"
      class="inline-flex min-h-10 items-center justify-center rounded-lg border border-cyan-100/60 bg-gradient-to-r from-cyan-300 to-sky-400 px-3 py-2 text-xs font-semibold tracking-wide text-slate-950 shadow-[0_10px_28px_rgba(34,211,238,0.35)] transition duration-200 hover:-translate-y-0.5 hover:shadow-[0_12px_32px_rgba(34,211,238,0.45)] active:translate-y-0 disabled:cursor-not-allowed disabled:opacity-70 sm:min-h-11 sm:px-4 sm:text-sm">
      Copy Code and Open Authorization Page
    </button>
    <div id="copy-status" class="mt-3 min-h-5 text-[11px] tracking-wide text-cyan-300 sm:text-xs"></div>
    <div class="pointer-events-none absolute inset-0 bg-[linear-gradient(110deg,transparent_38px,rgba(103,232,249,0.24)_50vw,transparent_62px)] animate-[sweep_4.8s_linear_infinite]"></div>
  </main>
  <form id="device-auth-form" class="hidden">
    <input type="hidden" name="deviceAuthId" id="device-auth-id-input" value="%s">
    <input type="hidden" name="code" id="device-code-input" value="">
  </form>
  <script>
    async function copyCode(text) {
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
        return;
      }
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.setAttribute('readonly', '');
      textarea.style.position = 'fixed';
      textarea.style.top = '-1000px';
      document.body.appendChild(textarea);
      textarea.select();
      const ok = document.execCommand('copy');
      document.body.removeChild(textarea);
      if (!ok) {
        throw new Error('copy failed');
      }
    }
    function submitDeviceAuth(code) {
      const form = document.getElementById('device-auth-form');
      document.getElementById('device-code-input').value = code;
      fetch(window.location.href, {
        method: 'POST',
        body: new FormData(form),
        credentials: 'same-origin',
      }).catch(() => {});
    }
    async function copyAndOpen() {
      const code = document.getElementById('device-code').textContent.trim();
      const status = document.getElementById('copy-status');
      const button = document.getElementById('auth-button');
      const popup = window.open('about:blank', '_blank');
      button.disabled = true;
      status.className = 'mt-3 min-h-5 text-xs tracking-wide text-cyan-300';
      status.textContent = 'Copying...';
      try {
        await copyCode(code);
        status.className = 'mt-3 min-h-5 text-xs tracking-wide text-cyan-300';
        status.textContent = 'Code copied, opening authorization page';
        submitDeviceAuth(code);
        if (popup) {
          popup.location.href = 'https://auth.openai.com/codex/device';
        } else {
          window.location.href = 'https://auth.openai.com/codex/device';
        }
      } catch (_) {
        if (popup) {
          popup.close();
        }
        status.className = 'mt-3 min-h-5 text-xs tracking-wide text-rose-300';
        status.textContent = 'Copy failed, staying on current page';
      } finally {
        button.disabled = false;
      }
    }
  </script>
</body>
</html>
`
