export interface ServerInfo {
  version: string
  bindPort: number
  vhostHTTPPort: number
  vhostHTTPSPort: number
  tcpmuxHTTPConnectPort: number
  kcpBindPort: number
  quicBindPort: number
  subdomainHost: string
  maxPoolCount: number
  maxPortsPerClient: number
  heartbeatTimeout: number
  allowPortsStr: string
  tlsForce: boolean

  // Stats
  totalTrafficIn: number
  totalTrafficOut: number
  curConns: number
  clientCounts: number
  proxyTypeCount: Record<string, number>
}

export interface HTTPPluginSettings {
  name: string
  addr: string
  path: string
  ops: string[]
  tlsVerify: boolean
}

export interface ServerSettings {
  configPath: string
  autoRestart: boolean
  bindAddr: string
  bindPort: number
  proxyBindAddr: string
  kcpBindPort: number
  quicBindPort: number
  vhostHTTPPort: number
  vhostHTTPSPort: number
  vhostHTTPTimeout: number
  tcpmuxHTTPConnectPort: number
  tcpmuxPassthrough: boolean
  subdomainHost: string
  authMethod: string
  authAdditionalScopes: string[]
  authToken: string
  authTokenSourceType: string
  authTokenSourceFile: string
  oidcIssuer: string
  oidcAudience: string
  oidcSkipExpiryCheck: boolean
  oidcSkipIssuerCheck: boolean
  tlsForce: boolean
  transportTLSCertFile: string
  transportTLSKeyFile: string
  transportTLSTrustedCaFile: string
  tcpMux: boolean
  tcpMuxKeepaliveInterval: number
  maxPoolCount: number
  heartbeatTimeout: number
  tcpKeepAlive: number
  quicKeepalivePeriod: number
  quicMaxIdleTimeout: number
  quicMaxIncomingStreams: number
  maxPortsPerClient: number
  userConnTimeout: number
  udpPacketSize: number
  natHoleAnalysisDataReserveHours: number
  detailedErrorsToClient: boolean
  allowPorts: string
  enablePrometheus: boolean
  dashboardAddr: string
  dashboardPort: number
  dashboardUser: string
  dashboardPassword: string
  dashboardAssetsDir: string
  dashboardPprofEnable: boolean
  dashboardTLSCertFile: string
  dashboardTLSKeyFile: string
  dashboardTLSTrustedCaFile: string
  logTo: string
  logLevel: string
  logMaxDays: number
  logDisablePrintColor: boolean
  sshTunnelGatewayPort: number
  sshPrivateKeyFile: string
  sshAutoGenKeyPath: string
  sshAuthorizedKeysFile: string
  httpPlugins: HTTPPluginSettings[]
  custom404Page: string
}
