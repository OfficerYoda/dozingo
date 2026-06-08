// Package avatar generates and resolves user avatar images and their public URLs.
package avatar

import (
	"fmt"
	"net/url"
	"strings"
)

type URLBuilder struct {
	scheme     string // "http" or "https"
	hostSuffix string // host[:port] without the bucket prefix
	bucket     string // bucket name, used as a leading subdomain
}

func NewURLBuilder(publicURL, bucket string) (*URLBuilder, error) {
	u, err := url.Parse(publicURL)
	if err != nil {
		return nil, fmt.Errorf("avatar: parse public URL %q: %w", publicURL, err)
	}

	return &URLBuilder{
		scheme:     u.Scheme,
		hostSuffix: u.Host,
		bucket:     bucket,
	}, nil
}

func (b *URLBuilder) URL(avatarKey string) *string {
	if b == nil {
		return nil
	}
	if strings.TrimSpace(avatarKey) == "" {
		return nil
	}
	u := url.URL{
		Scheme: b.scheme,
		Host:   b.bucket + "." + b.hostSuffix,
		Path:   "/" + avatarKey,
	}
	s := u.String()

	return &s
}
