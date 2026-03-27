import type { DecodePairResult, PairSharePayload } from '../types/pairing'

/** Generate a random hex secret key (16 chars = 8 bytes) */
export function generateSecretKey(): string {
  const array = new Uint8Array(8)
  crypto.getRandomValues(array)
  return Array.from(array)
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

/** Encode a PairSharePayload into a URL-safe Base64 share code */
export function encodePairConfig(payload: PairSharePayload): string {
  const json = JSON.stringify(payload)
  const bytes = new TextEncoder().encode(json)
  // btoa requires binary string
  let binary = ''
  bytes.forEach((b) => (binary += String.fromCharCode(b)))
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

/** Decode a share code back into a PairSharePayload */
export function decodePairConfig(code: string): DecodePairResult {
  try {
    // Restore URL-safe Base64 to standard Base64
    const base64 = code.replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64 + '=='.slice(0, (4 - (base64.length % 4)) % 4)
    const binary = atob(padded)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i)
    }
    const json = new TextDecoder().decode(bytes)
    const parsed = JSON.parse(json)

    // Validate required fields
    if (!parsed || typeof parsed !== 'object') {
      return { ok: false, error: 'Invalid payload format' }
    }
    if (parsed.v !== 1) {
      return { ok: false, error: `Unsupported share code version: ${parsed.v}` }
    }
    if (!parsed.type || !['stcp', 'xtcp', 'sudp'].includes(parsed.type)) {
      return { ok: false, error: 'Invalid proxy type in share code' }
    }
    if (!parsed.serverName || typeof parsed.serverName !== 'string') {
      return { ok: false, error: 'Missing serverName in share code' }
    }
    if (!parsed.secretKey || typeof parsed.secretKey !== 'string') {
      return { ok: false, error: 'Missing secretKey in share code' }
    }
    if (!parsed.serverAddr || typeof parsed.serverAddr !== 'string') {
      return { ok: false, error: 'Missing serverAddr in share code' }
    }

    return { ok: true, payload: parsed as PairSharePayload }
  } catch (_e) {
    return { ok: false, error: 'Failed to decode share code: invalid format' }
  }
}

/** Build the local access URL from bindAddr + bindPort + proxy type */
export function buildAccessUrl(
  bindAddr: string,
  bindPort: number,
  proxyType: string,
): string {
  const host = bindAddr || '127.0.0.1'
  if (proxyType === 'stcp' || proxyType === 'xtcp') {
    return `tcp://${host}:${bindPort}`
  }
  if (proxyType === 'sudp') {
    return `udp://${host}:${bindPort}`
  }
  return `${host}:${bindPort}`
}
