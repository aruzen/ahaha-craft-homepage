import type { ToyEntry } from '../../../types/toy'

export const rustGoGateway: ToyEntry = {
  id: 'toy-2',
  title: 'Rust×Go Hybrid API Gateway',
  summary: '高負荷APIをRustで、業務ロジックをGoで記述するハイブリッド構成の検証。',
  category: 'reference',
  tags: ['backend', 'performance', 'tooling'],
  difficulty: 'advanced',
  lastUpdated: '2025-07-18',
  heroImage: '/resource/toy-space/gateway.png',
  repositoryUrl: 'https://github.com/aruzen/rust-go-gateway',
  slug: 'rust-go-gateway',
  content: `## なぜハイブリッド？

- Rust: TLS終端とレート制御を担い GCレスで低レイテンシ
- Go: UseCaseレイヤの実装速度とライブラリ資産を重視

## 構成図

1. Rust (Axum) で HTTP/2 + QUIC を受ける
2. Wasmtime 経由で Go 側に context を渡す
3. Go (Fiber) がビジネスロジックを処理

## Rust 側 Middleware

\`\`\`rust
pub async fn guard(req: Request<Body>) -> Result<Response<Body>, Infallible> {
    if !token_pool.validate(&req) {
        return Ok(Response::builder()
            .status(StatusCode::TOO_MANY_REQUESTS)
            .body("retry".into())?);
    }
    Ok(next.run(req).await)
}
\`\`\`

## 課題とTODO

- ✅ Observability を OpenTelemetry で統合
- ⏳ Go 側 worker を Wasm 化
- 🔜 eBPF でシステムコール観測
`,
}
