package internal

import (
	"github.com/gofrs/uuid/v5"
	"strings"
)

func NewUuid() string {
	return strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1)
}
