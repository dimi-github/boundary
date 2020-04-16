package controller

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/mlock"
)

type Controller struct {
	conf   *Config
	logger hclog.Logger

	baseContext context.Context
	baseCancel  context.CancelFunc
}

func New(conf *Config) (*Controller, error) {
	c := &Controller{
		conf:   conf,
		logger: conf.Logger,
	}

	if c.logger == nil {
		c.logger = hclog.New(&hclog.LoggerOptions{
			Level: hclog.Trace,
		})
		conf.AllLoggers = append(conf.AllLoggers, c.logger)
	}

	if conf.SecureRandomReader == nil {
		conf.SecureRandomReader = rand.Reader
	}

	if !conf.RawConfig.DisableMlock {
		// Ensure our memory usage is locked into physical RAM
		if err := mlock.LockMemory(); err != nil {
			return nil, fmt.Errorf(
				"Failed to lock memory: %v\n\n"+
					"This usually means that the mlock syscall is not available.\n"+
					"Watchtower uses mlock to prevent memory from being swapped to\n"+
					"disk. This requires root privileges as well as a machine\n"+
					"that supports mlock. Please enable mlock on your system or\n"+
					"disable Watchtower from using it. To disable Watchtower from using it,\n"+
					"set the `disable_mlock` configuration option in your configuration\n"+
					"file.",
				err)
		}
	}

	c.logger = c.logger.Named("controller")

	c.baseContext, c.baseCancel = context.WithCancel(context.Background())

	return c, nil
}

func (c *Controller) Start() error {
	if err := c.startListeners(); err != nil {
		return fmt.Errorf("error starting controller listeners: %w", err)
	}
	return nil
}

func (c *Controller) Shutdown() error {
	if err := c.stopListeners(); err != nil {
		return fmt.Errorf("error stopping controller listeners: %w", err)
	}
	return nil
}

func (c *Controller) SetLogLevel(level hclog.Level) {
	for _, logger := range c.conf.AllLoggers {
		logger.SetLevel(level)
	}
}