#!/usr/bin/env bash
set -euo pipefail
# ===== Config =====
ISSUER="https://auth.openai.com"
CLIENT_ID="${CODEX_OAUTH_CLIENT_ID:-app_EMoamEEZ73f0CkXaXp7hrann}"
LOGIN_PAGE="${ISSUER}/codex/device"
DEVICE_CODE_URL="${ISSUER}/api/accounts/deviceauth/usercode"
POLL_URL="${ISSUER}/api/accounts/deviceauth/token"
TOKEN_URL="${ISSUER}/oauth/token"
REDIRECT_URI="${ISSUER}/deviceauth/callback"
MAX_WAIT_SECONDS=$((15 * 60)) # 15 min
# ===== Check deps =====
command -v curl >/dev/null || { echo "curl not found"; exit 1; }
command -v jq >/dev/null || { echo "jq not found"; exit 1; }
TMP_BODY="$(mktemp)"
trap 'rm -f "$TMP_BODY"' EXIT
echo "== Step 1: Request device code =="
HTTP_CODE="$(curl -sS -o "$TMP_BODY" -w "%{http_code}" \
  -X POST "$DEVICE_CODE_URL" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "$(jq -nc --arg cid "$CLIENT_ID" '{client_id:$cid}')")"
if [[ "$HTTP_CODE" != "200" ]]; then
  echo "Device code request failed: HTTP $HTTP_CODE"
  cat "$TMP_BODY"
  exit 1
fi
USER_CODE="$(jq -r '.user_code // empty' "$TMP_BODY")"
DEVICE_AUTH_ID="$(jq -r '.device_auth_id // empty' "$TMP_BODY")"
INTERVAL="$(jq -r '.interval // 5' "$TMP_BODY")"
[[ "$INTERVAL" =~ ^[0-9]+$ ]] || INTERVAL=5
(( INTERVAL < 3 )) && INTERVAL=3
if [[ -z "$USER_CODE" || -z "$DEVICE_AUTH_ID" ]]; then
  echo "Missing user_code or device_auth_id in response:"
  cat "$TMP_BODY"
  exit 1
fi
echo
echo "Open this page in browser:"
echo "  $LOGIN_PAGE"
echo
echo "Enter this code:"
echo "  $USER_CODE"
echo
echo "Waiting for authorization (poll every ${INTERVAL}s)..."
# ===== Step 2: Poll for authorization_code + code_verifier =====
START_TS="$(date +%s)"
AUTH_CODE=""
CODE_VERIFIER=""
while true; do
  NOW_TS="$(date +%s)"
  ELAPSED=$((NOW_TS - START_TS))
  if (( ELAPSED > MAX_WAIT_SECONDS )); then
    echo
    echo "Timed out after ${MAX_WAIT_SECONDS}s"
    exit 1
  fi
  sleep "$INTERVAL"
  HTTP_CODE="$(curl -sS -o "$TMP_BODY" -w "%{http_code}" \
    -X POST "$POLL_URL" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "$(jq -nc \
      --arg did "$DEVICE_AUTH_ID" \
      --arg ucode "$USER_CODE" \
      '{device_auth_id:$did,user_code:$ucode}')")"
  if [[ "$HTTP_CODE" == "200" ]]; then
    AUTH_CODE="$(jq -r '.authorization_code // empty' "$TMP_BODY")"
    CODE_VERIFIER="$(jq -r '.code_verifier // empty' "$TMP_BODY")"
    if [[ -n "$AUTH_CODE" && -n "$CODE_VERIFIER" ]]; then
      echo
      echo "Authorization confirmed."
      break
    fi
    echo
    echo "HTTP 200 but missing authorization_code/code_verifier:"
    cat "$TMP_BODY"
    exit 1
  elif [[ "$HTTP_CODE" == "403" || "$HTTP_CODE" == "404" ]]; then
    printf "."
    continue
  else
    echo
    echo "Polling failed: HTTP $HTTP_CODE"
    cat "$TMP_BODY"
    exit 1
  fi
done
# ===== Step 3: Exchange authorization_code for tokens =====
echo "== Step 3: Exchange token =="
HTTP_CODE="$(curl -sS -o "$TMP_BODY" -w "%{http_code}" \
  -X POST "$TOKEN_URL" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -H "Accept: application/json" \
  --data-urlencode "grant_type=authorization_code" \
  --data-urlencode "code=${AUTH_CODE}" \
  --data-urlencode "redirect_uri=${REDIRECT_URI}" \
  --data-urlencode "client_id=${CLIENT_ID}" \
  --data-urlencode "code_verifier=${CODE_VERIFIER}")"
if [[ "$HTTP_CODE" != "200" ]]; then
  echo "Token exchange failed: HTTP $HTTP_CODE"
  cat "$TMP_BODY"
  exit 1
fi
echo "Login success. Token response:"
cat "$TMP_BODY" | jq .
echo
echo "Tip: save safely, do NOT commit tokens to git."