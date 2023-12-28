package main

import (
	"flag"
	"os"

	sdkConfig "github.com/gatewayd-io/gatewayd-plugin-sdk/config"
	"github.com/gatewayd-io/gatewayd-plugin-sdk/logging"
	"github.com/gatewayd-io/gatewayd-plugin-sdk/metrics"
	p "github.com/gatewayd-io/gatewayd-plugin-sdk/plugin"
	v1 "github.com/gatewayd-io/gatewayd-plugin-sdk/plugin/v1"
	"github.com/gatewayd-io/plugin-template-go/plugin"
	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
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

	pluginInstance := plugin.NewTemplatePlugin(plugin.Plugin{
		Logger: logger,
	})

	var config *metrics.MetricsConfig
	if cfg, ok := plugin.PluginConfig["config"].(map[string]interface{}); ok {
		config = metrics.NewMetricsConfig(cfg)
	}
	if config != nil && config.Enabled {
		go metrics.ExposeMetrics(config, logger)
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
