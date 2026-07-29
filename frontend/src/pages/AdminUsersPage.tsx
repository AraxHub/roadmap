import { FormEvent, useEffect, useState } from 'react'
import {
  createUser,
  getFeedbackSchedule,
  getHome,
  listAdminUsers,
  listUserFeedback,
  requestUserFeedback,
  updateFeedbackMessage,
} from '@/api/client'
import type { AdminUser, CreatedUser, FeedbackItem, FeedbackSchedule, Role } from '@/api/types'
import { AppShell } from '@/components/AppShell'
import { CredentialsModal } from '@/components/CredentialsModal'
import { ElectricField } from '@/components/ElectricField'

function formatCountdown(seconds: number): string {
  if (seconds <= 0) return 'скоро'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  if (days > 0) return `${days} дн ${hours} ч`
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${hours} ч ${minutes} мин`
}

export function AdminUsersPage() {
  const [loginName, setLoginName] = useState('')
  const [role, setRole] = useState<Role>('user')
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)
  const [created, setCreated] = useState<CreatedUser | null>(null)
  const [users, setUsers] = useState<AdminUser[]>([])
  const [selected, setSelected] = useState<AdminUser | null>(null)
  const [feedback, setFeedback] = useState<FeedbackItem[]>([])
  const [roundBusy, setRoundBusy] = useState(false)
  const [botName, setBotName] = useState('')
  const [schedule, setSchedule] = useState<FeedbackSchedule | null>(null)
  const [secondsLeft, setSecondsLeft] = useState<number | null>(null)
  const [messageDraft, setMessageDraft] = useState('')
  const [messageBusy, setMessageBusy] = useState(false)
  const [messageSaved, setMessageSaved] = useState(false)

  const canRequestFeedback =
    !!selected && selected.role === 'user' && selected.telegram_linked && !roundBusy

  const messageDirty =
    schedule !== null && messageDraft.trim() !== schedule.message_text.trim()

  async function reloadUsers() {
    const res = await listAdminUsers()
    setUsers(res.users)
    if (selected) {
      const fresh = res.users.find((u) => u.id === selected.id) ?? null
      setSelected(fresh)
    }
  }

  useEffect(() => {
    void reloadUsers().catch((err: unknown) => {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки юзеров')
    })
    void getHome()
      .then((home) => setBotName(home.telegram_bot_username ?? ''))
      .catch(() => setBotName(''))
    void getFeedbackSchedule()
      .then((s) => {
        setSchedule(s)
        setSecondsLeft(s.seconds_to_next)
        setMessageDraft(s.message_text)
      })
      .catch(() => setSchedule(null))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (secondsLeft === null) return
    const id = window.setInterval(() => {
      setSecondsLeft((prev) => (prev === null ? prev : Math.max(0, prev - 60)))
    }, 60_000)
    return () => window.clearInterval(id)
  }, [secondsLeft === null])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setPending(true)
    try {
      const user = await createUser(loginName.trim(), role)
      setCreated(user)
      setLoginName('')
      await reloadUsers()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка')
    } finally {
      setPending(false)
    }
  }

  async function openUser(u: AdminUser) {
    setSelected(u)
    setFeedback([])
    setError(null)
    try {
      const res = await listUserFeedback(u.id)
      setFeedback(res.feedback)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка истории ОС')
    }
  }

  async function onRequestFeedback() {
    if (!selected) return
    setRoundBusy(true)
    setError(null)
    try {
      await requestUserFeedback(selected.id)
      await reloadUsers()
      const res = await listUserFeedback(selected.id)
      setFeedback(res.feedback)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось запросить ОС')
    } finally {
      setRoundBusy(false)
    }
  }

  async function onSaveMessage() {
    setMessageBusy(true)
    setError(null)
    setMessageSaved(false)
    try {
      const s = await updateFeedbackMessage(messageDraft)
      setSchedule(s)
      setSecondsLeft(s.seconds_to_next)
      setMessageDraft(s.message_text)
      setMessageSaved(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить текст ОС')
    } finally {
      setMessageBusy(false)
    }
  }

  return (
    <div className="relative min-h-screen">
      <ElectricField />
      <AppShell title="Админ">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="font-display text-4xl font-extrabold">Личные кабинеты</h1>
            <p className="mt-3 max-w-xl text-muted">
              Создай логин и роль. Привязка Telegram: пользователь жмёт Start у{' '}
              {botName ? (
                <a
                  className="text-link"
                  href={`https://t.me/${botName}`}
                  target="_blank"
                  rel="noreferrer"
                >
                  @{botName}
                </a>
              ) : (
                'бота'
              )}
              , затем отдельным сообщением присылает свой логин.
            </p>
          </div>
          <div className="flex flex-col items-end gap-2">
            {schedule && (
              <div className="rounded-xl border border-line bg-line/30 px-4 py-2 text-right">
                <div className="text-xs uppercase tracking-wide text-muted">
                  До опроса обратной связи
                </div>
                <div className="font-display text-lg font-bold">
                  {secondsLeft !== null ? formatCountdown(secondsLeft) : '—'}
                </div>
                <div className="text-xs text-muted">
                  {new Date(schedule.next_round_at).toLocaleString('ru-RU', {
                    weekday: 'short',
                    day: '2-digit',
                    month: '2-digit',
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </div>
              </div>
            )}
            <button
              type="button"
              disabled={!canRequestFeedback}
              onClick={() => void onRequestFeedback()}
              title={
                !selected
                  ? 'Выбери пользователя слева'
                  : !selected.telegram_linked
                    ? 'У пользователя нет привязанного Telegram'
                    : selected.role !== 'user'
                      ? 'Админам ОС не шлём'
                      : undefined
              }
              className={`btn btn-sm ${
                canRequestFeedback
                  ? 'btn-secondary'
                  : 'cursor-not-allowed border-line bg-line/50 text-muted opacity-70'
              }`}
            >
              {roundBusy ? 'Отправка…' : 'Запросить ОС сейчас'}
            </button>
          </div>
        </div>

        <div className="mt-6 grid gap-6 lg:grid-cols-2">
          <form
            onSubmit={(e) => void onSubmit(e)}
            className="space-y-4 rounded-3xl border border-electric/25 bg-surface/90 p-6"
          >
            <h2 className="font-display text-xl font-bold">Новый кабинет</h2>
            <label className="block text-sm text-muted">
              Логин
              <input
                className="mt-1.5 w-full rounded-xl border border-line bg-bg px-3 py-2.5 outline-none focus:border-electric"
                value={loginName}
                onChange={(e) => setLoginName(e.target.value)}
                required
              />
            </label>
            <label className="block text-sm text-muted">
              Роль
              <select
                className="mt-1.5 w-full rounded-xl border border-line bg-bg px-3 py-2.5 outline-none focus:border-electric"
                value={role}
                onChange={(e) => setRole(e.target.value as Role)}
              >
                <option value="user">user</option>
                <option value="admin">admin</option>
              </select>
            </label>
            {error ? <p className="text-sm text-danger">{error}</p> : null}
            <button type="submit" disabled={pending} className="btn btn-primary btn-lg w-full">
              {pending ? 'Создаём…' : 'Создать кабинет'}
            </button>
          </form>

          {schedule ? (
            <div className="rounded-3xl border border-line bg-surface/90 p-5">
              <h2 className="font-display text-xl font-bold">Текст просьбы об ОС</h2>
              <textarea
                className="mt-3 min-h-[110px] w-full rounded-xl border border-line bg-bg px-3 py-2.5 outline-none focus:border-electric"
                value={messageDraft}
                onChange={(e) => {
                  setMessageDraft(e.target.value)
                  setMessageSaved(false)
                }}
                rows={4}
              />
              <div className="mt-3 flex flex-wrap items-center gap-3">
                <button
                  type="button"
                  disabled={messageBusy || !messageDirty || !messageDraft.trim()}
                  onClick={() => void onSaveMessage()}
                  className={`btn btn-sm ${
                    messageBusy || !messageDirty || !messageDraft.trim()
                      ? 'cursor-not-allowed border-line bg-line/50 text-muted opacity-70'
                      : 'btn-primary'
                  }`}
                >
                  {messageBusy ? 'Сохраняем…' : 'Сохранить текст'}
                </button>
                {messageSaved ? <span className="text-sm text-muted">Сохранено</span> : null}
              </div>
            </div>
          ) : (
            <div className="rounded-3xl border border-line bg-surface/90 p-5 text-sm text-muted">
              Загрузка текста ОС…
            </div>
          )}
        </div>

        <div className="mt-10 grid gap-6 lg:grid-cols-[1fr_1.2fr]">
          <div className="rounded-3xl border border-line bg-surface/90 p-5">
            <h2 className="font-display text-xl font-bold">Пользователи</h2>
            <ul className="mt-4 space-y-2">
              {users.map((u) => (
                <li key={u.id}>
                  <button
                    type="button"
                    onClick={() => void openUser(u)}
                    className={`clickable-row flex w-full items-center justify-between rounded-xl px-3 py-2.5 text-left ${
                      selected?.id === u.id ? 'border-electric/40 bg-electric-soft/70' : ''
                    }`}
                  >
                    <span className="min-w-0">
                      <span className="font-medium">{u.login}</span>
                      <span className="ml-2 text-xs text-muted">{u.role}</span>
                      {u.role === 'user' ? (
                        <span className="mt-0.5 block truncate text-xs text-muted">
                          {u.stage_status === 'completed'
                            ? 'Завершён'
                            : u.stage_status === 'in_progress' && u.current_stage
                              ? `${u.current_stage.sprint_title} · ${u.current_stage.module_title} · ${u.current_stage.submodule_title}`
                              : 'Нет опубликованного курса'}
                        </span>
                      ) : null}
                    </span>
                    <span className="shrink-0 text-xs text-muted">
                      {u.is_blocked ? 'блок' : u.telegram_linked ? 'TG' : '—'}
                    </span>
                  </button>
                </li>
              ))}
              {users.length === 0 ? <li className="text-sm text-muted">Пока никого нет</li> : null}
            </ul>
          </div>

          <div className="rounded-3xl border border-line bg-surface/90 p-5">
            <h2 className="font-display text-xl font-bold">
              История ОС{selected ? ` · ${selected.login}` : ''}
            </h2>
            {!selected ? (
              <p className="mt-4 text-sm text-muted">Выбери пользователя слева</p>
            ) : feedback.length === 0 ? (
              <p className="mt-4 text-sm text-muted">Ответов пока нет</p>
            ) : (
              <ul className="mt-4 space-y-3">
                {feedback.map((f) => (
                  <li key={f.id} className="rounded-2xl border border-line bg-bg/60 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-muted">
                      <span>{f.status}</span>
                      <span>{new Date(f.requested_at).toLocaleString()}</span>
                    </div>
                    {f.answer_text ? (
                      <p className="mt-2 whitespace-pre-wrap text-sm">{f.answer_text}</p>
                    ) : (
                      <p className="mt-2 text-sm text-muted">Ждём ответ…</p>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </AppShell>

      {created ? <CredentialsModal user={created} onClose={() => setCreated(null)} /> : null}
    </div>
  )
}
