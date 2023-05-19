/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmdSource

import (
	"github.com/Fishwaldo/go-yocto/backends"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var SearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search For Sources accross all packages",
	Long: `search for sources accross all packages`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sources, err := backends.SearchSource("", args[0])
		if err == nil {
			td := pterm.TableData{{"Name", "Description", "Backend", "Url"}}
			for _, source := range sources {
				td = append(td, []string{source.Name, source.Description, source.BackendID, source.Url})
			}
			pterm.DefaultTable.WithHasHeader().WithData(
				td,
			).Render()
		}
	},
}

func init() {


	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
