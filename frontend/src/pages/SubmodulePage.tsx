import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { completeSubmodule, getSubmodule } from '@/api/client'
import type { MenuItem, SubmoduleView } from '@/api/types'
import { ContentBlocks } from '@/components/ContentBlocks'
import { ElectricField } from '@/components/ElectricField'

function MenuTree({
  items,
  moduleSlug,
  activeSubSlug,
}: {
  items: MenuItem[]
  moduleSlug: string
  activeSubSlug: string
}) {
  return (
    <ul className="space-y-3 text-sm">
      {items.map((sprint) => (
        <li key={sprint.id}>
          <div className="mb-1 text-xs uppercase tracking-[0.16em] text-muted">{sprint.title}</div>
          <ul className="space-y-2 border-l border-line pl-3">
            {(sprint.children || []).map((mod) => (
              <li key={mod.id}>
                <div className={`mb-1 font-medium ${mod.unlocked ? 'text-ink' : 'text-muted'}`}>
                  {mod.title}
                </div>
                <ul className="space-y-1">
                  {(mod.children || []).map((sm) => {
                    const active = mod.slug === moduleSlug && sm.slug === activeSubSlug
                    const cls = `block rounded-lg border px-2 py-1.5 transition ${
                      active
                        ? 'border-electric/40 bg-electric-soft text-electric-deep'
                        : sm.unlocked
                          ? 'clickable-row border-transparent text-ink'
                          : 'cursor-not-allowed border-transparent text-muted/70'
                    }`
                    if (!sm.unlocked) {
                      return (
                        <li key={sm.id}>
                          <span className={cls}>{sm.title}</span>
                        </li>
                      )
                    }
                    return (
                      <li key={sm.id}>
                        <Link to={`/modules/${mod.slug}/submodules/${sm.slug}`} className={cls}>
                          {sm.title}
                          {sm.completed ? ' ✓' : ''}
                        </Link>
                      </li>
                    )
                  })}
                </ul>
              </li>
            ))}
          </ul>
        </li>
      ))}
    </ul>
  )
}

export function SubmodulePage() {
  const { moduleSlug = '', submoduleSlug = '' } = useParams()
  const [data, setData] = useState<SubmoduleView | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [completing, setCompleting] = useState(false)

  useEffect(() => {
    let cancelled = false
    setData(null)
    setError(null)
    getSubmodule(moduleSlug, submoduleSlug)
      .then((v) => {
        if (!cancelled) setData(v)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Ошибка')
      })
    return () => {
      cancelled = true
    }
  }, [moduleSlug, submoduleSlug])

  async function onComplete() {
    if (!data) return
    setCompleting(true)
    try {
      await completeSubmodule(data.submodule.id)
      const next = await getSubmodule(moduleSlug, submoduleSlug)
      setData(next)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось завершить')
    } finally {
      setCompleting(false)
    }
  }

  return (
    <div className="relative min-h-screen">
      <ElectricField className="opacity-50" />
      <div className="relative z-10 mx-auto grid min-h-screen max-w-7xl grid-cols-1 lg:grid-cols-[280px_1fr]">
        <aside className="border-r border-line/80 bg-surface/85 p-5 backdrop-blur-md lg:sticky lg:top-0 lg:h-screen lg:overflow-y-auto">
          <Link to="/" className="brand-link font-display text-lg font-bold">
            <span className="brand-glow">Roadmap</span>
          </Link>
          <div className="mt-6">
            {data ? (
              <MenuTree items={data.menu} moduleSlug={moduleSlug} activeSubSlug={submoduleSlug} />
            ) : (
              <p className="text-sm text-muted">Меню…</p>
            )}
          </div>
        </aside>

        <section className="px-5 py-8 sm:px-8">
          <Link to={`/modules/${moduleSlug}`} className="text-link">
            ← к подмодулям
          </Link>

          {error ? <p className="mt-6 text-danger">{error}</p> : null}
          {!data && !error ? <p className="mt-6 text-muted">Загрузка…</p> : null}

          {data ? (
            <article className="mx-auto mt-4 max-w-3xl">
              <p className="text-xs uppercase tracking-[0.18em] text-electric-deep">
                {data.module.title}
              </p>
              <h1 className="mt-2 font-display text-4xl font-extrabold">{data.submodule.title}</h1>
              <div className="mt-8 rounded-3xl border border-line bg-surface/90 p-6 sm:p-8">
                <ContentBlocks blocks={data.blocks || []} />
              </div>

              <div className="mt-8 flex flex-wrap items-center gap-3">
                {data.submodule.completed ? (
                  <span className="rounded-full bg-ok/15 px-4 py-2 text-sm font-medium text-ok">
                    Уже завершено
                  </span>
                ) : (
                  <button
                    type="button"
                    disabled={completing}
                    onClick={() => void onComplete()}
                    className="btn btn-primary btn-md"
                  >
                    {completing ? 'Сохраняем…' : 'Завершить'}
                  </button>
                )}
              </div>
            </article>
          ) : null}
        </section>
      </div>
    </div>
  )
}
