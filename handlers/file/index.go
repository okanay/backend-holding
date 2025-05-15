// handlers/file/index.go
package FileHandler

import (
	FileRepository "github.com/okanay/backend-holding/repositories/file"
	R2Repository "github.com/okanay/backend-holding/repositories/r2"
)

type Handler struct {
	FileRepository *FileRepository.Repository
	R2Repository   *R2Repository.Repository
}

func NewHandler(f *FileRepository.Repository, r2 *R2Repository.Repository) *Handler {
	return &Handler{
		FileRepository: f,
		R2Repository:   r2,
	}
}
