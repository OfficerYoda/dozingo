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
	return profilePictureFromURL(url)
}

func RandomProfilePictureBots(seed string) (*storage.Image, error) {
	url := fmt.Sprintf("https://api.dicebear.com/10.x/bottts/svg?seed=%s", seed)
	return profilePictureFromURL(url)
}

func profilePictureFromURL(url string) (*storage.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return &storage.Image{}, fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer resp.Body.Close()

	avatarBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &storage.Image{}, fmt.Errorf("failed to read avatar: %w", err)
	}

	file := bytes.NewReader(avatarBytes)

	contentType := resp.Header.Get("Content-Type")
	exts, _ := mime.ExtensionsByType(contentType)
	extension := exts[0] // ".svg" (includes the dot)

	return &storage.Image{
		File:        file,
		ContentType: contentType,
		Extension:   extension,
	}, nil
}
