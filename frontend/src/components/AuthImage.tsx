import { useEffect, useState } from 'react'
import { fetchContentImageBlob } from '@/api/client'

/** Картинка контента: грузится с Authorization, т.к. &lt;img src&gt; не шлёт Bearer. */
export function AuthImage({
  imageId,
  alt = '',
  className,
}: {
  imageId: string
  alt?: string
  className?: string
}) {
  const [url, setUrl] = useState<string | null>(null)
  const [error, setError] = useState(false)

  useEffect(() => {
    let cancelled = false
    let objectUrl: string | null = null
    setUrl(null)
    setError(false)
    fetchContentImageBlob(imageId)
      .then((u) => {
        if (cancelled) {
          URL.revokeObjectURL(u)
          return
        }
        objectUrl = u
        setUrl(u)
      })
      .catch(() => {
        if (!cancelled) setError(true)
      })
    return () => {
      cancelled = true
      if (objectUrl) URL.revokeObjectURL(objectUrl)
    }
  }, [imageId])

  if (error) {
    return <p className="text-sm text-muted">Не удалось загрузить изображение</p>
  }
  if (!url) {
    return <p className="text-sm text-muted">Загрузка изображения…</p>
  }
  return <img src={url} alt={alt} className={className} />
}
