package service

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
)

const (
	AvatarDirectory = "../frontend/images/users"
	AvatarURLPrefix = "/users/"
	MaxAvatarBytes  = 5 << 20
	maxAvatarPixels = 4096 * 4096
)

// AvatarStore only writes server-generated names; it never uses upload filenames.
type AvatarStore struct {
	directory string
}

func NewAvatarStore(directory string) *AvatarStore {
	return &AvatarStore{directory: directory}
}

func (s *AvatarStore) save(source io.Reader) (url string, cleanup func() error, err error) {
	data, err := io.ReadAll(io.LimitReader(source, MaxAvatarBytes+1))
	if err != nil {
		return "", nil, fmt.Errorf("read avatar: %w", err)
	}
	if len(data) > MaxAvatarBytes {
		return "", nil, ErrAvatarTooLarge
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") || config.Width <= 0 || config.Height <= 0 ||
		config.Width > 4096 || config.Height > 4096 || config.Width*config.Height > maxAvatarPixels {
		return "", nil, ErrInvalidRequest
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", nil, ErrInvalidRequest
	}
	// Flatten transparency on white before encoding a real JPEG. Re-encoding also
	// removes client metadata and any appended payload.
	rgb := image.NewRGBA(decoded.Bounds())
	draw.Draw(rgb, rgb.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(rgb, rgb.Bounds(), decoded, decoded.Bounds().Min, draw.Over)
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, rgb, &jpeg.Options{Quality: 85}); err != nil {
		return "", nil, fmt.Errorf("encode avatar: %w", err)
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", nil, fmt.Errorf("avatar UUID: %w", err)
	}
	id[6] = id[6]&0x0f | 0x40
	id[8] = id[8]&0x3f | 0x80
	name := fmt.Sprintf("%x-%x-%x-%x-%x.jpg", id[:4], id[4:6], id[6:8], id[8:10], id[10:])
	if err := os.MkdirAll(s.directory, 0755); err != nil {
		return "", nil, fmt.Errorf("create avatar directory: %w", err)
	}
	path := filepath.Join(s.directory, name)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", nil, fmt.Errorf("create avatar: %w", err)
	}
	cleanup = func() error { return os.Remove(path) }
	_, writeErr := io.Copy(file, &encoded)
	err = errors.Join(writeErr, file.Close())
	if err != nil {
		return "", nil, errors.Join(fmt.Errorf("write avatar: %w", err), cleanup())
	}
	return AvatarURLPrefix + name, cleanup, nil
}
