package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

const uploadRoot = "ex"

// UploadFile uploads a new file for an exploit
// the file will have the execution bit set for the user
func (s *Server) UploadFile(serviceName string, exploitName string, fileName string, content io.ReadCloser) (filepath string, err error) {
	s.log.Println("Uploading file:", serviceName, exploitName, fileName)
	if strings.TrimSpace(fileName) == "" {
		return "", fmt.Errorf("invalid file name")
	}

	filepath = path.Join(uploadRoot, serviceName, exploitName, fileName)
	err = os.MkdirAll(path.Join(uploadRoot, serviceName, exploitName), 0755)
	if err != nil {
		s.log.Errorln("Cannot create directory for upload:", err)
		return "", fmt.Errorf("create dir: %w", err)
	}

	fl, err := os.OpenFile(filepath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0744)
	if err != nil {
		s.log.Errorln("Cannot create file:", err)
		return "", fmt.Errorf("create file: %w", err)
	}
	_, err = io.Copy(fl, content)
	if err != nil {
		s.log.Errorln("Cannot write:", err)
		return "", fmt.Errorf("writing: %w", err)
	}

	return filepath, nil
}
