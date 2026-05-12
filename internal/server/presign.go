package server

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/qsweber/geo-journal/internal/clients"
)

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

	url, data, err := s.s3.CreatePresignedPost(userID, clients.ImageMetadata{
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
