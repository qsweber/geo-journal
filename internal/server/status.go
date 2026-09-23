package server

import "time"

type StatusOutput struct {
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

func (s *ServerImpl) Status() (StatusOutput, error) {
	if s.initErr != nil {
		return StatusOutput{}, s.initErr
	}
	return StatusOutput{Text: "ok", Timestamp: time.Now().UTC().Format(time.RFC3339)}, nil
}
