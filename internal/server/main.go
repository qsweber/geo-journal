package server

type Server interface {
	Status() (StatusOutput, error)
	Images(userID string) (ImagesOutput, error)
	Presign(userID string, input PresignInput) (PresignOutput, error)
}

type ServerImpl struct {
	s3      S3Client
	initErr error
}

func New() Server {
	client, err := NewS3Client()
	return &ServerImpl{s3: client, initErr: err}
}
