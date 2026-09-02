import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ApiError, fetchDocVaults, type DocVault } from '../../api'
import './Docs.css'

const DocsVaults = () => {
  const [vaults, setVaults] = useState<DocVault[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    setIsLoading(true)
    fetchDocVaults({ signal: controller.signal })
      .then((response) => {
        setVaults(response.vaults ?? [])
        setError(null)
      })
      .catch((err) => {
        if (!controller.signal.aborted) {
          setError(err instanceof ApiError ? err.message : 'Docsの取得に失敗しました')
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsLoading(false)
        }
      })
    return () => controller.abort()
  }, [])

  return (
    <main className="docs-page">
      <div className="docs-shell">
        <header className="docs-header">
          <div>
            <h2>Docs</h2>
            <p>公開中のノートVault</p>
          </div>
        </header>
        {isLoading && <p className="docs-empty">読み込み中...</p>}
        {error && <p className="docs-error">{error}</p>}
        {!isLoading && !error && vaults.length === 0 && (
          <p className="docs-empty">公開中のVaultはありません。</p>
        )}
        <div className="docs-grid">
          {vaults.map((vault) => (
            <Link key={vault.slug} className="docs-card" to={`/docs/${vault.slug}`}>
              <h3>{vault.title}</h3>
              {vault.last_synced_at && (
                <p>更新: {new Date(vault.last_synced_at).toLocaleString()}</p>
              )}
            </Link>
          ))}
        </div>
      </div>
    </main>
  )
}

export default DocsVaults
