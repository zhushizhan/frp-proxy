package model

type HTTPPluginSettings struct {
	Name      string   `json:"name"`
	Addr      string   `json:"addr"`
	Path      string   `json:"path"`
	Ops       []string `json:"ops"`
	TLSVerify bool     `json:"tlsVerify"`
}

type ServerSettings struct {
	ConfigPath             string               `json:"configPath,omitempty"`
	AutoRestart            bool                 `json:"autoRestart"`
	BindAddr               string               `json:"bindAddr"`
	BindPort               int                  `json:"bindPort"`
	ProxyBindAddr          string               `json:"proxyBindAddr"`
	KCPBindPort            int                  `json:"kcpBindPort"`
	QUICBindPort           int                  `json:"quicBindPort"`
	VhostHTTPPort          int                  `json:"vhostHTTPPort"`
	VhostHTTPSPort         int                  `json:"vhostHTTPSPort"`
	VhostHTTPTimeout       int64                `json:"vhostHTTPTimeout"`
	TCPMuxHTTPConnectPort  int                  `json:"tcpmuxHTTPConnectPort"`
	TCPMuxPassthrough      bool                 `json:"tcpmuxPassthrough"`
	SubdomainHost          string               `json:"subdomainHost"`
	AuthMethod             string               `json:"authMethod"`
	AuthAdditionalScopes   []string             `json:"authAdditionalScopes"`
	AuthToken              string               `json:"authToken"`
	AuthTokenSourceType    string               `json:"authTokenSourceType"`
	AuthTokenSourceFile    string               `json:"authTokenSourceFile"`
	OIDCIssuer             string               `json:"oidcIssuer"`
	OIDCAudience           string               `json:"oidcAudience"`
	OIDCSkipExpiryCheck    bool                 `json:"oidcSkipExpiryCheck"`
	OIDCSkipIssuerCheck    bool                 `json:"oidcSkipIssuerCheck"`
	TLSForce               bool                 `json:"tlsForce"`
	TransportTLSCertFile   string               `json:"transportTLSCertFile"`
	TransportTLSKeyFile    string               `json:"transportTLSKeyFile"`
	TransportTLSTrustedCA  string               `json:"transportTLSTrustedCaFile"`
	TCPMux                 bool                 `json:"tcpMux"`
	TCPMuxKeepalive        int64                `json:"tcpMuxKeepaliveInterval"`
	MaxPoolCount           int64                `json:"maxPoolCount"`
	HeartbeatTimeout       int64                `json:"heartbeatTimeout"`
	TCPKeepAlive           int64                `json:"tcpKeepAlive"`
	QUICKeepalivePeriod    int                  `json:"quicKeepalivePeriod"`
	QUICMaxIdleTimeout     int                  `json:"quicMaxIdleTimeout"`
	QUICMaxIncomingStreams int                  `json:"quicMaxIncomingStreams"`
	MaxPortsPerClient      int64                `json:"maxPortsPerClient"`
	UserConnTimeout        int64                `json:"userConnTimeout"`
	UDPPacketSize          int64                `json:"udpPacketSize"`
	NatHoleRetentionHours  int64                `json:"natHoleAnalysisDataReserveHours"`
	DetailedErrorsToClient bool                 `json:"detailedErrorsToClient"`
	AllowPorts             string               `json:"allowPorts"`
	EnablePrometheus       bool                 `json:"enablePrometheus"`
	DashboardAddr          string               `json:"dashboardAddr"`
	DashboardPort          int                  `json:"dashboardPort"`
	DashboardUser          string               `json:"dashboardUser"`
	DashboardPassword      string               `json:"dashboardPassword"`
	DashboardAssetsDir     string               `json:"dashboardAssetsDir"`
	DashboardPprofEnable   bool                 `json:"dashboardPprofEnable"`
	DashboardTLSCertFile   string               `json:"dashboardTLSCertFile"`
	DashboardTLSKeyFile    string               `json:"dashboardTLSKeyFile"`
	DashboardTLSTrustedCA  string               `json:"dashboardTLSTrustedCaFile"`
	LogTo                  string               `json:"logTo"`
	LogLevel               string               `json:"logLevel"`
	LogMaxDays             int64                `json:"logMaxDays"`
	LogDisablePrintColor   bool                 `json:"logDisablePrintColor"`
	SSHTunnelGatewayPort   int                  `json:"sshTunnelGatewayPort"`
	SSHPrivateKeyFile      string               `json:"sshPrivateKeyFile"`
	SSHAutoGenKeyPath      string               `json:"sshAutoGenKeyPath"`
	SSHAuthorizedKeysFile  string               `json:"sshAuthorizedKeysFile"`
	HTTPPlugins            []HTTPPluginSettings `json:"httpPlugins"`
	Custom404Page          string               `json:"custom404Page"`
}
