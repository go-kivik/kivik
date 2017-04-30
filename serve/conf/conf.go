package conf

import (
	"os"
	"os/user"

	"github.com/spf13/viper"
)

// Conf represents a loaded configuration.
type Conf struct {
	*viper.Viper
}

// New returns an empty Conf.
func New() *Conf {
	return &Conf{Viper: viper.New()}
}

// Load loads the specified config file.
func Load(file string) (*Conf, error) {
	if file != "" {
		return load(file)
	}
	c, err := load("")
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		return c, nil
	}
	return c, nil
}

func load(file string) (*Conf, error) {
	v := viper.New()
	if file == "" {
		v.SetConfigName("serve")
		v.SetConfigType("toml")
		v.AddConfigPath(".")
		if u, err := user.Current(); err == nil {
			if u.HomeDir != "" {
				v.AddConfigPath(u.HomeDir + string(os.PathSeparator) + "kivik/")
			}
		}
		v.AddConfigPath("/etc/kivik/") // TODO: Add explicit support for Windows & MacOS
	} else {
		v.SetConfigFile(file)
	}
	return &Conf{v}, v.ReadInConfig()
}
