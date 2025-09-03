package config

import (
	"github.com/dv-net/xconfig"
	"github.com/dv-net/xconfig/decoders/xconfigyaml"
	"github.com/dv-net/xconfig/plugins/loader"
	"github.com/dv-net/xconfig/plugins/validate"
	"github.com/go-playground/validator/v10"
)

func Load[T any](configPaths []string, envPrefix string) (*T, error) {
	conf := new(T)

	loader, err := loader.NewLoader(map[string]loader.Unmarshal{
		"yaml": xconfigyaml.New().Unmarshal,
	})
	if err != nil {
		return nil, err
	}

	for _, path := range configPaths {
		if err := loader.AddFile(path, false); err != nil {
			return nil, err
		}
	}

	_, err = xconfig.Load(conf,
		xconfig.WithEnvPrefix(envPrefix),
		xconfig.WithLoader(loader),
		xconfig.WithPlugins(
			validate.New(func(a any) error {
				return validator.New().Struct(a)
			}),
		),
	)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
