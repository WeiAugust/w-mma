package auth

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type fakeRepo struct {
	user AdminUser
}

func (r *fakeRepo) FindByUsername(_ context.Context, username string) (AdminUser, error) {
	if username != r.user.Username {
		return AdminUser{}, errors.New("not found")
	}
	return r.user, nil
}

func (r *fakeRepo) TouchLastLogin(_ context.Context, _ int64) error {
	return nil
}

func TestLogin_ReturnsToken(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}

	svc := NewService(&fakeRepo{
		user: AdminUser{
			ID:           7,
			Username:     "admin",
			PasswordHash: string(hash),
			Status:       "active",
		},
	}, "test-secret")

	token, err := svc.Login(context.Background(), "admin", "secret")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}

	svc := NewService(&fakeRepo{
		user: AdminUser{
			ID:           7,
			Username:     "admin",
			PasswordHash: string(hash),
			Status:       "active",
		},
	}, "test-secret")

	_, err = svc.Login(context.Background(), "admin", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
