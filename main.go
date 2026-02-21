package main

import (
	"encoding/base64"
	"flag"
	"log"
	"maps"
	"os"
	"slices"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/gatewayd-io/gatewayd-plugin-js/plugin"
	sdkConfig "github.com/gatewayd-io/gatewayd-plugin-sdk/config"
	"github.com/gatewayd-io/gatewayd-plugin-sdk/logging"
	"github.com/gatewayd-io/gatewayd-plugin-sdk/metrics"
	p "github.com/gatewayd-io/gatewayd-plugin-sdk/plugin"
	v1 "github.com/gatewayd-io/gatewayd-plugin-sdk/plugin/v1"
	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/spf13/cast"
	pgQuery "github.com/wasilibs/go-pgquery"
)

func setupHelpers(runtime *goja.Runtime) error {
	if err := runtime.Set("btoa", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(runtime.NewTypeError("btoa requires 1 argument"))
		}
		return runtime.ToValue(
			base64.StdEncoding.EncodeToString([]byte(call.Arguments[0].String())))
	}); err != nil {
		return err
	}

	if err := runtime.Set("atob", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(runtime.NewTypeError("atob requires 1 argument"))
		}
		decoded, err := base64.StdEncoding.DecodeString(call.Arguments[0].String())
		if err != nil {
			panic(runtime.NewTypeError("atob: invalid base64 input: " + err.Error()))
		}
		return runtime.ToValue(string(decoded))
	}); err != nil {
		return err
	}

	if err := runtime.Set("parseSQL", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(runtime.NewTypeError("parseSQL requires 1 argument"))
		}
		qStr, err := pgQuery.ParseToJSON(call.Arguments[0].String())
		if err != nil {
			panic(runtime.NewTypeError("parseSQL: " + err.Error()))
		}
		return runtime.ToValue(qStr)
	}); err != nil {
		return err
	}

	return nil
}

func main() {
	sentryDSN := sdkConfig.GetEnv("SENTRY_DSN", "")
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSN,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Sentry SDK: %s", err.Error())
	}

	logLevel := flag.String("log-level", "info", "Log level")
	flag.Parse()

	logger := hclog.New(&hclog.LoggerOptions{
		Level:      logging.GetLogLevel(*logLevel),
		Output:     os.Stderr,
		JSONFormat: true,
		Color:      hclog.ColorOff,
	})

	pluginInstance := plugin.NewJSPlugin(&plugin.Plugin{
		Logger:   logger,
		VM:       goja.New(),
		Bindings: map[string]goja.Callable{},
	})

	registry := require.Registry{}
	registry.Enable(pluginInstance.Impl.VM)

	printer := console.StdPrinter{
		StdoutPrint: func(s string) { pluginInstance.Impl.Logger.Info(s) },
		StderrPrint: func(s string) { pluginInstance.Impl.Logger.Error(s) },
	}
	registry.RegisterNativeModule("console", console.RequireWithPrinter(printer))
	console.Enable(pluginInstance.Impl.VM)

	if err := pluginInstance.Impl.VM.Set("Value", pluginInstance.Impl.VM.ToValue(v1.NewValue)); err != nil {
		logger.Error("Failed to set Value helper function", "error", err)
		return
	}

	cfg := cast.ToStringMap(plugin.PluginConfig["config"])
	if cfg == nil {
		logger.Error("Failed to load plugin config")
		return
	}

	config := metrics.NewMetricsConfig(cfg)
	if config != nil && config.Enabled {
		go metrics.ExposeMetrics(config, logger)
	}

	scriptPath := cast.ToString(cfg["scriptPath"])
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		logger.Error("Failed to read script file", "error", err)
		return
	}
	logger.Debug("Read script file", "bytes", len(script), "path", scriptPath)

	if _, err = pluginInstance.Impl.VM.RunString(string(script)); err != nil {
		logger.Error("Failed to run JS code", "error", err)
		return
	}

	pluginInstance.Impl.RegisterFunctions(slices.Collect(maps.Keys(plugin.Hooks)))

	if err := setupHelpers(pluginInstance.Impl.VM); err != nil {
		logger.Error("Failed to register helper functions", "error", err)
		return
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
