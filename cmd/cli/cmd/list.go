package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/picogrid/legion-simulations/pkg/utils"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available simulations",
	Long:  `List all available simulations with their descriptions`,
	RunE:  listSimulations,
}

func listSimulations(cmd *cobra.Command, args []string) error {
	// Discover available simulations
	simInfos, err := utils.DiscoverSimulations()
	if err != nil {
		return fmt.Errorf("failed to discover simulations: %w", err)
	}

	if len(simInfos) == 0 {
		fmt.Println("No simulations found")
		return nil
	}

	// Create tabwriter for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tVERSION\tCATEGORY\tDESCRIPTION")
	_, _ = fmt.Fprintln(w, "----\t-------\t--------\t-----------")

	for _, info := range simInfos {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			info.Config.Name,
			info.Config.Version,
			info.Config.Category,
			info.Config.Description,
		)
	}

	return w.Flush()
}
