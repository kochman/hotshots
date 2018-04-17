package config

import "os"
import (
	"path"
	"path/filepath"
)

type Config struct {
	ListenURL       string
	PhotosDirectory string
	WebDirectory    string
}

func New() (*Config, error) {
	c := &Config{
		ListenURL:       "127.0.0.1:8000",
		PhotosDirectory: "/var/hotshots",
	}

	hotshotsDir, ok := os.LookupEnv("HOTSHOTS_DIR")
	if ok {
		c.PhotosDirectory = hotshotsDir
	}

	listenURL, ok := os.LookupEnv("HOTSHOTS_LISTEN_URL")
	if ok {
		c.ListenURL = listenURL
	}

	webDir, ok := os.LookupEnv("HOTSHOTS_WEB_DIR")
	if ok {
		c.WebDirectory = webDir
	} else {
		webDir, _ = os.Getwd()
		c.WebDirectory = filepath.Join(webDir, "web")
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
