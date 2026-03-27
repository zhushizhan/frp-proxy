// STCP/XTCP pairing share-code data structures

export type PairingProxyType = 'stcp' | 'xtcp' | 'sudp'

/** Encoded in the share code (what A sends to B) */
export interface PairSharePayload {
  /** Schema version */
  v: 1
  /** Proxy type */
  type: PairingProxyType
  /** The proxy name on A side (B's visitor.serverName) */
  serverName: string
  /** Shared secret key */
  secretKey: string
  /** frps server address (so B can verify it matches) */
  serverAddr: string
  /** Optional: A's user prefix (for serverUser field on B side) */
  serverUser?: string
  /** Optional: display label shown to B */
  label?: string
}

/** Result of decoding a share code */
export type DecodePairResult =
  | { ok: true; payload: PairSharePayload }
  | { ok: false; error: string }
