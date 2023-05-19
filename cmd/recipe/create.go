/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmdRecipe

import (
	"github.com/Fishwaldo/go-yocto/recipe"
	"github.com/Fishwaldo/go-yocto/utils"
	"github.com/spf13/cobra"
)

// createCmd represents the update command
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new recipe",
	Long: `Create a new recipe`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		utils.Logger.Info("Creating Recipe", utils.Logger.Args("backend", args[0], "name", args[1]))
		recipe.CreateRecipe(args[0], args[1])
	},
}

func init() {
}
