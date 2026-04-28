package config

import (
	core "github.com/ynhuu/gin-core"
	"go.uber.org/zap"
)

type Config struct {
	SessionDir    string            `yaml:"SessionDir"`
	TokenUsageDir string            `yaml:"TokenUsageDir"`
	Secret        []string          `yaml:"Secret"`
	Metrics       string            `yaml:"Metrics"`
	ModelPlanType map[string]string `yaml:"ModelPlanType"`
	SK            map[string]struct{}
}

func MustLoad(path string) Config {
	data, err := core.ReadYml[Config](path)
	if err != nil {
		zap.L().Fatal("MustLoad", zap.Error(err))
	}
	data.SK = make(map[string]struct{}, len(data.Secret))
	for _, v := range data.Secret {
		data.SK[v] = struct{}{}
	}
	return data
}

func (c Config) PlanType(model string) string {
	return c.ModelPlanType[model]
}
