// Minimal client for the backend JSON envelope (backend/docs/responses.md).
// Relative /api paths go through the Vite dev proxy.

type Envelope<T> = { success: true; data?: T } | { success: false; error: string }

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string) {
    super(code)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

// Network failures (backend/proxy unreachable) reject with the fetch TypeError, not ApiError.
export async function apiPost<T>(
  path: string,
  body?: unknown,
  extraHeaders?: Record<string, string>,
): Promise<T> {
  const headers: Record<string, string> = { ...extraHeaders }
  if (body !== undefined) headers['Content-Type'] = 'application/json'

  const response = await fetch(path, {
    method: 'POST',
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  })

  let payload: Envelope<T> | null = null
  try {
    payload = (await response.json()) as Envelope<T>
  } catch {
    // Non-JSON body (e.g. proxy error page); handled below.
  }

  if (!response.ok || !payload || !payload.success) {
    const code = payload && !payload.success ? payload.error : 'INTERNAL_ERROR'
    throw new ApiError(response.status, code)
  }
  return payload.data as T
}
