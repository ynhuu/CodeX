package types

type TokenUsageCompletedEvent struct {
	Type     string `json:"type"`
	Response struct {
		Usage TokenUsageItem `json:"usage"`
	} `json:"response"`
}

type TokenUsage struct {
	Date      string         `json:"date"`
	Today     TokenUsageItem `json:"today"`
	Yesterday TokenUsageItem `json:"yesterday"`
	Total     TokenUsageItem `json:"total"`
	UpdatedAt int64          `json:"updated_at"`
}

type TokenUsageItem struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

const TokenUsageTmpl = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent"/>
  <meta name="format-detection" content="telephone=no, email=no" >
  <title>Token Usage</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script>
    tailwind.config = {
      theme: {
        extend: {
          fontFamily: {
            display: ['"Space Grotesk"', '"Sora"', 'sans-serif'],
            mono: ['"Geist Mono"', '"JetBrains Mono"', '"IBM Plex Mono"', 'monospace'],
          },
        },
      },
    };
  </script>
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
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Geist+Mono:wght@500;600;700&family=JetBrains+Mono:wght@500;600;700&family=Space+Grotesk:wght@500;700&display=swap" rel="stylesheet">
</head>
<body class="relative flex min-h-screen items-center justify-center overflow-x-hidden bg-gradient-to-br from-slate-950 via-zinc-950 to-black px-3 pb-[calc(1rem+env(safe-area-inset-bottom))] pt-4 font-display text-cyan-100 sm:px-4 sm:py-6 overflow-hidden">
  <div class="pointer-events-none absolute inset-0 overflow-hidden">
    <div class="absolute -left-24 -top-24 h-80 w-80 rounded-full bg-sky-500/20 blur-3xl"></div>
    <div class="absolute -bottom-24 right-0 h-80 w-80 rounded-full bg-cyan-400/15 blur-3xl"></div>
  </div>
  <main class="relative mx-auto w-full max-w-2xl -translate-y-8 overflow-hidden rounded-2xl border border-cyan-300/40 bg-slate-950/70 px-3 py-4 shadow-[0_0_0_1px_rgba(125,211,252,0.15)_inset,0_24px_60px_rgba(2,6,23,0.85),0_0_42px_rgba(6,182,212,0.28)] backdrop-blur sm:translate-y-0 sm:px-5 sm:py-6 md:px-7">
    <div class="mb-3 text-[11px] uppercase tracking-[0.24em] text-cyan-200/70 sm:mb-4 sm:text-xs sm:tracking-[0.28em]">Token Usage</div>
    <div id="usage-grid" class="grid gap-2.5 sm:gap-3"></div>
    <div class="mt-3 text-[11px] text-cyan-100/65 sm:mt-4 sm:text-xs">Updated at: <span class="font-mono text-cyan-50">%s</span></div>
    <div class="pointer-events-none absolute inset-0 bg-[linear-gradient(110deg,transparent_38px,rgba(103,232,249,0.16)_50vw,transparent_62px)] animate-[sweep_5.4s_linear_infinite]"></div>
  </main>
  <script id="usage-data" type="application/json">%s</script>
  <script>
    function compactNumber(value, unit, digits) {
      return (value / unit).toFixed(digits).replace(/\.0+$/, '').replace(/(\.\d*[1-9])0+$/, '$1');
    }
    function formatTokens(value) {
      if (value >= 1000000000000) return compactNumber(value, 1000000000000, 2) + 'T';
      if (value >= 1000000000) return compactNumber(value, 1000000000, 2) + 'B';
      if (value >= 1000000) return compactNumber(value, 1000000, 2) + 'M';
      if (value >= 1000) return compactNumber(value, 1000, 1) + 'K';
      return value.toLocaleString();
    }
    function metric(label, value) {
      return '<section class="rounded-lg border border-cyan-300/35 bg-[linear-gradient(180deg,rgba(15,23,42,0.92),rgba(2,6,23,0.82))] px-2.5 py-2 shadow-[0_0_0_1px_rgba(125,211,252,0.12)_inset,0_0_18px_rgba(34,211,238,0.08),0_10px_24px_rgba(2,6,23,0.42)] sm:rounded-xl sm:px-3 sm:py-3">'
        + '<div class="text-[10px] uppercase tracking-[0.18em] text-cyan-200/60 sm:text-xs sm:tracking-[0.24em]">' + label + '</div>'
        + '<div class="mt-1.5 font-mono text-[10px] text-cyan-100/55 sm:mt-2 sm:text-[11px]">' + value.toLocaleString() + '</div>'
        + '<div class="mt-0.5 font-mono text-xl font-semibold text-cyan-50 sm:mt-1 sm:text-2xl" title="' + value.toLocaleString() + ' tokens">' + formatTokens(value) + '</div>'
        + '</section>';
    }
    function period(title, usage) {
      usage = usage || {};
      return '<section class="rounded-xl bg-slate-950/40 p-2.5 sm:rounded-2xl sm:p-3">'
        + '<div class="mb-2 text-[11px] font-semibold uppercase tracking-[0.2em] text-cyan-100/75 sm:text-xs sm:tracking-[0.22em]">' + title + '</div>'
        + '<div class="grid grid-cols-3 gap-2 sm:gap-3">'
        + metric('Input', Number(usage.input_tokens || 0))
        + metric('Output', Number(usage.output_tokens || 0))
        + metric('Total', Number(usage.total_tokens || 0))
        + '</div></section>';
    }
    const usage = JSON.parse(document.getElementById('usage-data').textContent || '{}');
    document.getElementById('usage-grid').innerHTML = period('Today', usage.today) + period('Yesterday', usage.yesterday) + period('Total', usage.total);
  </script>
</body>
</html>
`
