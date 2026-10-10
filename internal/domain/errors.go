package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnsupportedFormat = errors.New("unsupported file format")
	ErrFileTooLarge      = errors.New("file exceeds size limit")
	ErrEmptyContent      = errors.New("resume content is empty")
	ErrParseFailed       = errors.New("failed to parse resume")
	ErrExportUnsupported = errors.New("export format not supported")
	ErrInvalidStatus     = errors.New("invalid suggestion status transition")
)
