可以，按你要的 Codex 授权登录链路，关键请求是这 3 个（外加刷新 1 个）：
- 1) 申请 device code
  - POST https://auth.openai.com/api/accounts/deviceauth/usercode
  - Headers:
    - Content-Type: application/json
  - Body:
{
  "client_id": "app_EMoamEEZ73f0CkXaXp7hrann"
}
https://auth.openai.com/codex/device
- 2) 轮询用户是否完成授权
  - POST https://auth.openai.com/api/accounts/deviceauth/token
  - Headers:
    - Content-Type: application/json
  - Body:
{
  "device_auth_id": "<上一步返回的 device_auth_id>",
  "user_code": "<上一步返回的 user_code>"
}
- 3) 用 authorization_code 换 access/refresh token
  - POST https://auth.openai.com/oauth/token
  - Headers:
    - Content-Type: application/x-www-form-urlencoded
  - Body（form）:
grant_type=authorization_code
code=<轮询返回的 authorization_code>
redirect_uri=https://auth.openai.com/deviceauth/callback
client_id=app_EMoamEEZ73f0CkXaXp7hrann
code_verifier=<轮询返回的 code_verifier>
- 4) 后续 refresh token 刷新
  - POST https://auth.openai.com/oauth/token
  - Headers:
    - Content-Type: application/x-www-form-urlencoded
  - Body（form）:
grant_type=refresh_token
refresh_token=<当前 refresh_token>
client_id=app_EMoamEEZ73f0CkXaXp7hrann
如果你还要，我可以再给你一份可直接跑的 curl 模板（把变量位都留好）。
