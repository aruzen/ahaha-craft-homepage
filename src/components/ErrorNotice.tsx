import './ErrorNotice.css'

export interface ErrorDescriptor {
  message: string
  field?: string
  code?: string
}

interface ErrorNoticeProps extends ErrorDescriptor {
  title?: string
  onDismiss?: () => void
}

const ErrorNotice = ({ title = 'エラーが発生しました', message, field, code, onDismiss }: ErrorNoticeProps) => (
  <div className="error-notice">
    <div>
      <p className="error-title">{title}</p>
      <p className="error-message">
        {field ? <strong className="error-field">[{field}]</strong> : null}
        {message}
      </p>
      {code && <p className="error-code">code: {code}</p>}
    </div>
    {onDismiss && (
      <button className="error-dismiss" type="button" onClick={onDismiss} aria-label="閉じる">
        ×
      </button>
    )}
  </div>
)

export default ErrorNotice
