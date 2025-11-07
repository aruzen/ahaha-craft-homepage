# 認証 API

## POST /api/login

AhahaCraft のログインモーダルから呼び出される認証エンドポイントです。

### リクエスト
- **HTTP メソッド**: `POST`
- **Content-Type**: `application/json`
- **ボディ**
  ```json
  {
    "username": "ユーザー名",
    "password": "パスワード"
  }
  ```

### レスポンス
- **ステータス 200 OK**
  ```json
  {
    "token": "JWT もしくはセッショントークン",
    "isAdmin": true,
    "expiresAt": "2024-12-31T23:59:59Z"
  }
  ```
  - `token`: 認証トークン。クライアントは後続リクエストの `Authorization` ヘッダー (例: `Bearer <token>`) に設定します。
  - `isAdmin`: ユーザーが管理者の場合は `true`。
  - `expiresAt`: ISO 8601 形式のトークン有効期限。

### エラー
| ステータス | 説明 |
|------------|------|
| 400 Bad Request | リクエストボディが不正、または必須項目が欠落している場合 |
| 401 Unauthorized | 認証情報が無効な場合 |
| 429 Too Many Requests | レート制限に達した場合 |
| 500 Internal Server Error | サーバー内部でエラーが発生した場合 |

エラー時のレスポンス例:
```json
{
  "error": "invalid_credentials",
  "message": "ユーザー名またはパスワードが正しくありません"
}
```

### トークンの有効期限と更新
- `expiresAt` で示される期限を過ぎたトークンは無効になります。
- トークンの更新は以下のいずれかで行ってください。
  - リフレッシュトークンを使用する場合: `/api/token/refresh` に `POST` し、新しいアクセストークンを取得します。
  - リフレッシュトークンを用意していない場合: `/api/login` を再度呼び出し、再認証します。
- クライアント側では期限切れ前に更新を行うか、401 応答を受け取った際にログアウト処理を行ってください。
