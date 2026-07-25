import { useState } from 'react'
import type { CreatedUser } from '@/api/types'

export function CredentialsModal({
  user,
  onClose,
}: {
  user: CreatedUser
  onClose: () => void
}) {
  const [copied, setCopied] = useState(false)
  const payload = `login: ${user.login}\npassword: ${user.password}`

  async function copy() {
    await navigator.clipboard.writeText(payload)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1800)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-ink/40 p-4 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-2xl border border-electric/30 bg-surface p-6 shadow-[0_20px_60px_rgba(0,119,182,0.18)]">
        <h2 className="font-display text-2xl font-bold">Личный кабинет создан</h2>
        <p className="mt-2 text-sm text-muted">
          Пароль показывается один раз. Скопируй и передай человеку.
        </p>
        <div className="mt-5 space-y-3 rounded-xl border border-line bg-bg p-4 font-mono text-sm">
          <div>
            <div className="text-xs uppercase tracking-wide text-muted">login</div>
            <div className="mt-1 text-ink">{user.login}</div>
          </div>
          <div>
            <div className="text-xs uppercase tracking-wide text-muted">password</div>
            <div className="mt-1 text-ink">{user.password}</div>
          </div>
          <div className="text-xs text-muted">role: {user.role}</div>
        </div>
        <div className="mt-5 flex flex-col gap-2 sm:flex-row">
          <button type="button" onClick={() => void copy()} className="btn btn-primary btn-md flex-1">
            {copied ? 'Скопировано' : 'Скопировать'}
          </button>
          <button type="button" onClick={onClose} className="btn btn-ghost btn-md">
            Закрыть
          </button>
        </div>
      </div>
    </div>
  )
}
