package token

import (
	"time"

	"github.com/o1egl/paseto"
)

const (
	symmetricKey         = "cuzyouwillneverknowthissecretkey"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func NewPasetoMaker() PasetoMaker {
	return PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}
}

func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload := NewPayload(username, duration)

	token, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	return token, err
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, ErrExpiredToken
	}

	return payload, nil
}
