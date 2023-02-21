package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gophkeeper/pb"
	"gophkeeper/token"
)

var (
	tokenCachedDir      = os.Getenv("HOME") + "/.cache/gophkeeper/"
	tokenCachedFileName = "token"
)

func (c *Client) loadCachedToken(filepath string) error {
	tokenBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	payload, err := c.tm.VerifyToken(string(tokenBytes))
	if err != nil {
		if errors.Is(err, token.ErrInvalidToken) {
			return err
		}
		if errors.Is(err, token.ErrExpiredToken) {
			return err
		}
	}

	if c.config.User != payload.Username {
		return fmt.Errorf("config user name differs from token user name")
	}

	c.token = string(tokenBytes)

	return nil
}

func (c *Client) saveToken(dir, filename, token string) error {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(dir+filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(token))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) login(ctx context.Context) {
	c.log.Info().Msg("periodic auth check started...")

	_, err := c.tm.VerifyToken(string(c.token))
	if err == nil {
		c.log.Info().Msg("existing token is still valid")
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
			c.log.Warn().Msgf("server connection failed: %s", e.Message())
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

	c.saveToken(tokenCachedDir, tokenCachedFileName, tokenResponse.Value)
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

	err = c.saveToken(tokenCachedDir, tokenCachedFileName, tokenResponse.Value)
	if err != nil {
		c.log.Error().Err(err).Msgf("failed to save token")
	} else {
		c.log.Info().Msgf("successfully saved token to cache file")
	}

	c.token = tokenResponse.Value

	c.log.Info().Msgf("successfully registerd with user '%s'", c.config.User)
}
