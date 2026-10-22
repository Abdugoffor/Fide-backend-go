package application_repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	application_dto "auth-login/module/application/dto"
)

var (
	ErrVerificationNotFound = errors.New("verification record not found")
	verificationFileMutex   sync.Mutex
)

type VerificationRepository interface {
	Save(ctx context.Context, verification *application_dto.Verification) error
	Show(ctx context.Context, applicationID int64) (*application_dto.Verification, error)
	Delete(ctx context.Context, applicationID int64) error
}

type verificationRepository struct {
	path string
}

func NewVerificationRepository(path string) VerificationRepository {
	return &verificationRepository{path: path}
}

func (repository *verificationRepository) Save(ctx context.Context, verification *application_dto.Verification) error {
	verificationFileMutex.Lock()
	defer verificationFileMutex.Unlock()

	entries, err := repository.read(ctx)
	if err != nil {
		return err
	}

	for i := range entries {
		if entries[i].ApplicationID == verification.ApplicationID {
			entries[i] = *verification
			return repository.write(ctx, entries)
		}
	}

	entries = append(entries, *verification)

	return repository.write(ctx, entries)
}

func (repository *verificationRepository) Show(ctx context.Context, applicationID int64) (*application_dto.Verification, error) {
	verificationFileMutex.Lock()
	defer verificationFileMutex.Unlock()

	entries, err := repository.read(ctx)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.ApplicationID == applicationID {
			return &entry, nil
		}
	}

	return nil, ErrVerificationNotFound
}

func (repository *verificationRepository) Delete(ctx context.Context, applicationID int64) error {
	verificationFileMutex.Lock()
	defer verificationFileMutex.Unlock()

	entries, err := repository.read(ctx)
	if err != nil {
		return err
	}

	for i := range entries {
		if entries[i].ApplicationID == applicationID {
			entries = append(entries[:i], entries[i+1:]...)
			return repository.write(ctx, entries)
		}
	}

	return nil
}

func (repository *verificationRepository) read(ctx context.Context) ([]application_dto.Verification, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(repository.path)
	if errors.Is(err, os.ErrNotExist) {
		return make([]application_dto.Verification, 0), nil
	}
	if err != nil {
		return nil, err
	}

	var entries []application_dto.Verification
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func (repository *verificationRepository) write(ctx context.Context, entries []application_dto.Verification) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(repository.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	file, err := os.CreateTemp(dir, "application-codes-*.tmp")
	if err != nil {
		return err
	}

	defer os.Remove(file.Name())

	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}

	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return os.Rename(file.Name(), repository.path)
}
