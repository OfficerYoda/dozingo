package storage

import "bytes"

type Image struct {
	File        *bytes.Reader
	ContentType string
	Extension   string // Extension includes the dot ".svg"
}

func NewImage(file *bytes.Reader, contentType, extension string) *Image {
	return &Image{
		File:        file,
		ContentType: contentType,
		Extension:   extension,
	}
}
