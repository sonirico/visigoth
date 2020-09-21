package server

import (
	"errors"
	"strings"
)

var badURLError = errors.New("bad url")

func parseIndex(path string) (string, error) {
	pams := strings.Split(path, "/")
	// empty string, api, subapi, index
	if len(pams) < 4 {
		return "", badURLError
	}

	if len(pams[3]) < 1 {
		return "", badURLError
	}

	return pams[3], nil
}
