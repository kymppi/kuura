package kuura

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	LISTEN            string `env:"LISTEN" envDefault:"0.0.0.0:4000"`
	MANAGEMENT_LISTEN string `env:"MANAGEMENT_LISTEN" envDefault:"0.0.0.0:4001"`
	GO_ENV            string `env:"GO_ENV" envDefault:"production"`
	DATABASE_URL      string `env:"DATABASE_URL" envDefault:""`
	RUN_MIGRATIONS    bool   `env:"RUN_MIGRATIONS" envDefault:"false"`
	DEBUG             bool   `env:"DEBUG" envDefault:"false"`

	JWK_KEK_PATH string `env:"JWK_KEK_PATH" envDefault:"/var/kuura/.kek"`
	JWT_ISSUER   string `env:"JWT_ISSUER" envDefault:"kuura.midka.dev"`

	SRP_PRIME     string `env:"SRP_PRIME" envDefault:"00b14cbeb5826b34e3075714520de2af615885f244358e498a04de5dea9d79aa142f0624239261bb2309faa250a9c4b56229282f6ad7ef3b44c59521f32b30c62d057c25b7f7992618a3d1329390eaa0c1c12a13290101d77acd8d3969556868a8b4842a28cf2910c431efd3da63d61e5c6f032f745f539996157bc6b5f6bf8d7a3f7287950d84fec7d5227ed46a15206572acd370c8ae80b9a28b6938d1d8f89f6402c8f64d46459506506e6e2c51b43ddf344148243a6b82c409ef9aec540c26eff0d3124c08e49e52f2d8fd32acb1e6dac5c580110153c44631d324acbae652283c7258d999cd38befb9968906e221d7faa366709972b2a45c0736303e84f848ffed435f2f4185ab70fde271647bf26aebf86f8ac7211b965ea959298cfaeff206a60c55f3534ca05eaf71232762ec54398f1cb554002f901d0afdfb3ad84d4a2dce14b6afb0e4197a9a617342ad80310f5460762e5883251d664abe2d8e92678b2723e9eb7a28ae1d55efe2987611a950657f26398d4bf5ebecfa24bcec597"`
	SRP_GENERATOR string `env:"SRP_GENERATOR" envDefault:"2"`
}

func ParseConfig() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
