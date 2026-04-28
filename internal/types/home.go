package types

const HomeTmpl = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent"/>
  <meta name="format-detection" content="telephone=no, email=no" >
  <title>Codex Console</title>
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
      min-height: 100%;
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
<body class="relative flex min-h-[100dvh] items-center justify-center overflow-hidden bg-gradient-to-br from-slate-950 via-zinc-950 to-black px-4 py-6 font-display text-cyan-100">
  <div class="pointer-events-none absolute inset-0 overflow-hidden">
    <div class="absolute -left-24 -top-24 h-80 w-80 rounded-full bg-sky-500/20 blur-3xl"></div>
    <div class="absolute -bottom-24 right-0 h-80 w-80 rounded-full bg-cyan-400/15 blur-3xl"></div>
  </div>
  <main class="relative w-full max-w-xl overflow-hidden rounded-2xl border border-cyan-300/40 bg-slate-950/70 px-5 py-6 shadow-[0_0_0_1px_rgba(125,211,252,0.15)_inset,0_24px_60px_rgba(2,6,23,0.85),0_0_42px_rgba(6,182,212,0.24)] backdrop-blur sm:px-7 sm:py-8">
    <div class="mb-3 text-[11px] uppercase tracking-[0.28em] text-cyan-200/70">Codex Console</div>
    <h1 class="text-3xl font-semibold tracking-tight text-cyan-50 sm:text-4xl">Manage Codex access</h1>
    <p class="mt-3 text-sm leading-6 text-cyan-100/70">Enter your API key, then add a Codex session with device authorization or inspect token usage.</p>
    <label class="mt-5 block text-[11px] uppercase tracking-[0.22em] text-cyan-200/70" for="sk-input">Secret Key</label>
    <input id="sk-input" type="text" autocomplete="off" placeholder="sk..." class="mt-2 w-full rounded-xl border border-cyan-300/35 bg-slate-950/80 px-3 py-3 font-mono text-sm text-cyan-50 outline-none transition placeholder:text-cyan-100/30 focus:border-cyan-200/80 focus:ring-2 focus:ring-cyan-300/20">
    <div id="sk-status" class="mt-2 min-h-4 text-xs text-rose-300"></div>
    <div class="mt-6 grid gap-3 sm:grid-cols-2">
      <a id="cli-link" href="/codex/cli" class="rounded-xl border border-cyan-200/50 bg-cyan-300 px-4 py-4 text-slate-950 shadow-[0_10px_28px_rgba(34,211,238,0.25)] transition hover:-translate-y-0.5 hover:shadow-[0_14px_34px_rgba(34,211,238,0.35)]">
        <div class="text-sm font-semibold">Device Login</div>
        <div class="mt-1 text-xs text-slate-800/75">Add a new session</div>
      </a>
      <a id="usage-link" href="/codex/usage" class="rounded-xl border border-cyan-300/35 bg-slate-900/70 px-4 py-4 text-cyan-50 shadow-[0_10px_28px_rgba(2,6,23,0.35)] transition hover:-translate-y-0.5 hover:border-cyan-200/60">
        <div class="text-sm font-semibold">Token Usage</div>
        <div class="mt-1 text-xs text-cyan-100/60">View token usage</div>
      </a>
    </div>
    <div class="pointer-events-none absolute inset-0 bg-[linear-gradient(110deg,transparent_38px,rgba(103,232,249,0.16)_50vw,transparent_62px)] animate-[sweep_5.4s_linear_infinite]"></div>
  </main>
  <script>
    const skInput = document.getElementById('sk-input');
    const skStatus = document.getElementById('sk-status');
    const cliLink = document.getElementById('cli-link');
    const usageLink = document.getElementById('usage-link');
    function requireSK(event) {
      if (skInput.value.trim()) return;
      event.preventDefault();
      skStatus.textContent = 'Secret Key is required.';
      skInput.focus();
    }
    function updateLinks() {
      const auth = skInput.value.trim();
      const query = auth ? '?auth=' + encodeURIComponent(auth) : '';
      cliLink.href = '/codex/cli' + query;
      usageLink.href = '/codex/usage' + query;
      skStatus.textContent = '';
    }
    cliLink.addEventListener('click', requireSK);
    usageLink.addEventListener('click', requireSK);
    skInput.addEventListener('input', updateLinks);
    updateLinks();
  </script>
</body>
</html>
`
