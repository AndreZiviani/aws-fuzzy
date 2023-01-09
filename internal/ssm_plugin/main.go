package ssm_plugin

import (
	//	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/AndreZiviani/aws-fuzzy/internal/afconfig"
	wraperror "github.com/gjbae1212/go-wraperror"
)

/*
// we dont want to include ssm plugin from different architectures
// so assets is defined per arch/os
// more info: https://www.digitalocean.com/community/tutorials/building-go-applications-for-different-operating-systems-and-architectures#using-goos-and-goarch-filename-suffixes

var assets embed.FS
*/

// GetAsset returns asset file.
func GetAsset(filename string) ([]byte, error) {
	return assets.ReadFile("assets/" + filename)
}

// GetSsmPluginName returns filename for aws ssm plugin.
func GetSsmPluginName() string {
	if strings.ToLower(runtime.GOOS) == "windows" {
		return "session-manager-plugin.exe"
	} else {
		return "session-manager-plugin"
	}
}

// GetSsmPlugin returns filepath for aws ssm plugin.
func GetSsmPlugin() ([]byte, error) {
	return GetAsset(getSSMPluginKey())
}

func getSSMPluginKey() string {
	return fmt.Sprintf("plugin/%s_%s/%s",
		strings.ToLower(runtime.GOOS), strings.ToLower(runtime.GOARCH), GetSsmPluginName())
}

func ExtractAssets() (string, error) {
	plugin, err := GetSsmPlugin()
	if err != nil {
		return "", err
	}

	cfg := afconfig.NewDefaultConfig()
	configDir, _ := cfg.ConfigFolder()
	pluginPath := filepath.Join(configDir, GetSsmPluginName())
	info, err := os.Stat(pluginPath)

	if os.IsNotExist(err) {
		err := ioutil.WriteFile(pluginPath, plugin, 0755)
		return pluginPath, err
	}

	if int(info.Size()) != len(plugin) {
		// extract or update the ssm-plugin
		err := ioutil.WriteFile(pluginPath, plugin, 0755)
		return "", err
	}

	return pluginPath, nil
}

func RunPlugin(args ...string) error {
	process, err := ExtractAssets()
	if err != nil {
		return nil
	}

	call := exec.Command(process, args...)
	call.Stderr = os.Stderr
	call.Stdout = os.Stdout
	call.Stdin = os.Stdin

	// ignore signal(sigint)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	done := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-sigs:
			case <-done:
				break
			}
		}
	}()
	defer close(done)

	// run subprocess
	if err := call.Run(); err != nil {
		return WrapError(err)
	}
	return nil
}

func WrapError(err error) error {
	if err != nil {
		// Get program counter and line number
		pc, _, line, _ := runtime.Caller(1)
		// Get function name from program counter
		fn := runtime.FuncForPC(pc).Name()
		// Refine function name
		details := strings.Split(fn, "/")
		fn = details[len(details)-1]
		// Build chain
		chainErr := wraperror.Error(err)
		return chainErr.Wrap(fmt.Errorf("[err][%s:%d]", fn, line))
	}
	return nil
}
