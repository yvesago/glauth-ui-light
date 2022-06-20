package config

import "time"

// config file.
type Backend struct {
	BaseDN        string
	Datastore     string
	Insecure      bool     // For LDAP and owncloud backend only
	Servers       []string // For LDAP and owncloud backend only
	NameFormat    string
	GroupFormat   string
	SSHKeyAttr    string
	UseGraphAPI   bool   // For ownCloud backend only
	Plugin        string // Path to plugin library, for plugin backend only
	PluginHandler string // Name of plugin's main handler function
	Database      string // For Database backends only
	AnonymousDSE  bool   // For Config and Database backends only
}
type Helper struct {
	Enabled       bool
	BaseDN        string
	Datastore     string
	Plugin        string // Path to plugin library, for plugin backend only
	PluginHandler string // Name of plugin's main handler function
	Database      string // For MySQL backend only TODO REname to match plugin
}
type Frontend struct {
	AllowedBaseDNs []string // For LDAP backend only
	Listen         string
	Cert           string
	Key            string
	TLS            bool
}
type LDAP struct {
	Enabled bool
	Listen  string
}
type LDAPS struct {
	Enabled bool
	Listen  string
	Cert    string
	Key     string
}
type API struct {
	Cert        string
	Enabled     bool
	Internals   bool
	Key         string
	Listen      string
	SecretToken string
	TLS         bool
}
type Behaviors struct {
	IgnoreCapabilities    bool
	LimitFailedBinds      bool
	NumberOfFailedBinds   int
	PeriodOfFailedBinds   time.Duration
	BlockFailedBindsFor   time.Duration
	PruneSourceTableEvery time.Duration
	PruneSourcesOlderThan time.Duration
}
type Capability struct {
	Action string `toml:"action,omitempty"`
	Object string `toml:"object,omitempty"`
}
type User struct {
	Name          	string       `toml:"name,omitempty"`
	OtherGroups   	[]int        `toml:"othergroups,omitempty"`
	PassSHA256    	string       `toml:"passsha256,omitempty"`
	PassBcrypt    	string       `toml:"passbcrypt,omitempty"`
	PassAppSHA256 	[]string     `toml:"passappsha256,omitempty"`
	PassAppBcrypt 	[]string     `toml:"passappbcrypt,omitempty"`
	PrimaryGroup  	int          `toml:"primarygroup,omitempty"`
	Capabilities  	[]Capability `toml:"capabilities,omitempty"`
	SSHKeys       	[]string     `toml:"sshkeys,omitempty"`
	OTPSecret     	string       `toml:"otpsecret,omitempty"`
	Yubikey       	string       `toml:"yubikey,omitempty"`
	Disabled      	bool         `toml:"disabled,omitempty"`
	UIDNumber   	int                    `toml:"uidnumber,omitempty"`
	Mail        	string                 `toml:"mail,omitempty"`
	LoginShell  	string                 `toml:"loginShell,omitempty"`
	GivenName   	string                 `toml:"givenname,omitempty"`
	Telephone	string                 `toml:"telephonenumber,omitempty"`
	Mobile      	string                 `toml:"mobile,omitempty"`
	SN          	string                 `toml:"sn,omitempty"`
	Homedir     	string                 `toml:"homeDir,omitempty"`
	CustomAttrs 	map[string]interface{} `toml:"customattrs,omitempty"`
}
type Group struct {
	Name string `toml:"name,omitempty"`
	//	UnixID        int    `toml:"unixid,omitempty"` // TODO: remove after deprecating UnixID on User and Group
	GIDNumber     int   `toml:"gidnumber,omitempty"`
	IncludeGroups []int `toml:"includegroups,omitempty"`
}
type Config struct {
	API                API
	Backend            Backend // Deprecated
	Backends           []Backend
	Helper             Helper
	Behaviors          Behaviors
	Debug              bool
	WatchConfig        bool
	YubikeyClientID    string
	YubikeySecret      string
	Frontend           Frontend
	LDAP               LDAP
	LDAPS              LDAPS
	Groups             []Group
	Syslog             bool
	Users              []User
	ConfigFile         string
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	AwsRegion          string
}
