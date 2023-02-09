package client

import (
	"context"
	"errors"
	"os"
	"time"

	"gophkeeper/pb"
	"gophkeeper/token"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	tokenCachedDir  = "/.cache/gophkeeper/"
	tokenCachedFile = "token"
)

func (c *Client) loadCachedToken() {
	home := os.Getenv("HOME")

	tokenBytes, err := os.ReadFile(home + tokenCachedDir + tokenCachedFile)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to load cached token")
		return
	}

	c.log.Info().Msg("successfully loaded cached token")

	payload, err := c.tm.VerifyToken(string(tokenBytes))
	if err != nil {
		if errors.Is(err, token.ErrInvalidToken) {
			c.log.Error().Err(err).Msg("cached token is not valid")
			return
		}
		if errors.Is(err, token.ErrExpiredToken) {
			c.log.Error().Err(err).Msg("cached token has expired")
			return
		}
	}

	if c.config.User != payload.Username {
		c.log.Error().Msgf("cached token does not belong to you, %s", c.config.User)
		return
	}

	c.token = string(tokenBytes)
}

func (c *Client) saveToken(token string) {
	home := os.Getenv("HOME")

	err := os.MkdirAll(home+tokenCachedDir, 0700)
	if err != nil {
		c.log.Error().Err(err).Msgf("field to create cache dir")
		return
	}

	file, err := os.OpenFile(home+tokenCachedDir+tokenCachedFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		c.log.Error().Err(err).Msgf("field to open/create token cache file")
		return
	}
	defer file.Close()

	_, err = file.Write([]byte(token))
	if err != nil {
		c.log.Error().Err(err).Msgf("field to write token to cache file")
	}

	c.log.Info().Msgf("successfully saved token to cache file")
}

func (c *Client) login(ctx context.Context) {
	c.log.Info().Msg("periodic logging in...")

	_, err := c.tm.VerifyToken(string(c.token))
	if err == nil {
		c.log.Info().Msg("already authenticated")
		return
	}
	if errors.Is(err, token.ErrExpiredToken) {
		c.log.Warn().Msg("token has expired...renewing")
	}

	tokenResponse, err := c.g.Login(ctx, &pb.User{Name: c.config.User, Password: c.config.Password})
	if err != nil {
		e, ok := status.FromError(err)
		if !ok {
			c.log.Error().Err(err).Msgf("failed to parse login attempt error")
			return
		}

		switch e.Code() {
		case codes.Unavailable:
			c.log.Warn().Msgf("server unavailable: %s", e.Message())
		case codes.InvalidArgument:
			c.log.Error().Msgf("%s: user must be 3-100 letter/digits, password - 6-100 letters", e.Message())
		case codes.NotFound:
			switch e.Message() {
			case "user not found":
				c.log.Info().Msg("user does not exist")
				c.register(ctx)
			case "incorrect password":
				c.log.Error().Msgf("incorrect password for user, %s", c.config.User)
			}
		case codes.Internal:
			c.log.Error().Msgf("failed to login: %s", e.Message())
		}
		return
	}

	c.token = tokenResponse.Value
	c.log.Info().Msgf("successfully logged in with user '%s'", c.config.User)

	c.saveToken(tokenResponse.Value)
}

func (c *Client) loginJob(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(time.Second * 5))

	c.log.Info().Msg("started periodic login check")

	for {
		select {
		case <-ctx.Done():
			c.log.Info().Msg("periodic login check stopped")
			return
		case <-ticker.C:
			c.login(ctx)
		}
	}
}

func (c *Client) register(ctx context.Context) {
	c.log.Info().Msgf("trying to register with user '%s'...", c.config.User)

	tokenResponse, err := c.g.Register(ctx, &pb.User{Name: c.config.User, Password: c.config.Password})
	if err != nil {
		e, ok := status.FromError(err)
		if !ok {
			c.log.Error().Err(err).Msgf("failed to parse register attempt error")
			return
		}

		switch e.Code() {
		case codes.AlreadyExists:
			c.log.Error().Msgf("user '%s' already exists", c.config.User)
		case codes.Internal:
			c.log.Error().Msgf("failed to register: %s", e.Message())
		}
		return
	}

	c.saveToken(tokenResponse.Value)
	c.token = tokenResponse.Value

	c.log.Info().Msgf("successfully registerd with user '%s'", c.config.User)
}
