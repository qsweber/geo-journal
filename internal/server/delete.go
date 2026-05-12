package server

import (
	"errors"

	"github.com/qsweber/geo-journal/internal/clients"
)

type DeleteInput struct {
	ImageID string
}

var (
	ErrImageIDRequired = errors.New("imageID is required")
	ErrImageNotFound   = errors.New("image not found")
)

func (s *ServerImpl) Delete(userID string, input DeleteInput) error {
	if s.initErr != nil {
		return s.initErr
	}

	if input.ImageID == "" {
		return ErrImageIDRequired
	}

	err := s.s3.DeleteImage(userID, input.ImageID)
	if errors.Is(err, clients.ErrImageNotFound) {
		return ErrImageNotFound
	}
	return err
}
