package UserHandler

import (
	TokenRepository "github.com/okanay/backend-holding/repositories/token"
	UserRepository "github.com/okanay/backend-holding/repositories/user"
)

type Handler struct {
	UserRepository  *UserRepository.Repository
	TokenRepository *TokenRepository.Repository
}

func NewHandler(u *UserRepository.Repository, t *TokenRepository.Repository) *Handler {
	return &Handler{
		UserRepository:  u,
		TokenRepository: t,
	}
}
