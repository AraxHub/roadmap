import type { ComponentPropsWithoutRef } from 'react'
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

export function Markdown({ source }: { source: string }) {
  return (
    <div className="md-body">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          code: MarkdownCode,
        }}
      >
        {source || '_Контент пока пуст._'}
      </ReactMarkdown>
    </div>
  )
}
