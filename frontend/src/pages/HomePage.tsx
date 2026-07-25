import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { getHome } from '@/api/client'
import type { HomeView } from '@/api/types'
import { useAuth } from '@/auth/AuthContext'
import { AppShell } from '@/components/AppShell'
import { ElectricField } from '@/components/ElectricField'

function formatRemaining(deadlineISO: string, now = Date.now()): string {
  const ms = new Date(deadlineISO).getTime() - now
  if (ms <= 0) return '0 дн 0 ч'
  const totalHours = Math.floor(ms / (1000 * 60 * 60))
  const days = Math.floor(totalHours / 24)
  const hours = totalHours % 24
  return `${days} дн ${hours} ч`
}

export function HomePage() {
  const { user } = useAuth()
  const [data, setData] = useState<HomeView | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    getHome()
      .then((v) => {
        if (!cancelled) setData(v)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Ошибка')
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="relative min-h-screen">
      <ElectricField />
      <AppShell title="Главная">
        <section className="relative">
          <h1 className="font-display text-4xl font-extrabold tracking-tight sm:text-5xl">
            Модули обучения
          </h1>
          <p className="mt-3 max-w-2xl text-muted">
            Иди по цепочке спринт → модуль → подмодуль. Следующий раздел откроется после кнопки
            «Завершить».
          </p>
          {user?.role === 'admin' ? (
            <div className="mt-6 flex flex-wrap gap-3">
              <Link to="/admin/roadmap" className="btn btn-primary btn-md">
                Наполнить роадмап
              </Link>
              <Link to="/admin/users" className="btn btn-secondary btn-md">
                Создать ЛК
              </Link>
            </div>
          ) : null}
        </section>

        {error ? <p className="mt-8 text-danger">{error}</p> : null}
        {!data && !error ? <p className="mt-8 text-muted">Загрузка роадмапа…</p> : null}

        <div className="relative mt-10 space-y-10">
          {data?.feedback_required ? (
            <div className="absolute inset-0 z-20 flex items-start justify-center rounded-3xl bg-white/80 p-6 backdrop-blur-sm sm:p-10">
              <div className="max-w-md rounded-3xl border border-electric/30 bg-surface px-6 py-8 text-center shadow-[0_20px_60px_rgba(0,119,182,0.12)]">
                <div className="mx-auto mb-4 h-14 w-14 rounded-full border-2 border-electric-deep/40 bg-electric-soft" />
                <h2 className="font-display text-2xl font-bold">Нужна обратная связь</h2>
                <p className="mt-3 text-muted">
                  Роадмап временно закрыт. Ответь на сообщение бота
                  {data.telegram_bot_username ? (
                    <>
                      {' '}
                      <a
                        className="text-link"
                        href={`https://t.me/${data.telegram_bot_username}`}
                        target="_blank"
                        rel="noreferrer"
                      >
                        @{data.telegram_bot_username}
                      </a>
                    </>
                  ) : (
                    ' в Telegram'
                  )}{' '}
                  — после ответа доступ откроется снова.
                </p>
              </div>
            </div>
          ) : null}

          {data?.sprints.map((sprint) => (
            <section
              key={sprint.id}
              className={`rounded-2xl p-1 ${sprint.is_overdue ? 'ring-2 ring-danger/70' : ''}`}
            >
              <div
                className={`mb-4 flex items-end justify-between gap-4 border-b pb-3 ${
                  sprint.is_overdue ? 'border-danger/40' : 'border-line'
                }`}
              >
                <div>
                  <p
                    className={`text-xs uppercase tracking-[0.18em] ${
                      sprint.is_overdue ? 'text-danger' : 'text-electric-deep'
                    }`}
                  >
                    спринт
                  </p>
                  <h2
                    className={`font-display text-2xl font-bold ${
                      sprint.is_overdue ? 'text-danger' : ''
                    }`}
                  >
                    {sprint.title}
                  </h2>
                  {sprint.description ? (
                    <p className="mt-1 text-sm text-muted">{sprint.description}</p>
                  ) : null}
                  {sprint.show_timer && sprint.deadline_at ? (
                    <p
                      className={`mt-2 text-sm font-medium ${
                        sprint.is_overdue ? 'text-danger' : 'text-electric-deep'
                      }`}
                    >
                      {sprint.is_overdue
                        ? 'Срок вышел'
                        : `Осталось: ${formatRemaining(sprint.deadline_at)}`}
                    </p>
                  ) : null}
                </div>
                <span
                  className={`rounded-full px-3 py-1 text-xs font-medium ${
                    sprint.is_overdue
                      ? 'bg-danger/10 text-danger'
                      : sprint.unlocked
                        ? 'bg-electric-soft text-electric-deep'
                        : 'bg-line/60 text-muted'
                  }`}
                >
                  {sprint.is_overdue ? 'просрочен' : sprint.unlocked ? 'открыт' : 'закрыт'}
                </span>
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                {sprint.modules.map((mod) => {
                  const pct =
                    mod.total_count > 0
                      ? Math.round((mod.completed_count / mod.total_count) * 100)
                      : 0
                  const body = (
                    <div
                      className={`group relative overflow-hidden rounded-2xl border p-5 ${
                        mod.unlocked
                          ? 'clickable-card border-electric/30 bg-surface'
                          : 'border-line bg-surface/60 opacity-60'
                      }`}
                    >
                      <div className="absolute right-0 top-0 h-16 w-16 translate-x-6 -translate-y-6 rounded-full bg-electric/10 blur-2xl transition group-hover:bg-electric/20" />
                      <h3 className="font-display text-xl font-bold">{mod.title}</h3>
                      {mod.description ? (
                        <p className="mt-2 text-sm text-muted">{mod.description}</p>
                      ) : null}
                      <div className="mt-4 flex items-center justify-between text-xs text-muted">
                        <span>
                          {mod.completed_count}/{mod.total_count} · {pct}%
                        </span>
                        <span>{mod.unlocked ? 'войти →' : 'заблокировано'}</span>
                      </div>
                      <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-line">
                        <div
                          className="h-full rounded-full bg-electric transition-all"
                          style={{ width: `${pct}%` }}
                        />
                      </div>
                    </div>
                  )

                  return mod.unlocked ? (
                    <Link key={mod.id} to={`/modules/${mod.slug}`}>
                      {body}
                    </Link>
                  ) : (
                    <div key={mod.id}>{body}</div>
                  )
                })}
              </div>
            </section>
          ))}

          {data && data.sprints.length === 0 ? (
            <p className="rounded-2xl border border-dashed border-line bg-surface/70 p-8 text-muted">
              Роадмап пока пуст. Админ может наполнить его через API/конфиг БД.
            </p>
          ) : null}
        </div>
      </AppShell>
    </div>
  )
}
