package JobHandler

import (
	FileRepository "github.com/okanay/backend-holding/repositories/file"
	JobRepository "github.com/okanay/backend-holding/repositories/job"
	R2Repository "github.com/okanay/backend-holding/repositories/r2"
)

type Handler struct {
	FileRepository *FileRepository.Repository
	R2Repository   *R2Repository.Repository
	JobRepository  *JobRepository.Repository
}

func NewHandler(f *FileRepository.Repository, r2 *R2Repository.Repository, j *JobRepository.Repository) *Handler {
	return &Handler{
		FileRepository: f,
		R2Repository:   r2,
		JobRepository:  j,
	}
}
