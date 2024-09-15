package main

import (
<<<<<<< HEAD
	"tokamak-sybil-resistance/database"
	"tokamak-sybil-resistance/database/statedb"
	"tokamak-sybil-resistance/routes"
	"tokamak-sybil-resistance/service/account"
=======
	"fmt"
	"os"
	"os/signal"
	"tokamak-sybil-resistance/config"
	"tokamak-sybil-resistance/log"
	"tokamak-sybil-resistance/node"
>>>>>>> d06ea65 (Added node initialisation for sequencer and initialisation cmd in main)

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/tracerr"

	// "github.com/hermeznetwork/tracerr"
	"github.com/urfave/cli"
)

const (
	flagCfg     = "cfg"
	flagMode    = "mode"
	flagSK      = "privatekey"
	flagYes     = "yes"
	flagBlock   = "block"
	modeSync    = "sync"
	modeCoord   = "coord"
	nMigrations = "nMigrations"
	flagAccount = "account"
	flagPath    = "path"
)

// Config is the configuration of the node execution
type Config struct {
	mode node.Mode
	node *config.Node
}

func parseCli(c *cli.Context) (*Config, error) {
	cfg, err := getConfig(c)
	if err != nil {
		if err := cli.ShowAppHelp(c); err != nil {
			panic(err)
		}
		return nil, tracerr.Wrap(err)
	}
	return cfg, nil
}

func getConfig(c *cli.Context) (*Config, error) {
	var cfg Config
	mode := c.String(flagMode)
	nodeCfgPath := c.String(flagCfg)
	var err error
	switch mode {
	case modeSync:
		cfg.mode = node.ModeSynchronizer
		cfg.node, err = config.LoadNode(nodeCfgPath, false)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	case modeCoord:
		cfg.mode = node.ModeCoordinator
		cfg.node, err = config.LoadNode(nodeCfgPath, true)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	default:
		return nil, tracerr.Wrap(fmt.Errorf("invalid mode \"%v\"", mode))
	}

	return &cfg, nil
}

func waitSigInt() {
	stopCh := make(chan interface{})

	// catch ^C to send the stop signal
	ossig := make(chan os.Signal, 1)
	signal.Notify(ossig, os.Interrupt)
	const forceStopCount = 3
	go func() {
		n := 0
		for sig := range ossig {
			if sig == os.Interrupt {
				log.Info("Received Interrupt Signal")
				stopCh <- nil
				n++
				if n == forceStopCount {
					log.Fatalf("Received %v Interrupt Signals", forceStopCount)
				}
			}
		}
	}()
	<-stopCh
}

func cmdRun(c *cli.Context) error {
	cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	// TODO: Initialize lof library
	// log.Init(cfg.node.Log.Level, cfg.node.Log.Out)
	innerNode, err := node.NewNode(cfg.mode, cfg.node, c.App.Version)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error starting node: %w", err))
	}
	innerNode.Start()
	waitSigInt()
	innerNode.Stop()

	return nil
}

func main() {
<<<<<<< HEAD
	stateDB := statedb.InitNewStateDB()
	account.NewAccount(stateDB)
	database.InitDB()
=======
	app := cli.NewApp()
	app.Name = "tokamak-node"
	app.Version = "v1"

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     flagMode,
			Usage:    fmt.Sprintf("Set node `MODE` (can be \"%v\" or \"%v\")", modeSync, modeCoord),
			Required: true,
		},
		&cli.StringFlag{
			Name:     flagCfg,
			Usage:    "Node configuration `FILE`",
			Required: false,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "run",
			Aliases: []string{},
			Usage:   "Run the tokamak-node in the indicated mode",
			Action:  cmdRun,
			Flags:   flags,
		},
	}

<<<<<<< HEAD
>>>>>>> 34f8275 (Added initial setup for node in sequencer)
=======
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("\nError: %v\n", tracerr.Sprint(err))
		os.Exit(1)
	}

>>>>>>> 2d9a15e (updated dependencies)
	router := gin.Default()
	routes.Account(router)
	routes.Link(router)
	router.Run("localhost:8080")
}
