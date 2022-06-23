package config

type Ctmp struct {
	Users  []User  `toml:"users,omitempty"`
	Groups []Group `toml:"groups,omitempty"`
}

type WebConfig struct {
	AppName    string
	AppDesc    string
	MaskOTP    bool
	Port       string
	DBfile     string
	Debug      bool
	Tests      bool
	Sec        Sec
	SSL        SSL
	Logs       Logs
	Locale     Locale
	PassPolicy PassPolicy
	CfgUsers   CfgUsers
	CfgGroups  CfgGroups
}

type Sec struct {
	CSRFrandom     string
	TrustedProxies []string
}

type SSL struct {
	Crt string
	Key string
}

type Logs struct {
	Path          string
	RotationCount uint
}

type Locale struct {
	Lang  string
	Path  string
	Langs []string
}

type PassPolicy struct {
	Min              int
	Max              int
	AllowReadSSHA256 bool
	Entropy          int64
}

type CfgUsers struct {
	Start         int
	GIDAdmin      int
	GIDcanChgPass int
	GIDuseOtp     int
}

type CfgGroups struct {
	Start int
}
