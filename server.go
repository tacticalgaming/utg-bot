package main

import (
	"fmt"
)

type Server struct {
	config *Config
}

func (s *Server) Init(config *Config) error {
	if config == nil {
		return fmt.Errorf("nil config")
	}
	s.config = config
	return nil
}

func (s *Server) handle(uid string, params []string) (string, error) {
	if len(params) >= 1 {
		if params[0] == "help" {
			return getServerHelp(), nil
		}
	}

	return getServerHelp(), nil
}

func getServerHelp() string {
	text := "Управление сервером\n"
	text = "update - обновление сервера\n"
	text = "restart - перезапуск сервера\n"

	return text
}
