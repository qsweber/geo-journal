package server

import "strconv"

type ImageOutput struct {
	ID           string `json:"id"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	TakenAt      string `json:"taken_at"`
	Name         string `json:"name"`
	PresignedURL string `json:"presigned_url"`
}

type ImagesOutput struct {
	Images []ImageOutput `json:"images"`
}

func (s *ServerImpl) Images(userID string) (ImagesOutput, error) {
	if s.initErr != nil {
		return ImagesOutput{}, s.initErr
	}
	images, err := s.s3.GetImages(userID)
	if err != nil {
		return ImagesOutput{}, err
	}

	output := make([]ImageOutput, 0, len(images))
	for _, image := range images {
		output = append(output, ImageOutput{
			ID:           image.ID,
			Latitude:     image.Metadata.Latitude,
			Longitude:    image.Metadata.Longitude,
			TakenAt:      strconv.FormatInt(image.Metadata.TakenAt.Unix(), 10),
			Name:         image.Metadata.Name,
			PresignedURL: image.PresignedURL,
		})
	}

	return ImagesOutput{Images: output}, nil
}
