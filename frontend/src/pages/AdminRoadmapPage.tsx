import { FormEvent, useCallback, useEffect, useState } from 'react'
import {
  createModule,
  createSprint,
  createSubmodule,
  deleteContentImage,
  deleteModule,
  deleteSprint,
  deleteSubmodule,
  getAdminTree,
  getSubmoduleContent,
  putSubmoduleContent,
  updateModule,
  updateSprint,
  updateSubmodule,
  uploadSubmoduleImage,
  uploadSubmoduleMarkdown,
} from '@/api/client'
import type { AdminTree, ContentBlock } from '@/api/types'
import { AppShell } from '@/components/AppShell'
import { AuthImage } from '@/components/AuthImage'
import { ElectricField } from '@/components/ElectricField'
import { blockRefMarker, collectReferencedBlockIds } from '@/components/Markdown'

function newBlockId() {
  return crypto.randomUUID()
}

function insertAfter(
  blocks: ContentBlock[],
  selectedId: string | null,
  block: ContentBlock,
): ContentBlock[] {
  if (!selectedId) return [...blocks, block]
  const idx = blocks.findIndex((b) => b.id === selectedId)
  if (idx < 0) return [...blocks, block]
  const next = [...blocks]
  next.splice(idx + 1, 0, block)
  return next
}

export function AdminRoadmapPage() {
  const [tree, setTree] = useState<AdminTree | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)

  const [sprintTitle, setSprintTitle] = useState('')

  const [moduleSprintId, setModuleSprintId] = useState('')
  const [moduleTitle, setModuleTitle] = useState('')

  const [subModuleId, setSubModuleId] = useState('')
  const [subTitle, setSubTitle] = useState('')
  const [subBody, setSubBody] = useState('')

  const [editSubId, setEditSubId] = useState<string | null>(null)
  const [editBlocks, setEditBlocks] = useState<ContentBlock[]>([])
  const [selectedBlockId, setSelectedBlockId] = useState<string | null>(null)

  const referenced = collectReferencedBlockIds(editBlocks)

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
    const blocks = c.blocks || []
    setEditBlocks(blocks)
    setSelectedBlockId(blocks[0]?.id ?? null)
    setCopiedId(null)
  }

  async function saveEditor() {
    if (!editSubId) return
    await run(async () => {
      await putSubmoduleContent(editSubId, editBlocks)
    })
  }

  async function onUpload(subId: string, file: File | null) {
    if (!file) return
    await run(async () => {
      const c = await uploadSubmoduleMarkdown(subId, file)
      if (editSubId === subId) {
        setEditBlocks(c.blocks || [])
        setSelectedBlockId(c.blocks?.[0]?.id ?? null)
      }
    })
  }

  function addMarkdownBlock() {
    const block: ContentBlock = { id: newBlockId(), type: 'markdown', md: '' }
    setEditBlocks((prev) => insertAfter(prev, selectedBlockId, block))
    setSelectedBlockId(block.id)
  }

  function addAnswerBlock() {
    const block: ContentBlock = { id: newBlockId(), type: 'answer', md: '', title: 'Ответ' }
    setEditBlocks((prev) => insertAfter(prev, selectedBlockId, block))
    setSelectedBlockId(block.id)
  }

  async function addImageBlock(file: File | null) {
    if (!file || !editSubId) return
    setBusy(true)
    setError(null)
    try {
      const uploaded = await uploadSubmoduleImage(editSubId, file)
      const block: ContentBlock = {
        id: newBlockId(),
        type: 'image',
        image_id: uploaded.id,
        alt: '',
      }
      setEditBlocks((prev) => insertAfter(prev, selectedBlockId, block))
      setSelectedBlockId(block.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки картинки')
    } finally {
      setBusy(false)
    }
  }

  async function copyBlockMarker(blockId: string) {
    const marker = blockRefMarker(blockId)
    try {
      await navigator.clipboard.writeText(marker)
      setCopiedId(blockId)
      window.setTimeout(() => setCopiedId((id) => (id === blockId ? null : id)), 2000)
    } catch {
      setError('Не удалось скопировать — скопируй маркер вручную: ' + marker)
    }
  }

  function moveBlock(id: string, dir: -1 | 1) {
    setEditBlocks((prev) => {
      const idx = prev.findIndex((b) => b.id === id)
      const nextIdx = idx + dir
      if (idx < 0 || nextIdx < 0 || nextIdx >= prev.length) return prev
      const next = [...prev]
      const [item] = next.splice(idx, 1)
      next.splice(nextIdx, 0, item)
      return next
    })
  }

  async function removeBlock(id: string) {
    const block = editBlocks.find((b) => b.id === id)
    if (!block) return
    if (block.type === 'image' && block.image_id) {
      try {
        await deleteContentImage(block.image_id)
      } catch {
        // блок всё равно уберём из редактора
      }
    }
    setEditBlocks((prev) => prev.filter((b) => b.id !== id))
    if (selectedBlockId === id) setSelectedBlockId(null)
  }

  function updateBlockMD(id: string, md: string) {
    setEditBlocks((prev) =>
      prev.map((b) => (b.id === id && (b.type === 'markdown' || b.type === 'answer') ? { ...b, md } : b)),
    )
  }

  function updateAnswerTitle(id: string, title: string) {
    setEditBlocks((prev) =>
      prev.map((b) => (b.id === id && b.type === 'answer' ? { ...b, title } : b)),
    )
  }

  function updateBlockAlt(id: string, alt: string) {
    setEditBlocks((prev) =>
      prev.map((b) => (b.id === id && b.type === 'image' ? { ...b, alt } : b)),
    )
  }

  const allModules =
    tree?.sprints.flatMap((s) => s.modules.map((m) => ({ ...m, sprintTitle: s.title }))) ?? []

  return (
    <div className="relative min-h-screen">
      <ElectricField />
      <AppShell title="Наполнение">
        <h1 className="font-display text-4xl font-extrabold">Роадмап</h1>
        <p className="mt-2 text-muted">
          Создавай спринты → модули → подмодули, редактируй контент блоками и публикуй.
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
              placeholder="# Начальный Markdown (опционально)…"
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
                              Контент
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
                                  await updateSubmodule(sm.id, {
                                    module_id: sm.module_id,
                                    title: sm.title,
                                    slug: sm.slug,
                                    position: sm.position,
                                    is_published: !sm.is_published,
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
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-ink/40 p-3 backdrop-blur-sm sm:p-6">
            <div className="flex h-[80vh] w-[80vw] max-w-none flex-col rounded-2xl border border-electric/30 bg-surface p-5 sm:p-6">
              <h3 className="font-display text-xl font-bold">Редактор контента</h3>
              <p className="mt-1 shrink-0 text-sm text-muted">
                Ученику видны только блоки «Текст». Ответ и картинку: «Скопировать» → вставь Ctrl+V
                в нужное место MD. Пока маркер не вставлен — на уроке блок не показывается.
              </p>

              <div className="mt-3 flex shrink-0 flex-wrap gap-2">
                <button type="button" className="btn btn-secondary btn-sm" onClick={addMarkdownBlock}>
                  Добавить текст (MD)
                </button>
                <button type="button" className="btn btn-secondary btn-sm" onClick={addAnswerBlock}>
                  Добавить скрытый материал
                </button>
                <label className="btn btn-secondary btn-sm">
                  Добавить изображение
                  <input
                    type="file"
                    accept="image/png,image/jpeg,image/webp,image/gif"
                    className="hidden"
                    onChange={(e) => {
                      void addImageBlock(e.target.files?.[0] ?? null)
                      e.target.value = ''
                    }}
                  />
                </label>
              </div>

              <div className="mt-4 min-h-0 flex-1 space-y-3 overflow-y-auto pr-1">
                {editBlocks.length === 0 ? (
                  <p className="text-sm text-muted">Пока пусто — добавь блок кнопками выше.</p>
                ) : null}
                {editBlocks.map((block, index) => {
                  const selected = selectedBlockId === block.id
                  const isCanvas = block.type === 'markdown'
                  const isInserted = isCanvas || referenced.has(block.id)
                  return (
                    <div
                      key={block.id}
                      role="button"
                      tabIndex={0}
                      onClick={() => setSelectedBlockId(block.id)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') setSelectedBlockId(block.id)
                      }}
                      className={`rounded-xl border p-3 ${
                        selected ? 'border-electric/50 bg-electric-soft/40' : 'border-line bg-bg'
                      }`}
                    >
                      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
                        <span className="text-xs font-medium uppercase tracking-wide text-muted">
                          {block.type === 'markdown'
                            ? 'Текст'
                            : block.type === 'answer'
                              ? 'Скрытый материал'
                              : 'Изображение'}{' '}
                          · #{index + 1}
                          {!isCanvas ? (
                            <span
                              className={`ml-2 normal-case tracking-normal ${
                                isInserted ? 'text-ok' : 'text-danger'
                              }`}
                            >
                              {isInserted ? '· вставлен' : '· не вставлен (скрыт на уроке)'}
                            </span>
                          ) : null}
                        </span>
                        <div className="flex flex-wrap gap-1" onClick={(e) => e.stopPropagation()}>
                          <button
                            type="button"
                            className="btn btn-secondary btn-sm"
                            onClick={() => void copyBlockMarker(block.id)}
                          >
                            {copiedId === block.id ? 'Скопировано' : 'Скопировать'}
                          </button>
                          <button
                            type="button"
                            className="btn btn-ghost btn-sm"
                            disabled={index === 0}
                            onClick={() => moveBlock(block.id, -1)}
                          >
                            ↑
                          </button>
                          <button
                            type="button"
                            className="btn btn-ghost btn-sm"
                            disabled={index === editBlocks.length - 1}
                            onClick={() => moveBlock(block.id, 1)}
                          >
                            ↓
                          </button>
                          <button
                            type="button"
                            className="btn btn-danger btn-sm"
                            onClick={() => void removeBlock(block.id)}
                          >
                            Удалить
                          </button>
                        </div>
                      </div>

                      {block.type === 'markdown' || block.type === 'answer' ? (
                        <div className="space-y-2" onClick={(e) => e.stopPropagation()}>
                          {block.type === 'answer' ? (
                            <label className="block text-xs text-muted">
                              Название спойлера
                              <input
                                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm font-medium text-ink"
                                value={block.title ?? 'Ответ'}
                                placeholder="Ответ"
                                onChange={(e) => updateAnswerTitle(block.id, e.target.value)}
                              />
                            </label>
                          ) : null}
                          <textarea
                            className="min-h-[280px] w-full resize-y rounded-lg border border-line bg-surface p-3 font-mono text-sm leading-relaxed"
                            value={block.md}
                            placeholder={
                              block.type === 'answer'
                                ? 'Markdown ответа…'
                                : 'Markdown полотна… Вставляй сюда {{block:…}} через Ctrl+V'
                            }
                            onChange={(e) => updateBlockMD(block.id, e.target.value)}
                          />
                        </div>
                      ) : (
                        <div className="space-y-2" onClick={(e) => e.stopPropagation()}>
                          <AuthImage
                            imageId={block.image_id}
                            alt={block.alt || ''}
                            className="max-h-48 w-full rounded-lg object-contain"
                          />
                          <input
                            className="w-full rounded-lg border border-line bg-surface px-2 py-1.5 text-sm"
                            placeholder="Подпись (alt)"
                            value={block.alt || ''}
                            onChange={(e) => updateBlockAlt(block.id, e.target.value)}
                          />
                        </div>
                      )}
                      {!isCanvas ? (
                        <p className="mt-2 font-mono text-xs text-muted">{blockRefMarker(block.id)}</p>
                      ) : null}
                    </div>
                  )
                })}
              </div>

              <div className="mt-4 flex shrink-0 gap-2">
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
