import { useCallback, useEffect, useMemo, useState, type FormEvent } from 'react'
import {
  ApiError,
  disableDocVault,
  fetchAdminDocBranches,
  fetchAdminDocVaults,
  fetchHueAreYouRecords,
  registerDocVault,
  rescanDocVault,
  syncDocVault,
  type DocVault,
  type HueAreYouRecord,
  type SessionData,
  type SessionResponce,
} from '../../api'
import { colorToHex } from '../../data/colors'
import ErrorNotice, { type ErrorDescriptor } from '../../components/ErrorNotice'
import './AdminDashboard.css'

type AdminNavItem = 'user-management' | 'hue-results' | 'docs'

interface AdminDashboardProps {
  username: string
  session: SessionResponce
}

const sections = [
  {
    id: 'general',
    label: 'General',
    items: [
      {
        id: 'user-management' as AdminNavItem,
        label: 'ユーザー管理',
        description: 'ユーザーデータの閲覧・編集機能（近日追加）',
      },
    ],
  },
  {
    id: 'hue',
    label: 'Hue Are You',
    items: [
      {
        id: 'hue-results' as AdminNavItem,
        label: '結果参照',
        description: 'hue-are-you APIの結果データを取得',
      },
    ],
  },
  {
    id: 'docs',
    label: 'Docs',
    items: [
      {
        id: 'docs' as AdminNavItem,
        label: 'Docs管理',
        description: 'Git branch vaultの登録・同期',
      },
    ],
  },
]

const AdminDashboard = ({ username, session }: AdminDashboardProps) => {
  const [activeItem, setActiveItem] = useState<AdminNavItem>('user-management')

  const sessionPayload = useMemo<SessionData>(
    () => ({
      user_id: session.user_id,
      token: session.token,
    }),
    [session]
  )

  return (
    <div className="admin-dashboard">
      <aside className="admin-sidebar">
        <div className="admin-profile">
          <span className="profile-label">ログイン中</span>
          <strong>{username}</strong>
          <span className="profile-role">role: {session.role}</span>
        </div>
        {sections.map((section) => (
          <div key={section.id} className="nav-section">
            <p className="nav-section-title">{section.label}</p>
            <ul>
              {section.items.map((item) => (
                <li key={item.id}>
                  <button
                    type="button"
                    className={item.id === activeItem ? 'nav-item active' : 'nav-item'}
                    onClick={() => setActiveItem(item.id)}
                  >
                    <span className="nav-item-label">{item.label}</span>
                    <span className="nav-item-description">{item.description}</span>
                  </button>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </aside>
      <main className="admin-content">
        {activeItem === 'user-management' && <UserManagementPlaceholder />}
        {activeItem === 'hue-results' && <HueResultsPanel session={sessionPayload} />}
        {activeItem === 'docs' && <DocsAdminPanel session={sessionPayload} />}
      </main>
    </div>
  )
}

const UserManagementPlaceholder = () => (
  <section className="admin-card">
    <header>
      <h2>ユーザー管理</h2>
      <p>ユーザー管理ツールは現在設計中です。今後のリリースをお待ちください。</p>
    </header>
    <ul className="todo-list">
      <li>ユーザー検索・フィルタ</li>
      <li>ロール変更 / 無効化操作</li>
      <li>活動ログの確認</li>
    </ul>
  </section>
)

interface HueResultsPanelProps {
  session: SessionData
}

const HueResultsPanel = ({ session }: HueResultsPanelProps) => {
  const [rangeStart, setRangeStart] = useState(0)
  const [rangeEnd, setRangeEnd] = useState(24)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErrorDescriptor | null>(null)
  const [records, setRecords] = useState<HueAreYouRecord[]>([])
  const [lastFetchedAt, setLastFetchedAt] = useState<Date | null>(null)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    if (rangeStart < 0 || rangeEnd < 0 || rangeStart > rangeEnd) {
      setError({ message: 'データ範囲が不正です', field: 'data-range' })
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const response = await fetchHueAreYouRecords({
        session,
        dataRange: [rangeStart, rangeEnd],
      })

      setRecords(response.records ?? [])
      setLastFetchedAt(new Date())
    } catch (err) {
      if (err instanceof ApiError) {
        setError({ message: err.message, field: err.field, code: err.code })
      } else {
        setError({ message: 'データ取得に失敗しました' })
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <section className="admin-card">
      <header className="results-header">
        <div>
          <h2>Hue Are You 結果参照</h2>
          <p>APIから結果データを取得して確認できます。</p>
        </div>
        {lastFetchedAt && (
          <span className="timestamp">最終更新: {lastFetchedAt.toLocaleString()}</span>
        )}
      </header>

      <form className="results-form" onSubmit={handleSubmit}>
        <label>
          開始インデックス
          <input
            type="number"
            value={rangeStart}
            min={0}
            onChange={(e) => setRangeStart(Number(e.target.value))}
          />
        </label>
        <label>
          終了インデックス
          <input
            type="number"
            value={rangeEnd}
            min={0}
            onChange={(e) => setRangeEnd(Number(e.target.value))}
          />
        </label>
        <button type="submit" disabled={isLoading}>
          {isLoading ? '取得中...' : 'データ取得'}
        </button>
      </form>

      {error && <ErrorNotice {...error} onDismiss={() => setError(null)} />}

      <div className="records-list">
        {records.length === 0 && !isLoading && (
          <p className="empty-state">データがありません。条件を設定して取得してください。</p>
        )}
        {records.map((record, index) => (
          <details key={`${record.name}-${index}`} className="record-card">
            <summary>
              <span>
                #{index + 1} {record.name}
              </span>
              <span className="word-count">{Object.keys(record.choice).length}語</span>
            </summary>
            <div className="record-body">
              {Object.entries(record.choice).map(([word, color]) => (
                <span
                  key={`${word}-${color}`}
                  className="word-chip"
                  style={{ backgroundColor: colorToHex[color as keyof typeof colorToHex] ?? '#eee' }}
                >
                  {word}
                  <strong>{color}</strong>
                </span>
              ))}
            </div>
          </details>
        ))}
      </div>
    </section>
  )
}

interface DocsAdminPanelProps {
  session: SessionData
}

const DocsAdminPanel = ({ session }: DocsAdminPanelProps) => {
  const [branches, setBranches] = useState<string[]>([])
  const [vaults, setVaults] = useState<DocVault[]>([])
  const [branch, setBranch] = useState('')
  const [slug, setSlug] = useState('')
  const [title, setTitle] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<ErrorDescriptor | null>(null)
  const [message, setMessage] = useState<string | null>(null)

  const loadDocsAdminData = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [branchResponse, vaultResponse] = await Promise.all([
        fetchAdminDocBranches(session),
        fetchAdminDocVaults(session),
      ])
      setBranches(branchResponse.branches ?? [])
      setVaults(vaultResponse.vaults ?? [])
      setBranch((current) => current || branchResponse.branches?.[0] || '')
    } catch (err) {
      setError(err instanceof ApiError ? { message: err.message, field: err.field, code: err.code } : { message: 'Docs管理情報の取得に失敗しました' })
    } finally {
      setIsLoading(false)
    }
  }, [session])

  const handleRegister = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!branch) {
      setError({ message: 'branchを選択してください', field: 'branch' })
      return
    }
    setIsLoading(true)
    setError(null)
    setMessage(null)
    try {
      await registerDocVault({ session, branch, slug: slug || undefined, title: title || undefined })
      setSlug('')
      setTitle('')
      setMessage('Vaultを登録しました')
      await loadDocsAdminData()
    } catch (err) {
      setError(err instanceof ApiError ? { message: err.message, field: err.field, code: err.code } : { message: 'Vault登録に失敗しました' })
    } finally {
      setIsLoading(false)
    }
  }

  const runVaultAction = async (vaultSlug: string, action: 'sync' | 'rescan' | 'disable') => {
    setIsLoading(true)
    setError(null)
    setMessage(null)
    try {
      if (action === 'sync') {
        await syncDocVault(session, vaultSlug)
        setMessage(`${vaultSlug} を同期しました`)
      } else if (action === 'rescan') {
        await rescanDocVault(session, vaultSlug)
        setMessage(`${vaultSlug} を再スキャンしました`)
      } else {
        await disableDocVault(session, vaultSlug)
        setMessage(`${vaultSlug} を無効化しました`)
      }
      await loadDocsAdminData()
    } catch (err) {
      setError(err instanceof ApiError ? { message: err.message, field: err.field, code: err.code } : { message: 'Vault操作に失敗しました' })
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    void loadDocsAdminData()
  }, [loadDocsAdminData])

  return (
    <section className="admin-card docs-admin-card">
      <header className="results-header">
        <div>
          <h2>Docs管理</h2>
          <p>ローカルGitリポジトリのbranchをVaultとして登録・同期します。</p>
        </div>
        <button type="button" className="secondary-btn" onClick={loadDocsAdminData} disabled={isLoading}>
          再読み込み
        </button>
      </header>

      <form className="results-form" onSubmit={handleRegister}>
        <label>
          branch
          <select value={branch} onChange={(e) => setBranch(e.target.value)}>
            {branches.length === 0 && <option value="">branchなし</option>}
            {branches.map((branchName) => (
              <option key={branchName} value={branchName}>
                {branchName}
              </option>
            ))}
          </select>
        </label>
        <label>
          slug
          <input value={slug} onChange={(e) => setSlug(e.target.value)} placeholder="未指定ならbranch名から生成" />
        </label>
        <label>
          title
          <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="未指定ならbranch名" />
        </label>
        <button type="submit" disabled={isLoading || !branch}>
          登録
        </button>
      </form>

      {error && <ErrorNotice {...error} onDismiss={() => setError(null)} />}
      {message && <p className="success-banner">{message}</p>}

      <div className="records-list">
        {vaults.length === 0 && !isLoading && <p className="empty-state">登録済みVaultはありません。</p>}
        {vaults.map((vault) => (
          <div key={vault.slug} className="doc-vault-admin-row">
            <div>
              <strong>{vault.title}</strong>
              <span>{vault.slug}</span>
              <span>branch: {vault.branch}</span>
              <span>status: {vault.status}</span>
              {vault.last_synced_at && <span>synced: {new Date(vault.last_synced_at).toLocaleString()}</span>}
            </div>
            <div className="doc-vault-actions">
              <button type="button" onClick={() => runVaultAction(vault.slug, 'sync')} disabled={isLoading}>
                同期
              </button>
              <button type="button" onClick={() => runVaultAction(vault.slug, 'rescan')} disabled={isLoading}>
                再スキャン
              </button>
              <button type="button" className="danger-btn" onClick={() => runVaultAction(vault.slug, 'disable')} disabled={isLoading}>
                無効化
              </button>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}

export default AdminDashboard
