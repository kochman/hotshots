package config

type Config struct {
	ListenURL string
}

func New() (*Config, error) {
	return &Config{
		ListenURL: "127.0.0.1:8000",
	}, nil
}
