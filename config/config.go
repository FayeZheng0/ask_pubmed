package config

type (
	Config struct {
		Admins   []string `yaml:"admins"`
		Botastic Botastic `json:"botastic" yaml:"botastic"`
	}

	Botastic struct {
		Endpoint  string `json:"endpoint" yaml:"endpoint"`
		AppID     string `json:"app_id" yaml:"app_id"`
		AppSecret string `json:"app_secret" yaml:"app_secret"`
		BotID     uint64 `json:"bot_id" yaml:"bot_id"`
	}
)
