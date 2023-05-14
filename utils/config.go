package utils

import (
	"os"

	"github.com/spf13/viper"
)

type configData struct {
	BaseDir string
	KDEConfig struct {
		Release string
		DefaultBranch string
		AccessToken string
		KDEGitLabURL string
	}
}

var Config configData

func (c *configData) InitConfig() (err error) {

	viper.SetConfigName("go-yocto")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config")
	viper.ReadInConfig()
	err = viper.Unmarshal(&c)
	if err != nil {
		Logger.Error("Failed to unmarshal config", Logger.Args("error", err))
	}
	if _, err = os.Stat(c.BaseDir); os.IsNotExist(err) {
		Logger.Error("BaseDir does not exist", Logger.Args("error", err, "basedir", c.BaseDir))
		os.Exit(-1)
	}
	return err
}
