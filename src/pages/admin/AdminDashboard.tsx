import { useMemo, useState, type FormEvent } from 'react'
import {
  ApiError,
  fetchHueAreYouRecords,
  type HueAreYouRecord,
  type SessionData,
  type SessionResponce,
} from '../../api'
import { colorToHex } from '../../data/colors'
import ErrorNotice, { type ErrorDescriptor } from '../../components/ErrorNotice'
import './AdminDashboard.css'

type AdminNavItem = 'user-management' | 'hue-results'

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
        {activeItem === 'user-management' ? (
          <UserManagementPlaceholder />
        ) : (
          <HueResultsPanel session={sessionPayload} />
        )}
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

export default AdminDashboard
