package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	errandz "github.com/polygon-io/errands-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultErrandsPort = 5555
)

type errandsCmd struct {
	viper    *viper.Viper
	api      *errandz.ErrandsAPI
	endpoint string
}

func NewCommand() (*cobra.Command, error) {
	ec := &errandsCmd{
		viper: viper.New(),
	}

	cmd := &cobra.Command{
		Use:   "errands",
		Short: "provides a CLI for interacting with the errands service",
		Long: `
		errands is a CLI for interacting with the errands service. It provides
		several subcommands for errand actions like list and delete. It can
		also port-forward the errands server automatically so you don't have to.
		`,
		PersistentPreRunE: ec.rootPersistentPreRun,
	}

	cmd.PersistentFlags().Bool("bootstrap", true, "port-forward the errands server")
	cmd.PersistentFlags().Int("port", defaultErrandsPort, "port for the errands server")
	cmd.PersistentFlags().String("endpoint", "", "if you need to connect to an endpoint other than localhost then use this flag")

	list, err := ec.newListCommand()
	if err != nil {
		return nil, fmt.Errorf("create list command: %w", err)
	}

	delete, err := ec.newDeleteCommand()
	if err != nil {
		return nil, fmt.Errorf("create delete command: %w", err)
	}

	cmd.AddCommand(list)
	cmd.AddCommand(delete)

	ec.viper.SetEnvPrefix("POLY_ERRANDS")
	ec.viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_")) // Make sure env vars use underscore instead of dash
	ec.viper.AutomaticEnv()
	ec.viper.SetConfigName("config")
	ec.viper.AddConfigPath(".") // We prefer local config files over global config files.
	ec.viper.AddConfigPath("$HOME/.errands")

	return cmd, nil
}

// bindViperFlagsPreRun binds the flags for a command in PreRunE.
// This has to be done in pre-run because it can only run for the command+subcommands that are actually going to execute.
func (ec *errandsCmd) bindViperFlagsPreRun(cmd *cobra.Command, _ []string) error {
	if err := ec.viper.BindPFlags(cmd.PersistentFlags()); err != nil {
		return fmt.Errorf("bind persistent flags: %w", err)
	}

	if err := ec.viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("bind pflags: %w", err)
	}

	// if bootstrapping then we need to check if the port is available.
	if ec.viper.GetBool("bootstrap") {
		if ec.viper.GetString("endpoint") != "" {
			return fmt.Errorf("cannot use --endpoint with --bootstrap (enabled by default)")
		}

		port := ec.viper.GetInt("port")
		ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))

		if err != nil {
			return fmt.Errorf("port %d is unavailable: %w", port, err)
		}

		if err := ln.Close(); err != nil {
			return fmt.Errorf("close test port listener: %w", err)
		}
	}

	endpoint := ec.viper.GetString("endpoint")
	if endpoint == "" {
		endpoint = fmt.Sprintf("http://localhost:%d", ec.viper.GetInt("port"))
	}
	ec.endpoint = endpoint
	ec.api = errandz.New(endpoint)

	return nil
}

func (e *errandsCmd) rootPersistentPreRun(cmd *cobra.Command, args []string) error {
	if err := e.bindViperFlagsPreRun(cmd, args); err != nil {
		return fmt.Errorf("bind viper flags pre run: %w", err)
	}

	return nil
}

func (ec *errandsCmd) portForwardErrandsServer(ctx context.Context) (*exec.Cmd, error) {
	port := ec.viper.GetInt("port")
	if port == 0 {
		return nil, fmt.Errorf("port is required")
	}

	cmd := createKubeCmdContext(ctx, "polygon", "port-forward", "svc/errands-api", fmt.Sprintf("%d:80", port))
	cmd.Stdout = io.Discard
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start port-forward: %w", err)
	}

	// using the metrics endpoint to check if the server is up
	endpoint := fmt.Sprintf("%s/metrics", ec.endpoint)
	return cmd, poll(ctx, endpoint, 5*time.Second)
}

func poll(ctx context.Context, endpoint string, timeout time.Duration) error {
	endtime := time.Now().Add(timeout)

	for time.Now().Before(endtime) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond) // prevent spamming

			resp, err := http.Get(endpoint)
			if err != nil {
				continue // assume the server is not yet up.
			}

			ioutil.ReadAll(resp.Body) // drain body
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("resp status code: %d", resp.StatusCode)
			}

			// we're good.
			return nil
		}
	}

	return errors.New("timed out waiting for errands server to start")
}

func createKubeCmdContext(ctx context.Context, namespace, subCommand string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "kubectl", subCommand)
	if namespace != "" {
		cmd.Args = append(cmd.Args, "--namespace", namespace)
	}

	cmd.Args = append(cmd.Args, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}
