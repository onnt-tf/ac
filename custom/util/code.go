package util

import (
	"fmt"

	"github.com/google/uuid"
)

func GenerateCode(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.New().String())
}
