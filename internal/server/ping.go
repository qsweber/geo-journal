package server

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type StatusOutput struct {
	Text string `json:"text"`
}

type ImageOutput struct {
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	TakenAt      string `json:"taken_at"`
	Name         string `json:"name"`
	PresignedURL string `json:"presigned_url"`
}

type ImagesOutput struct {
	Images []ImageOutput `json:"images"`
}

type PresignInput struct {
	Latitude  string
	Longitude string
	TakenAt   string
	Name      string
}

type PresignOutput struct {
	URL  string            `json:"url"`
	Data map[string]string `json:"data"`
}

func (s *ServerImpl) Status() (StatusOutput, error) {
	if s.initErr != nil {
		return StatusOutput{}, s.initErr
	}
	return StatusOutput{Text: "ok"}, nil
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
			Latitude:     image.Metadata.Latitude,
			Longitude:    image.Metadata.Longitude,
			TakenAt:      strconv.FormatInt(image.Metadata.TakenAt.Unix(), 10),
			Name:         image.Metadata.Name,
			PresignedURL: image.PresignedURL,
		})
	}

	return ImagesOutput{Images: output}, nil
}

func (s *ServerImpl) Presign(userID string, input PresignInput) (PresignOutput, error) {
	if s.initErr != nil {
		return PresignOutput{}, s.initErr
	}

	takenAtSeconds, err := strconv.ParseInt(input.TakenAt, 10, 64)
	if err != nil {
		return PresignOutput{}, fmt.Errorf("taken_at must be a unix timestamp: %w", err)
	}

	if input.Name == "" || input.Latitude == "" || input.Longitude == "" {
		return PresignOutput{}, errors.New("name, latitude and longitude are required")
	}

	url, data, err := s.s3.CreatePresignedPost(userID, ImageMetadata{
		Name:      input.Name,
		TakenAt:   time.Unix(takenAtSeconds, 0).UTC(),
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
	})
	if err != nil {
		return PresignOutput{}, err
	}

	return PresignOutput{URL: url, Data: data}, nil
}
