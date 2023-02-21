package client

import (
	"context"
	"time"
)

func (c *Client) cleanJob(ctx context.Context) {
	ticker := time.NewTicker(c.config.Clean)

	c.log.Info().Msg("started periodic db cleaning")

	for {
		select {
		case <-ctx.Done():
			c.log.Info().Msg("finished periodic cleaning db")
			return
		case <-ticker.C:
			c.clean(ctx)
		}
	}
}

func (c *Client) clean(ctx context.Context) {
	deletedSecrets, err := c.storage.CleanSecrets(ctx)
	if err != nil {
		c.log.Error().Msg("failed to clean deleted secrets")
		return
	}

	c.log.Info().Msgf("cleaned up %v deleted secrets", len(deletedSecrets))
}
