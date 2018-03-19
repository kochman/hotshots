package config

import "os"
import "path"

type Config struct {
	ListenURL       string
	PhotosDirectory string
	PusherDestURL   string
}

func New() (*Config, error) {
	c := &Config{
		ListenURL:       "127.0.0.1:8000",
		PhotosDirectory: "/var/hotshots",
		PusherDestURL:   "http://127.0.0.1:8000",
	}

	dir, ok := os.LookupEnv("HOTSHOTS_DIR")
	if ok {
		c.PhotosDirectory = dir
	}

	listen, ok := os.LookupEnv("HOTSHOTS_LISTEN_URL")
	if ok {
		c.ListenURL = listen
	}

	dest, ok := os.LookupEnv("HOTSHOTS_PUSHER_DEST_URL")
	if ok {
		c.PusherDestURL = dest
	}

	return c, nil
}

func (c *Config) ImgFolder() string {
	return path.Join(c.PhotosDirectory, "/img")
}
func (c *Config) ConfFolder() string {
	return path.Join(c.PhotosDirectory, "/conf.d")
}

func (c *Config) StormFile() string {
	return path.Join(c.ConfFolder(), "/hotshot.db")
}
