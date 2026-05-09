package server

type StatusOutput struct {
	Text string `json:"text"`
}

func (s *ServerImpl) Status() (StatusOutput, error) {
	if s.initErr != nil {
		return StatusOutput{}, s.initErr
	}
	return StatusOutput{Text: "ok"}, nil
}
