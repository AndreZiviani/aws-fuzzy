package ssm_plugin

import (
	"embed"
)

//go:embed assets/plugin/windows_amd64/session-manager-plugin.exe
var assets embed.FS
