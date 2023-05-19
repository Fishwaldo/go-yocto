/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/Fishwaldo/go-yocto/utils"
	"github.com/spf13/cobra"
	"github.com/Fishwaldo/go-yocto/cmd/recipe"
	"github.com/spf13/viper"
)

// recipeCmd represents the recipe command
var recipeCmd = &cobra.Command{
	Use:   "recipe",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("recipe called")
	},
}

func init() {
	rootCmd.AddCommand(recipeCmd)
	recipeCmd.AddCommand(cmdRecipe.CreateCmd)

	cmdRecipe.CreateCmd.Flags().StringP("layer", "l", ".", "Layer Directory to create recipe in")
	if err := viper.BindPFlag("yocto.layerdirectory", cmdRecipe.CreateCmd.Flags().Lookup("layer")); err != nil {
		utils.Logger.Error("Failed to Bind Flag", utils.Logger.Args("flag", "layer", "error", err))
	}
}
