import {
  useEffect,
  useId,
  useRef,
  useState,
  type DragEvent,
  type KeyboardEvent,
} from 'react'
import { CloseIcon, PlusIcon, UserIcon } from '../Icons/Icons.tsx'
import './AvatarUpload.css'

const ACCEPTED_TYPES = ['image/png', 'image/jpeg', 'image/webp']
const MAX_SIZE = 5 * 1024 * 1024

type AvatarUploadProps = {
  onChange: (file: File | null) => void
}

function validateFile(file: File) {
  if (!ACCEPTED_TYPES.includes(file.type)) {
    return 'Поддерживаются PNG, JPEG и WebP'
  }
  if (file.size > MAX_SIZE) {
    return 'Файл больше 5 МБ'
  }
  return ''
}

function AvatarUpload({ onChange }: AvatarUploadProps) {
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [isDragOver, setIsDragOver] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const tileRef = useRef<HTMLDivElement>(null)
  const previewUrlRef = useRef<string | null>(null)
  const errorId = useId()

  // Revoke the last preview URL on unmount.
  useEffect(() => () => {
    if (previewUrlRef.current) URL.revokeObjectURL(previewUrlRef.current)
  }, [])

  const replacePreview = (file: File | null) => {
    if (previewUrlRef.current) URL.revokeObjectURL(previewUrlRef.current)
    previewUrlRef.current = file ? URL.createObjectURL(file) : null
    setPreviewUrl(previewUrlRef.current)
    onChange(file)
  }

  const handleFile = (file: File | undefined) => {
    if (!file) return
    const message = validateFile(file)
    setError(message)
    if (!message) replacePreview(file)
  }

  const openPicker = () => inputRef.current?.click()

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      openPicker()
    }
  }

  const handleDragOver = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    setIsDragOver(true)
  }

  const handleDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    setIsDragOver(false)
    handleFile(event.dataTransfer.files[0])
  }

  const handleRemove = () => {
    replacePreview(null)
    setError('')
    tileRef.current?.focus()
  }

  const tileClassName = [
    'avatar-upload__tile',
    isDragOver && 'is-dragover',
    previewUrl && 'has-preview',
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <div className="avatar-upload">
      <div className="avatar-upload__frame">
        <div
          ref={tileRef}
          className={tileClassName}
          role="button"
          tabIndex={0}
          aria-label={previewUrl ? 'Заменить аватарку' : 'Загрузить аватарку'}
          aria-describedby={error ? errorId : undefined}
          onClick={openPicker}
          onKeyDown={handleKeyDown}
          onDragOver={handleDragOver}
          onDragLeave={() => setIsDragOver(false)}
          onDrop={handleDrop}
        >
          {previewUrl ? (
            <img className="avatar-upload__preview" src={previewUrl} alt="" />
          ) : (
            <>
              <span className="avatar-upload__circle" aria-hidden="true">
                <UserIcon />
                <span className="avatar-upload__badge">
                  <PlusIcon />
                </span>
              </span>
              <span className="avatar-upload__title">Загрузить аватарку</span>
              <span className="avatar-upload__hint">
                Перетащите файл сюда или <span>выберите</span>
              </span>
            </>
          )}
        </div>
        {previewUrl && (
          <button
            type="button"
            className="avatar-upload__remove"
            aria-label="Удалить аватар"
            onClick={handleRemove}
          >
            <CloseIcon />
          </button>
        )}
      </div>
      <input
        ref={inputRef}
        type="file"
        accept={ACCEPTED_TYPES.join(',')}
        hidden
        tabIndex={-1}
        onChange={(event) => {
          handleFile(event.target.files?.[0])
          // Allow re-selecting the same file after removal.
          event.target.value = ''
        }}
      />
      {error && (
        <p className="form-error" id={errorId} role="alert">
          {error}
        </p>
      )}
    </div>
  )
}

export default AvatarUpload
