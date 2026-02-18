package ufc

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	stdhtml "html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type ImageMirror interface {
	MirrorImage(ctx context.Context, rawURL string) (string, error)
}

type passthroughImageMirror struct{}

func (p passthroughImageMirror) MirrorImage(_ context.Context, rawURL string) (string, error) {
	return strings.TrimSpace(rawURL), nil
}

type LocalImageMirror struct {
	client       *http.Client
	storageDir   string
	publicBase   string
	publicPrefix string
}

type LocalImageMirrorConfig struct {
	StorageDir   string
	PublicBase   string
	PublicPrefix string
	Client       *http.Client
}

func NewLocalImageMirror(cfg LocalImageMirrorConfig) *LocalImageMirror {
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	prefix := strings.TrimSpace(cfg.PublicPrefix)
	if prefix == "" {
		prefix = "/media-cache/ufc"
	}
	return &LocalImageMirror{
		client:       client,
		storageDir:   cfg.StorageDir,
		publicBase:   strings.TrimRight(strings.TrimSpace(cfg.PublicBase), "/"),
		publicPrefix: "/" + strings.Trim(strings.TrimSpace(prefix), "/"),
	}
}

func (m *LocalImageMirror) MirrorImage(ctx context.Context, rawURL string) (string, error) {
	target := strings.TrimSpace(rawURL)
	if target == "" {
		return "", nil
	}
	if m.storageDir == "" || m.publicBase == "" {
		return "", nil
	}

	parsed, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", nil
	}

	hash := sha1.Sum([]byte(target))
	ext := imageExtension(parsed.Path)
	fileName := hex.EncodeToString(hash[:]) + ext
	relativePath := path.Join(strings.TrimPrefix(m.publicPrefix, "/"), fileName)
	localPath := filepath.Join(m.storageDir, fileName)

	if _, err := os.Stat(localPath); err == nil {
		return m.publicBase + "/" + relativePath, nil
	}

	if err := os.MkdirAll(m.storageDir, 0o755); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; w-mma-image-mirror/1.0)")
	req.Header.Set("Accept", "image/*,*/*;q=0.8")

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("mirror fetch failed with status %s", resp.Status)
	}

	tmpPath := localPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(file, io.LimitReader(resp.Body, 8<<20))
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return "", closeErr
	}
	if err := os.Rename(tmpPath, localPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	return m.publicBase + "/" + relativePath, nil
}

func imageExtension(rawPath string) string {
	clean := strings.TrimSpace(stdhtml.UnescapeString(rawPath))
	ext := strings.ToLower(filepath.Ext(clean))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".avif":
		return ext
	default:
		return ".jpg"
	}
}
