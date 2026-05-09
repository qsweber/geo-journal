package server

import "errors"

type DeleteInput struct {
	ImageID string
}

func (s *ServerImpl) Delete(userID string, input DeleteInput) error {
	if s.initErr != nil {
		return s.initErr
	}

	if input.ImageID == "" {
		return errors.New("image_id is required")
	}

	return s.s3.DeleteImage(userID, input.ImageID)
}
