import type { ComponentPropsWithoutRef, ReactNode } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import hljs from 'highlight.js/lib/core'
import go from 'highlight.js/lib/languages/go'
import bash from 'highlight.js/lib/languages/bash'
import json from 'highlight.js/lib/languages/json'
import sql from 'highlight.js/lib/languages/sql'
import yaml from 'highlight.js/lib/languages/yaml'
import plaintext from 'highlight.js/lib/languages/plaintext'
import 'highlight.js/styles/atom-one-dark.min.css'
import type { ContentBlock } from '@/api/types'
import { AuthImage } from '@/components/AuthImage'

/** Схема URL для картинок: ![alt](roadmap-image:UUID) */
export const ROADMAP_IMAGE_SCHEME = 'roadmap-image:'

/** Маркер вставки блока в MD — копируй и вставляй Ctrl+V. */
export function blockRefMarker(blockId: string): string {
  return `{{block:${blockId}}}`
}

const BLOCK_REF_RE = /\{\{block:([0-9a-fA-F-]{36})\}\}/g

hljs.registerLanguage('go', go)
hljs.registerLanguage('golang', go)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('shell', bash)
hljs.registerLanguage('sh', bash)
hljs.registerLanguage('json', json)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('yml', yaml)
hljs.registerLanguage('plaintext', plaintext)
hljs.registerLanguage('text', plaintext)

function highlightCode(code: string, language?: string): string {
  const lang = (language || 'go').toLowerCase()
  try {
    if (hljs.getLanguage(lang)) {
      return hljs.highlight(code, { language: lang, ignoreIllegals: true }).value
    }
  } catch {
    // fall through
  }
  return hljs.highlight(code, { language: 'go', ignoreIllegals: true }).value
}

function MarkdownCode({
  className,
  children,
  ...props
}: ComponentPropsWithoutRef<'code'>) {
  const text = String(children).replace(/\n$/, '')
  const match = /language-(\w+)/.exec(className || '')
  const isBlock = Boolean(match) || text.includes('\n')

  if (!isBlock) {
    return (
      <code className={className} {...props}>
        {children}
      </code>
    )
  }

  const html = highlightCode(text, match?.[1])

  return (
    <code
      className={`hljs language-${match?.[1] || 'go'}`}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}

function MarkdownImage({ src, alt }: ComponentPropsWithoutRef<'img'>) {
  if (src?.startsWith(ROADMAP_IMAGE_SCHEME)) {
    const imageId = src.slice(ROADMAP_IMAGE_SCHEME.length).trim()
    if (!imageId) return null
    return (
      <figure className="content-image">
        <AuthImage
          imageId={imageId}
          alt={alt || ''}
          className="max-h-[70vh] w-full max-w-full rounded-xl object-contain"
        />
        {alt ? (
          <figcaption className="mt-2 text-center text-sm text-muted">{alt}</figcaption>
        ) : null}
      </figure>
    )
  }
  return (
    <img
      src={src}
      alt={alt || ''}
      className="max-h-[70vh] w-full max-w-full rounded-xl object-contain"
    />
  )
}

/** Достаёт video id из youtube.com / youtu.be / shorts / embed / live. */
export function extractYoutubeId(url: string): string | null {
  try {
    const u = new URL(url.trim())
    const host = u.hostname.replace(/^www\./, '')
    if (host === 'youtu.be') {
      const id = u.pathname.split('/').filter(Boolean)[0]
      return id?.split('?')[0] || null
    }
    if (host === 'youtube.com' || host === 'm.youtube.com' || host === 'music.youtube.com') {
      if (u.pathname === '/watch' || u.pathname.startsWith('/watch')) {
        return u.searchParams.get('v')
      }
      const parts = u.pathname.split('/').filter(Boolean)
      if (parts[0] === 'embed' || parts[0] === 'shorts' || parts[0] === 'live' || parts[0] === 'v') {
        return parts[1] || null
      }
    }
  } catch {
    // ignore
  }
  return null
}

function YouTubeEmbed({ videoId }: { videoId: string }) {
  return (
    <div className="youtube-embed">
      <iframe
        src={`https://www.youtube-nocookie.com/embed/${encodeURIComponent(videoId)}`}
        title="YouTube video"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
        allowFullScreen
        loading="lazy"
        referrerPolicy="strict-origin-when-cross-origin"
      />
    </div>
  )
}

function MarkdownLink({ href, children, ...props }: ComponentPropsWithoutRef<'a'>) {
  const id = href ? extractYoutubeId(href) : null
  if (id) {
    return <YouTubeEmbed videoId={id} />
  }
  return (
    <a href={href} {...props}>
      {children}
    </a>
  )
}

type Segment =
  | { kind: 'md'; text: string }
  | { kind: 'answer'; text: string }
  | { kind: 'ref'; id: string }

/** Режет MD на текст, :::answer и {{block:id}}. */
export function splitMarkdownSegments(source: string): Segment[] {
  const answerRe = /^:::answer[ \t]*\r?\n([\s\S]*?)^:::[ \t]*$/gm
  const withAnswers: Segment[] = []
  let last = 0
  let match: RegExpExecArray | null
  while ((match = answerRe.exec(source)) !== null) {
    if (match.index > last) {
      withAnswers.push({ kind: 'md', text: source.slice(last, match.index) })
    }
    withAnswers.push({ kind: 'answer', text: match[1].replace(/^\n+|\n+$/g, '') })
    last = match.index + match[0].length
  }
  if (last < source.length) {
    withAnswers.push({ kind: 'md', text: source.slice(last) })
  }
  if (withAnswers.length === 0) {
    withAnswers.push({ kind: 'md', text: source })
  }

  const out: Segment[] = []
  for (const seg of withAnswers) {
    if (seg.kind !== 'md') {
      out.push(seg)
      continue
    }
    BLOCK_REF_RE.lastIndex = 0
    let mdLast = 0
    let refMatch: RegExpExecArray | null
    const text = seg.text
    while ((refMatch = BLOCK_REF_RE.exec(text)) !== null) {
      if (refMatch.index > mdLast) {
        out.push({ kind: 'md', text: text.slice(mdLast, refMatch.index) })
      }
      out.push({ kind: 'ref', id: refMatch[1] })
      mdLast = refMatch.index + refMatch[0].length
    }
    if (mdLast < text.length) {
      out.push({ kind: 'md', text: text.slice(mdLast) })
    }
  }
  return out
}

/** Все id блоков, на которые есть ссылка в markdown-полотнах. */
export function collectReferencedBlockIds(blocks: ContentBlock[]): Set<string> {
  const ids = new Set<string>()
  const re = /\{\{block:([0-9a-fA-F-]{36})\}\}/g
  for (const b of blocks) {
    if (b.type !== 'markdown' && b.type !== 'answer') continue
    let m: RegExpExecArray | null
    re.lastIndex = 0
    while ((m = re.exec(b.md)) !== null) {
      ids.add(m[1])
    }
  }
  return ids
}

function MarkdownChunk({ source }: { source: string }) {
  if (!source.trim()) return null
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        code: MarkdownCode,
        img: MarkdownImage,
        a: MarkdownLink,
      }}
    >
      {source}
    </ReactMarkdown>
  )
}

type BlockMap = Map<string, ContentBlock>

function RenderBlockEmbed({
  block,
  byId,
  stack,
}: {
  block: ContentBlock
  byId: BlockMap
  stack: Set<string>
}) {
  if (stack.has(block.id)) return null
  const nextStack = new Set(stack)
  nextStack.add(block.id)

  switch (block.type) {
    case 'markdown':
      return (
        <MarkdownBody source={block.md} byId={byId} stack={nextStack} emptyFallback={false} />
      )
    case 'answer':
      return (
        <details className="answer-block">
          <summary className="answer-summary">{block.title?.trim() || 'Ответ'}</summary>
          <div className="answer-body mt-3">
            <MarkdownBody
              source={block.md || '_Пусто_'}
              byId={byId}
              stack={nextStack}
              emptyFallback={false}
            />
          </div>
        </details>
      )
    case 'image':
      return (
        <figure className="content-image">
          <AuthImage
            imageId={block.image_id}
            alt={block.alt || ''}
            className="max-h-[70vh] w-full max-w-full rounded-xl object-contain"
          />
          {block.alt ? (
            <figcaption className="mt-2 text-center text-sm text-muted">{block.alt}</figcaption>
          ) : null}
        </figure>
      )
    default:
      return null
  }
}

function MarkdownBody({
  source,
  byId,
  stack,
  emptyFallback,
}: {
  source: string
  byId: BlockMap
  stack: Set<string>
  emptyFallback: boolean
}) {
  const text = source || (emptyFallback ? '_Контент пока пуст._' : '')
  if (!text) return null

  const segments = splitMarkdownSegments(text)
  const nodes: ReactNode[] = []

  segments.forEach((seg, i) => {
    if (seg.kind === 'md') {
      nodes.push(<MarkdownChunk key={`md-${i}`} source={seg.text} />)
      return
    }
    if (seg.kind === 'answer') {
      nodes.push(
        <details key={`answer-${i}`} className="answer-block">
          <summary className="answer-summary">Ответ</summary>
          <div className="answer-body mt-3">
            <MarkdownChunk source={seg.text || '_Пусто_'} />
          </div>
        </details>,
      )
      return
    }
    const ref = byId.get(seg.id)
    if (!ref) {
      nodes.push(
        <p key={`missing-${i}`} className="text-sm text-muted">
          Блок не найден
        </p>,
      )
      return
    }
    nodes.push(
      <div key={`ref-${seg.id}-${i}`}>
        <RenderBlockEmbed block={ref} byId={byId} stack={stack} />
      </div>,
    )
  })

  return <div className="md-body space-y-6">{nodes}</div>
}

/** Простой MD без резолва блоков (админ-превью текста и т.п.). */
export function Markdown({ source }: { source: string }) {
  return (
    <MarkdownBody
      source={source}
      byId={new Map()}
      stack={new Set()}
      emptyFallback
    />
  )
}

export function MarkdownWithBlocks({
  source,
  blocks,
}: {
  source: string
  blocks: ContentBlock[]
}) {
  const byId = new Map(blocks.map((b) => [b.id, b]))
  return (
    <MarkdownBody source={source} byId={byId} stack={new Set()} emptyFallback />
  )
}
