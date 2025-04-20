package config

type SourceConfig struct {
	Name    string
	Enabled bool
	Order   int
}

var sourceConfigs = []SourceConfig{
	{Name: "DSCA", Enabled: true, Order: 1},        // Defense news first
	{Name: "WhiteHouse", Enabled: true, Order: 2},   // Government news second
	{Name: "CNN", Enabled: true, Order: 3},         // General news last
}

func GetEnabledSources() []SourceConfig {
	enabled := make([]SourceConfig, 0)
	for _, cfg := range sourceConfigs {
		if cfg.Enabled {
			enabled = append(enabled, cfg)
		}
	}
	// Sort by Order
	for i := 0; i < len(enabled)-1; i++ {
		for j := i + 1; j < len(enabled); j++ {
			if enabled[i].Order > enabled[j].Order {
				enabled[i], enabled[j] = enabled[j], enabled[i]
			}
		}
	}
	return enabled
}