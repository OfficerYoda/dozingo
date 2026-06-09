package avatar

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/officeryoda/dozingo/internal/storage"
)

const dicebearTimeout = 10 * time.Second

func RandomProfilePicture(seed string) (*storage.Image, error) {
	url := fmt.Sprintf("https://api.dicebear.com/10.x/miniavs/svg?seed=%s", seed)
	return avatarFromURL(url)
}

func avatarFromURL(url string) (*storage.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dicebearTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build avatar request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

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
