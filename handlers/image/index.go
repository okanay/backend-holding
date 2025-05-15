// handlers/image/index.go
package ImageHandler

import (
	ImageRepository "github.com/okanay/backend-holding/repositories/image"
	R2Repository "github.com/okanay/backend-holding/repositories/r2"
)

type Handler struct {
	ImageRepository *ImageRepository.Repository
	R2Repository    *R2Repository.Repository
}

func NewHandler(i *ImageRepository.Repository, r2 *R2Repository.Repository) *Handler {
	return &Handler{
		ImageRepository: i,
		R2Repository:    r2,
	}
}
