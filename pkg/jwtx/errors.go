package jwtx

import "errors"

var (
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotYetValid = errors.New("token not yet valid")
	ErrTokenInvalid     = errors.New("token invalid")
)
