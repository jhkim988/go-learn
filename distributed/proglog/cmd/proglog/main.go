package main

import (
	"log"

	"github.com/jhkim988/proglog/internal/agent"
	"github.com/jhkim988/proglog/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	cli := &cli{}

	cmd := &cobra.Command{
		Use:     "proglog",
		PreRunE: cli.setupConfig,
		RunE:    cli.run,
	}

	if err := setupFlags(cmd); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

type cli struct {
	cfg cfg
}

type cfg struct {
	agent.Config
	ServerTLSConfig config.TLSConfig
	PeerTLSConfig   config.TLSConfig
}

func setupFlags(cmd *cobra.Command) error {
	panic("unimplemented")
}

func (*cli) setupConfig(cmd *cobra.Command, args []string) error {
	panic("unimplemented")
}
func (*cli) run(cmd *cobra.Command, args []string) error {
	panic("unimplemented")
}
