import type { ContentBlock } from '@/api/types'
import { MarkdownWithBlocks } from '@/components/Markdown'

/**
 * Ученику показываются только markdown-полотна.
 * Answer/image видны, только если вставлены маркером {{block:id}} в MD.
 */
export function ContentBlocks({ blocks }: { blocks: ContentBlock[] }) {
  const canvases = blocks.filter((b) => b.type === 'markdown')

  if (!canvases.length) {
    return <MarkdownWithBlocks source="" blocks={blocks} />
  }

  return (
    <div className="content-blocks space-y-6">
      {canvases.map((block) => (
        <div key={block.id}>
          <MarkdownWithBlocks source={block.md} blocks={blocks} />
        </div>
      ))}
    </div>
  )
}
