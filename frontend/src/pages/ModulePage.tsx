import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { getModule } from '@/api/client'
import type { ModuleView } from '@/api/types'
import { AppShell } from '@/components/AppShell'
import { ElectricField } from '@/components/ElectricField'

export function ModulePage() {
  const { moduleSlug = '' } = useParams()
  const [data, setData] = useState<ModuleView | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    getModule(moduleSlug)
      .then((v) => {
        if (!cancelled) setData(v)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Ошибка')
      })
    return () => {
      cancelled = true
    }
  }, [moduleSlug])

  return (
    <div className="relative min-h-screen">
      <ElectricField />
      <AppShell title="Модуль">
        <Link to="/" className="text-link">
          ← к модулям
        </Link>

        {error ? <p className="mt-6 text-danger">{error}</p> : null}
        {!data && !error ? <p className="mt-6 text-muted">Загрузка…</p> : null}

        {data ? (
          <>
            <h1 className="mt-4 font-display text-4xl font-extrabold">{data.module.title}</h1>
            {data.module.description ? (
              <p className="mt-3 max-w-2xl text-muted">{data.module.description}</p>
            ) : null}

            <ul className="mt-10 space-y-3">
              {data.submodules.map((sm, idx) => {
                const row = (
                  <div
                    className={`flex items-center justify-between gap-4 rounded-2xl border px-5 py-4 ${
                      sm.unlocked
                        ? 'clickable-card border-electric/30 bg-surface'
                        : 'border-line bg-surface/50 opacity-55'
                    }`}
                  >
                    <div className="flex items-center gap-4">
                      <span className="font-display text-lg text-electric-deep">
                        {String(idx + 1).padStart(2, '0')}
                      </span>
                      <div>
                        <div className="font-medium">{sm.title}</div>
                        <div className="text-xs text-muted">
                          {sm.completed ? 'завершён' : sm.unlocked ? 'доступен' : 'закрыт'}
                        </div>
                      </div>
                    </div>
                    {sm.unlocked ? (
                      <span className="text-sm text-electric-deep">открыть →</span>
                    ) : null}
                  </div>
                )

                return sm.unlocked ? (
                  <li key={sm.id}>
                    <Link to={`/modules/${moduleSlug}/submodules/${sm.slug}`}>{row}</Link>
                  </li>
                ) : (
                  <li key={sm.id}>{row}</li>
                )
              })}
            </ul>
          </>
        ) : null}
      </AppShell>
    </div>
  )
}
