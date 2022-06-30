package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	errandz "github.com/polygon-io/errands-go"
	"github.com/polygon-io/errands-server/schemas"
	"github.com/spf13/cobra"
)

func (ec *errandsCmd) newListCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "returns a list of errands from our errands API",
		PreRunE: ec.bindViperFlagsPreRun,
		RunE:    ec.run,
	}

	cmd.Flags().String("type", "", "filter by errand type")
	cmd.Flags().String("status", "", "filter by status; comma delimited")
	cmd.Flags().Int("port", 5555, "localhost port for the errands server")

	return cmd, nil
}

func (ec *errandsCmd) run(cmd *cobra.Command, args []string) error {
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

	jobs, err := listErrandsForTopic(ec.api, ec.viper.GetString("type"), ec.viper.GetString("status"))
	if err != nil {
		return fmt.Errorf("get errands: %w", err)
	}

	for _, job := range jobs {
		name := job.Name
		if len(name) > 100 {
			name = name[:100]
		}
		fmt.Printf("%100s: %10s | %s | %s\n", name, job.Status, job.ID, time.UnixMilli(job.Created))
	}

	return nil
}

func filterByStatus(jobs []schemas.Errand, status string) []schemas.Errand {
	if status == "" {
		return jobs
	}

	statusFilter := make(map[schemas.Status]bool)
	for _, s := range strings.Split(status, ",") {
		statusFilter[schemas.Status(s)] = true
	}

	var filtered []schemas.Errand
	for _, job := range jobs {
		if statusFilter[job.Status] {
			filtered = append(filtered, job)
		}
	}

	return filtered
}

func listErrandsForTopic(api *errandz.ErrandsAPI, errandType string, status string) ([]schemas.Errand, error) {
	if errandType == "" {
		return nil, errors.New("errand type is required")
	}

	jobs, err := api.ListErrands("type", errandType)
	if err != nil {
		return nil, err
	}

	results := filterByStatus(jobs.Results, status)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Created > results[j].Created
	})

	return results, nil
}
