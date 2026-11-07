// Client-side storage of the backend JWT pair (backend/docs/auth.md).
// "Remember me" keeps the pair in localStorage (survives browser restarts, shared by tabs);
// otherwise sessionStorage is used, matching the backend's short refresh session.

const STORAGE_KEY = 'preditto.auth'

export interface StoredTokens {
  access_token: string
  refresh_token: string
  // Epoch milliseconds when the access token expires (client clock).
  access_expires_at: number
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  // Access token lifetime in seconds.
  expires_in: number
}

function storages(): Storage[] {
  const result: Storage[] = []
  try {
    result.push(window.localStorage)
  } catch {
    // Storage may be blocked (privacy settings); the user then stays signed out.
  }
  try {
    result.push(window.sessionStorage)
  } catch {
    // See above.
  }
  return result
}

function readFrom(storage: Storage): StoredTokens | null {
  try {
    const raw = storage.getItem(STORAGE_KEY)
    if (!raw) return null
    const value = JSON.parse(raw) as Partial<StoredTokens>
    if (
      typeof value.access_token === 'string' &&
      typeof value.refresh_token === 'string' &&
      typeof value.access_expires_at === 'number'
    ) {
      return value as StoredTokens
    }
  } catch {
    // Corrupted entry: treat as signed out.
  }
  return null
}

function findStorage(): { storage: Storage; tokens: StoredTokens } | null {
  for (const storage of storages()) {
    const tokens = readFrom(storage)
    if (tokens) return { storage, tokens }
  }
  return null
}

export function getTokens(): StoredTokens | null {
  return findStorage()?.tokens ?? null
}

export function hasTokens(): boolean {
  return getTokens() !== null
}

function toStored(pair: TokenPair): StoredTokens {
  return {
    access_token: pair.access_token,
    refresh_token: pair.refresh_token,
    access_expires_at: Date.now() + pair.expires_in * 1000,
  }
}

// Saves the pair from login: `remember` selects the persistent storage.
export function saveTokens(pair: TokenPair, remember: boolean): void {
  clearTokens()
  const storage = remember ? window.localStorage : window.sessionStorage
  try {
    storage.setItem(STORAGE_KEY, JSON.stringify(toStored(pair)))
  } catch {
    // Quota or blocked storage: the session lasts only until the next page load.
  }
}

// Replaces the pair after refresh rotation, keeping the storage chosen at login.
// Does nothing when the user signed out meanwhile.
export function replaceTokens(pair: TokenPair): void {
  const current = findStorage()
  if (!current) return
  try {
    current.storage.setItem(STORAGE_KEY, JSON.stringify(toStored(pair)))
  } catch {
    // See saveTokens.
  }
}

export function clearTokens(): void {
  for (const storage of storages()) {
    try {
      storage.removeItem(STORAGE_KEY)
    } catch {
      // Nothing to clear.
    }
  }
}
