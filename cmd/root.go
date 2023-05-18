/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/Fishwaldo/go-yocto/backends"
	"github.com/Fishwaldo/go-yocto/parsers"
	"github.com/Fishwaldo/go-yocto/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-yocto",
	Short: "Manage Yocto Recipes from Sources",
	Long: `Manage Yocto Recipes from Sources`,
}

var cfgFile string

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/go-yocto.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv()

	utils.InitLogger()

	if err := utils.Config.InitConfig(); err != nil {
		utils.Logger.Error("Failed to initialize Logger", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	if err := backends.Init(); err != nil {
		utils.Logger.Error("Failed to initialize Backends", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	if err := parsers.InitParsers(); err != nil {
		utils.Logger.Error("Failed to initialize Parsers", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	if err := backends.LoadCache(); err != nil {
		utils.Logger.Error("Failed to Load Cache", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
}
