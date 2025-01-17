package bootstrap

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"fmt"
)

// Initialize initializes all necessary components (config, MySQL, logger)
func Initialize() error {
	if err := logger.InitLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger, err: %w", err)
	}
	if err := database.InitMySQL(); err != nil {
		return fmt.Errorf("failed to initialize MySQL, err: %w", err)
	}
	return nil
}
