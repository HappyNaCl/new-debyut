package services

import (
	"context"
	"net/http"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/snowflake"
	"github.com/HappyNaCl/new-debyut/internal/common/storage"
	"go.uber.org/zap"
)

const maxAvatarBytes = 2 << 20

var avatarExtensions = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

func uploadAvatar(ctx context.Context, store *storage.Client, folder, ownerId string, content []byte) (string, error) {
	if len(content) == 0 || len(content) > maxAvatarBytes {
		return "", apperrors.ErrInvalidAvatar
	}

	contentType := http.DetectContentType(content)
	extension, ok := avatarExtensions[contentType]
	if !ok {
		return "", apperrors.ErrInvalidAvatar
	}

	key := "avatars/" + folder + "/" + ownerId + "/" + snowflake.Generate() + "." + extension
	url, err := store.Upload(ctx, key, content, contentType)
	if err != nil {
		l.Logger.Error("failed to upload avatar", zap.String("key", key), zap.Error(err))
		return "", apperrors.ErrAvatarUploadFailed
	}
	return url, nil
}

func deleteAvatar(ctx context.Context, store *storage.Client, url string) {
	if url == "" {
		return
	}
	if err := store.DeleteByUrl(ctx, url); err != nil {
		l.Logger.Warn("failed to delete old avatar", zap.String("url", url), zap.Error(err))
	}
}
