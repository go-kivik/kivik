package serve

func (s *Service) log(level Level, msg string) {
	if s.Logger == nil {
		return
	}
	s.Logger.WriteLog(level, msg)
}

// Debug logs a debug message to the registered logger.
func (s *Service) Debug(msg string) {
	s.log(LogLevelDebug, msg)
}

// Info logs an informational message to the registered logger.
func (s *Service) Info(msg string) {
	s.log(LogLevelInfo, msg)
}

// Warn logs a warning message to the registered logger.
func (s *Service) Warn(msg string) {
	s.log(LogLevelWarn, msg)
}

// Error logs an error message to the registered logger.
func (s *Service) Error(msg string) {
	s.log(LogLevelError, msg)
}
