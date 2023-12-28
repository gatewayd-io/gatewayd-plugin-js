package main

import (
	"encoding/base64"
	"flag"
	"os"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/gatewayd-io/gatewayd-plugin-js/plugin"
	sdkConfig "github.com/gatewayd-io/gatewayd-plugin-sdk/config"
	"github.com/gatewayd-io/gatewayd-plugin-sdk/logging"
	"github.com/gatewayd-io/gatewayd-plugin-sdk/metrics"
	p "github.com/gatewayd-io/gatewayd-plugin-sdk/plugin"
	v1 "github.com/gatewayd-io/gatewayd-plugin-sdk/plugin/v1"
	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/spf13/cast"
	pgQuery "github.com/wasilibs/go-pgquery"
	"golang.org/x/exp/maps"
)

func main() {
	// Parse command line flags, passed by GatewayD via the plugin config
	logLevel := flag.String("log-level", "info", "Log level")
	flag.Parse()

	logger := hclog.New(&hclog.LoggerOptions{
		Level:      logging.GetLogLevel(*logLevel),
		Output:     os.Stderr,
		JSONFormat: true,
		Color:      hclog.ColorOff,
	})

	pluginInstance := plugin.NewJSPlugin(plugin.Plugin{
		Logger:   logger,
		VM:       goja.New(),
		Bindings: map[string]goja.Callable{},
	})

	// Enable the VM to require modules.
	registry := require.Registry{}
	registry.Enable(pluginInstance.Impl.VM)

	// Enable the VM to print to stdout and stderr.
	printer := console.StdPrinter{
		StdoutPrint: func(s string) { pluginInstance.Impl.Logger.Info(s) },
		StderrPrint: func(s string) { pluginInstance.Impl.Logger.Error(s) },
	}
	registry.RegisterNativeModule("console", console.RequireWithPrinter(printer))
	console.Enable(pluginInstance.Impl.VM)

	// Provide the Value helper to the JS code to make it easier to create
	// new values inside v1.Struct.
	pluginInstance.Impl.VM.Set("Value", pluginInstance.Impl.VM.ToValue(v1.NewValue))

	if cfg := cast.ToStringMap(plugin.PluginConfig["config"]); cfg != nil {
		config := metrics.NewMetricsConfig(cfg)
		if config != nil && config.Enabled {
			go metrics.ExposeMetrics(config, logger)
		}

		scriptPath := cast.ToString(cfg["scriptPath"])
		// Read the JS code from the script file.
		script, err := os.ReadFile(scriptPath)
		if err != nil {
			logger.Error("Failed to read script file", "error", err)
			return
		}
		logger.Debug("Read script file", "bytes", len(script), "path", scriptPath)

		_, err = pluginInstance.Impl.VM.RunString(string(script))
		if err != nil {
			logger.Error("Failed to run JS code", "error", err)
			return
		}

		// Register the JS functions as Go functions.
		pluginInstance.Impl.RegisterFunctions(maps.Keys(plugin.Hooks))

		// Register helper functions.
		pluginInstance.Impl.VM.Set("btoa", func(call goja.FunctionCall) goja.Value {
			return pluginInstance.Impl.VM.ToValue(base64.RawStdEncoding.EncodeToString([]byte(call.Arguments[0].String())))
		})
		pluginInstance.Impl.VM.Set("atob", func(call goja.FunctionCall) goja.Value {
			str, _ := base64.RawStdEncoding.DecodeString(call.Arguments[0].String())
			return pluginInstance.Impl.VM.ToValue(string(str))
		})
		pluginInstance.Impl.VM.Set("parseSQL", func(call goja.FunctionCall) goja.Value {
			qStr, _ := pgQuery.ParseToJSON(call.Arguments[0].String())
			return pluginInstance.Impl.VM.ToValue(qStr)
		})
	}

	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: goplugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   sdkConfig.GetEnv("MAGIC_COOKIE_KEY", ""),
			MagicCookieValue: sdkConfig.GetEnv("MAGIC_COOKIE_VALUE", ""),
		},
		Plugins: v1.GetPluginSetMap(map[string]goplugin.Plugin{
			plugin.PluginID.GetName(): pluginInstance,
		}),
		GRPCServer: p.DefaultGRPCServer,
		Logger:     logger,
	})
}
