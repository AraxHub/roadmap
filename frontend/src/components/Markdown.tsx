import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

export function Markdown({ source }: { source: string }) {
  return (
    <div className="md-body">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{source || '_Контент пока пуст._'}</ReactMarkdown>
    </div>
  )
}
