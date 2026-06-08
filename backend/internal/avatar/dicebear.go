package avatar

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/officeryoda/dozingo/internal/storage"
)

func RandomProfilePicture(seed string) (*storage.Image, error) {
	url := fmt.Sprintf("https://api.dicebear.com/10.x/miniavs/svg?seed=%s", seed)
	return avatarFromURL(url)
}

func avatarFromURL(url string) (*storage.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer resp.Body.Close()

	avatarBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read avatar: %w", err)
	}

	file := bytes.NewReader(avatarBytes)

	contentType := resp.Header.Get("Content-Type")
	exts, _ := mime.ExtensionsByType(contentType)
	if len(exts) == 0 {
		return nil, fmt.Errorf("no extension known for content type %q", contentType)
	}
	extension := exts[0] // includes the dot: ".svg"

	return &storage.Image{
		File:        file,
		ContentType: contentType,
		Extension:   extension,
	}, nil
}
