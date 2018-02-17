package config

import "path"

type Config struct {
	ListenURL 		string
	PhotosDirectory	string
}

func New() (*Config, error) {
	return &Config{
		ListenURL: "127.0.0.1:8000",
		PhotosDirectory: "/var/hotshots",
	}, nil
}

func (c *Config) ImgFolder() string {
	return path.Join(c.PhotosDirectory, "/img")
}
func (c *Config) ConfFolder() string {
	return path.Join(c.PhotosDirectory, "/conf.d")
}

func (c *Config) StormFile() string{
	return path.Join(c.ConfFolder(), "/hotshot.db")
}