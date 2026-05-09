package server

import "github.com/qsweber/geo-journal/internal/clients"

type Server interface {
	Status() (StatusOutput, error)
	Images(userID string) (ImagesOutput, error)
	Presign(userID string, input PresignInput) (PresignOutput, error)
	Delete(userID string, input DeleteInput) error
}

type ServerImpl struct {
	s3      clients.S3Client
	initErr error
}

func New() Server {
	client, err := clients.NewS3Client()
	return &ServerImpl{s3: client, initErr: err}
}
