import { FormEvent, useCallback, useEffect, useState } from 'react'
import {
  createModule,
  createSprint,
  createSubmodule,
  deleteModule,
  deleteSprint,
  deleteSubmodule,
  getAdminTree,
  getSubmoduleContent,
  putSubmoduleContent,
  updateModule,
  updateSprint,
  updateSubmodule,
  uploadSubmoduleMarkdown,
} from '@/api/client'
import type { AdminTree } from '@/api/types'
import { AppShell } from '@/components/AppShell'
import { ElectricField } from '@/components/ElectricField'

export function AdminRoadmapPage() {
  const [tree, setTree] = useState<AdminTree | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  const [sprintTitle, setSprintTitle] = useState('')

  const [moduleSprintId, setModuleSprintId] = useState('')
  const [moduleTitle, setModuleTitle] = useState('')

  const [subModuleId, setSubModuleId] = useState('')
  const [subTitle, setSubTitle] = useState('')
  const [subBody, setSubBody] = useState('')

  const [editSubId, setEditSubId] = useState<string | null>(null)
  const [editBody, setEditBody] = useState('')

  const reload = useCallback(async () => {
    const data = await getAdminTree()
    setTree(data)
    if (!moduleSprintId && data.sprints[0]) setModuleSprintId(data.sprints[0].id)
    const firstMod = data.sprints[0]?.modules[0]
    if (!subModuleId && firstMod) setSubModuleId(firstMod.id)
  }, [moduleSprintId, subModuleId])

  useEffect(() => {
    reload().catch((err: unknown) => setError(err instanceof Error ? err.message : 'Ошибка'))
  }, [reload])

  async function run(fn: () => Promise<unknown>) {
    setBusy(true)
    setError(null)
    try {
      await fn()
      await reload()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка')
    } finally {
      setBusy(false)
    }
  }

  async function onCreateSprint(e: FormEvent) {
    e.preventDefault()
    await run(async () => {
      await createSprint({
        title: sprintTitle,
        is_published: true,
        description: '',
      })
      setSprintTitle('')
    })
  }

  async function onCreateModule(e: FormEvent) {
    e.preventDefault()
    await run(async () => {
      await createModule({
        sprint_id: moduleSprintId,
        title: moduleTitle,
        is_published: true,
        description: '',
      })
      setModuleTitle('')
    })
  }

  async function onCreateSubmodule(e: FormEvent) {
    e.preventDefault()
    await run(async () => {
      await createSubmodule({
        module_id: subModuleId,
        title: subTitle,
        is_published: true,
        body_md: subBody,
      })
      setSubTitle('')
      setSubBody('')
    })
  }

  async function openEditor(subId: string) {
    setEditSubId(subId)
    const c = await getSubmoduleContent(subId)
    setEditBody(c.body_md || '')
  }

  async function saveEditor() {
    if (!editSubId) return
    await run(async () => {
      await putSubmoduleContent(editSubId, editBody)
    })
  }

  async function onUpload(subId: string, file: File | null) {
    if (!file) return
    await run(async () => {
      await uploadSubmoduleMarkdown(subId, file)
      if (editSubId === subId) {
        const c = await getSubmoduleContent(subId)
        setEditBody(c.body_md || '')
      }
    })
  }

  const allModules =
    tree?.sprints.flatMap((s) => s.modules.map((m) => ({ ...m, sprintTitle: s.title }))) ?? []

  return (
    <div className="relative min-h-screen">
      <ElectricField />
      <AppShell title="Наполнение">
        <h1 className="font-display text-4xl font-extrabold">Роадмап</h1>
        <p className="mt-2 text-muted">
          Создавай спринты → модули → подмодули, заливай MD и публикуй.
        </p>
        {error ? <p className="mt-4 text-danger">{error}</p> : null}

        <div className="mt-8 grid gap-6 lg:grid-cols-3">
          <form
            onSubmit={(e) => void onCreateSprint(e)}
            className="space-y-3 rounded-2xl border border-electric/25 bg-surface/90 p-5"
          >
            <h2 className="font-display text-xl font-bold">+ Спринт</h2>
            <input
              className="w-full rounded-xl border border-line bg-bg px-3 py-2"
              placeholder="Название"
              value={sprintTitle}
              onChange={(e) => setSprintTitle(e.target.value)}
              required
            />
            <p className="text-xs text-muted">Новый спринт добавится в конец списка.</p>
            <button disabled={busy} className="btn btn-primary btn-md w-full">
              Создать спринт
            </button>
          </form>

          <form
            onSubmit={(e) => void onCreateModule(e)}
            className="space-y-3 rounded-2xl border border-electric/25 bg-surface/90 p-5"
          >
            <h2 className="font-display text-xl font-bold">+ Модуль</h2>
            <select
              className="w-full rounded-xl border border-line bg-bg px-3 py-2"
              value={moduleSprintId}
              onChange={(e) => setModuleSprintId(e.target.value)}
              required
            >
              <option value="">Спринт…</option>
              {tree?.sprints.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.title}
                </option>
              ))}
            </select>
            <input
              className="w-full rounded-xl border border-line bg-bg px-3 py-2"
              placeholder="Название"
              value={moduleTitle}
              onChange={(e) => setModuleTitle(e.target.value)}
              required
            />
            <button disabled={busy || !moduleSprintId} className="btn btn-primary btn-md w-full">
              Создать модуль
            </button>
          </form>

          <form
            onSubmit={(e) => void onCreateSubmodule(e)}
            className="space-y-3 rounded-2xl border border-electric/25 bg-surface/90 p-5"
          >
            <h2 className="font-display text-xl font-bold">+ Подмодуль</h2>
            <select
              className="w-full rounded-xl border border-line bg-bg px-3 py-2"
              value={subModuleId}
              onChange={(e) => setSubModuleId(e.target.value)}
              required
            >
              <option value="">Модуль…</option>
              {allModules.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.sprintTitle} / {m.title}
                </option>
              ))}
            </select>
            <input
              className="w-full rounded-xl border border-line bg-bg px-3 py-2"
              placeholder="Название"
              value={subTitle}
              onChange={(e) => setSubTitle(e.target.value)}
              required
            />
            <textarea
              className="h-28 w-full rounded-xl border border-line bg-bg px-3 py-2 font-mono text-sm"
              placeholder="# Markdown контент…"
              value={subBody}
              onChange={(e) => setSubBody(e.target.value)}
            />
            <button disabled={busy || !subModuleId} className="btn btn-primary btn-md w-full">
              Создать подмодуль
            </button>
          </form>
        </div>

        <section className="mt-10 space-y-6">
          <h2 className="font-display text-2xl font-bold">Дерево</h2>
          {!tree ? <p className="text-muted">Загрузка…</p> : null}
          {tree?.sprints.map((s) => (
            <div key={s.id} className="rounded-2xl border border-line bg-surface/90 p-5">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div className="text-xs uppercase tracking-wide text-electric-deep">спринт</div>
                  <div className="font-display text-xl font-bold">
                    {s.title}{' '}
                    <span className="text-sm font-normal text-muted">
                      #{s.position} · {s.is_published ? 'pub' : 'draft'} · {s.slug}
                    </span>
                  </div>
                </div>
                <div className="flex gap-2">
                  <button
                    type="button"
                    className="btn btn-ghost btn-sm"
                    onClick={() =>
                      void run(() =>
                        updateSprint(s.id, {
                          title: s.title,
                          slug: s.slug,
                          description: s.description,
                          position: s.position,
                          is_published: !s.is_published,
                        }),
                      )
                    }
                  >
                    {s.is_published ? 'Снять с публикации' : 'Опубликовать'}
                  </button>
                  <button
                    type="button"
                    className="btn btn-danger btn-sm"
                    onClick={() => {
                      if (confirm(`Удалить спринт «${s.title}»?`)) {
                        void run(() => deleteSprint(s.id))
                      }
                    }}
                  >
                    Удалить
                  </button>
                </div>
              </div>

              <div className="mt-4 space-y-4 pl-2">
                {s.modules.map((m) => (
                  <div key={m.id} className="rounded-xl border border-line/80 bg-bg/60 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <div className="font-medium">
                        {m.title}{' '}
                        <span className="text-xs text-muted">
                          #{m.position} · {m.is_published ? 'pub' : 'draft'}
                        </span>
                      </div>
                      <div className="flex gap-2">
                        <button
                          type="button"
                          className="btn btn-ghost btn-sm"
                          onClick={() =>
                            void run(() =>
                              updateModule(m.id, {
                                sprint_id: m.sprint_id,
                                title: m.title,
                                slug: m.slug,
                                description: m.description,
                                position: m.position,
                                is_published: !m.is_published,
                              }),
                            )
                          }
                        >
                          {m.is_published ? 'draft' : 'publish'}
                        </button>
                        <button
                          type="button"
                          className="btn btn-danger btn-sm"
                          onClick={() => {
                            if (confirm(`Удалить модуль «${m.title}»?`)) {
                              void run(() => deleteModule(m.id))
                            }
                          }}
                        >
                          Удалить
                        </button>
                      </div>
                    </div>

                    <ul className="mt-3 space-y-2">
                      {m.submodules.map((sm) => (
                        <li
                          key={sm.id}
                          className="flex flex-wrap items-center justify-between gap-2 rounded-lg bg-surface px-3 py-2 text-sm"
                        >
                          <span>
                            {sm.title}{' '}
                            <span className="text-xs text-muted">
                              #{sm.position} · {sm.is_published ? 'pub' : 'draft'}
                            </span>
                          </span>
                          <div className="flex flex-wrap gap-2">
                            <button
                              type="button"
                              className="btn btn-secondary btn-sm"
                              onClick={() => void openEditor(sm.id)}
                            >
                              MD
                            </button>
                            <label className="btn btn-ghost btn-sm">
                              .md файл
                              <input
                                type="file"
                                accept=".md,text/markdown,text/plain"
                                className="hidden"
                                onChange={(e) =>
                                  void onUpload(sm.id, e.target.files?.[0] ?? null)
                                }
                              />
                            </label>
                            <button
                              type="button"
                              className="btn btn-ghost btn-sm"
                              onClick={() =>
                                void run(async () => {
                                  const c = await getSubmoduleContent(sm.id)
                                  await updateSubmodule(sm.id, {
                                    module_id: sm.module_id,
                                    title: sm.title,
                                    slug: sm.slug,
                                    position: sm.position,
                                    is_published: !sm.is_published,
                                    body_md: c.body_md,
                                  })
                                })
                              }
                            >
                              {sm.is_published ? 'draft' : 'publish'}
                            </button>
                            <button
                              type="button"
                              className="btn btn-danger btn-sm"
                              onClick={() => {
                                if (confirm(`Удалить «${sm.title}»?`)) {
                                  void run(() => deleteSubmodule(sm.id))
                                }
                              }}
                            >
                              Удалить
                            </button>
                          </div>
                        </li>
                      ))}
                    </ul>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </section>

        {editSubId ? (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-ink/40 p-4 backdrop-blur-sm">
            <div className="flex max-h-[90vh] w-full max-w-3xl flex-col rounded-2xl border border-electric/30 bg-surface p-5">
              <h3 className="font-display text-xl font-bold">Редактор Markdown</h3>
              <textarea
                className="mt-3 min-h-[50vh] flex-1 rounded-xl border border-line bg-bg p-3 font-mono text-sm"
                value={editBody}
                onChange={(e) => setEditBody(e.target.value)}
              />
              <div className="mt-4 flex gap-2">
                <button
                  type="button"
                  disabled={busy}
                  onClick={() => void saveEditor()}
                  className="btn btn-primary btn-md"
                >
                  Сохранить
                </button>
                <button
                  type="button"
                  onClick={() => setEditSubId(null)}
                  className="btn btn-ghost btn-md"
                >
                  Закрыть
                </button>
              </div>
            </div>
          </div>
        ) : null}
      </AppShell>
    </div>
  )
}
