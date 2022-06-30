package cmd

import (
	"context"
	"fmt"
	"os"

	errandz "github.com/polygon-io/errands-go"
	"github.com/spf13/cobra"
)

func (ec *errandsCmd) newDeleteCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "deletes errands by ID, type, or status",
		RunE:    ec.delete,
		PreRunE: ec.bindViperFlagsPreRun,
	}

	cmd.Flags().String("type", "", "Filter by errand type")
	cmd.Flags().String("status", "failed", "Filter by status; comma delimited")
	cmd.Flags().String("id", "", "ID of the errand to delete")
	cmd.Flags().Bool("dry-run", false, "Don't actually delete anything. Only used for bulk deletion")

	return cmd, nil
}

func (ec *errandsCmd) delete(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if ec.viper.GetBool("bootstrap") {
		portCmd, err := ec.portForwardErrandsServer(ctx)
		if err != nil {
			return fmt.Errorf("port-forward: %w", err)
		}

		defer func() {
			if err := portCmd.Process.Signal(os.Interrupt); err != nil {
				fmt.Printf("error killing port-forward: %s\n", err)
			}
		}()
	}

	if id := ec.viper.GetString("id"); id != "" {
		if err := deleteErrand(ec.api, id); err != nil {
			return fmt.Errorf("failed to delete errand %s: %w", id, err)
		}

		return nil
	}

	jobs, err := listErrandsForTopic(ec.api, ec.viper.GetString("type"), ec.viper.GetString("status"))
	if err != nil {
		return fmt.Errorf("failed to get errands: %w", err)
	}

	for _, job := range jobs {
		name := job.Name
		if len(name) > 100 {
			name = name[:100]
		}

		if ec.viper.GetBool("dry-run") {
			fmt.Printf("(dry-run) delete %s: (%s) %s\n", job.ID, job.Status, name)
			continue
		}

		fmt.Printf("deleting %s\n", job.ID)
		if err := deleteErrand(ec.api, job.ID); err != nil {
			fmt.Printf("failed to delete errand %s: %e", job.ID, err)
		}
	}

	return nil
}

func deleteErrand(api *errandz.ErrandsAPI, id string) error {
	_, err := api.DeleteErrand(id)

	return err
}
