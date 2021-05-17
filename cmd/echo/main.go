package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/polygon-io/errands-server/schemas"
	log "github.com/sirupsen/logrus"

	"github.com/polygon-io/errands-go"
)

const (
	paramEchoText = "echo"
	paramFailErrand = "fail"
)

type Config struct {
	ErrandsURL   string `envconfig:"ERRANDS_URL" required:"true"`
	ErrandsTopic string `split_words:"true" default:"echo"`
}

func run() error {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return fmt.Errorf("process env config: %w", err)
	}

	errandsClient := errands.New(cfg.ErrandsURL)
	processor, err := errandsClient.NewProcessor(cfg.ErrandsTopic, 1, handleErrand)
	if err != nil {
		return fmt.Errorf("new errand processor: %w", err)
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a kill signal
	<-sigs

	// Tell the processor to quit
	processor.Quit <- 1
	return nil
}

func handleErrand(errand *schemas.Errand) (map[string]interface{}, error) {
	if echo, exists := errand.Data[paramEchoText]; exists {
		log.WithFields(log.Fields{
			"errand_id": errand.ID,
			"echo": echo,
		}).Info("got something to echo")
	}

	if shouldFail, exists := errand.Data[paramFailErrand]; exists {
		if shouldFail.(bool) {
			return nil, errors.New("you told me to fail")
		}
	}

	return nil, nil
}

func main() {
	if err := run(); err != nil {
		log.WithError(err).Fatal("application closing fatally")
	}
}
