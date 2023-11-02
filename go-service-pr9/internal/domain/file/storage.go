package file

import (
	"context"
	"mime/multipart"
)

type Storage interface {
	Insert(ctx context.Context, fileData multipart.File, header *multipart.FileHeader) (string, error)
	FindOne(ctx context.Context, id string) (File, error)
	Update(ctx context.Context, name string, id string) error
	Delete(ctx context.Context, id string) error
	FindAll(ctx context.Context) ([]File, error)
}
