package JobHandler

import (
	FileRepository "github.com/okanay/backend-holding/repositories/file"
	JobRepository "github.com/okanay/backend-holding/repositories/job"
	R2Repository "github.com/okanay/backend-holding/repositories/r2"
	"github.com/okanay/backend-holding/services/cache"
)

type Handler struct {
	FileRepository *FileRepository.Repository
	R2Repository   *R2Repository.Repository
	JobRepository  *JobRepository.Repository
	Cache          *cache.Cache
}

func NewHandler(f *FileRepository.Repository, r2 *R2Repository.Repository, j *JobRepository.Repository, c *cache.Cache) *Handler {
	return &Handler{
		FileRepository: f,
		R2Repository:   r2,
		JobRepository:  j,
		Cache:          c,
	}
}
