package auth

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
)

type AdminUser struct {
	ID           int64
	Username     string
	PasswordHash string
	Status       string
}

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (AdminUser, error)
	TouchLastLogin(ctx context.Context, userID int64) error
}

type Service struct {
	repo     UserRepository
	secret   []byte
	tokenTTL time.Duration
	now      func() time.Time
}

func NewService(repo UserRepository, jwtSecret string) *Service {
	return &Service{
		repo:     repo,
		secret:   []byte(jwtSecret),
		tokenTTL: 24 * time.Hour,
		now:      time.Now,
	}
}

func (s *Service) Login(ctx context.Context, username string, password string) (string, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if user.Status != "" && user.Status != "active" {
		return "", ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	now := s.now()
	claims := AccessClaims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	_ = s.repo.TouchLastLogin(ctx, user.ID)
	return tokenString, nil
}

type StaticUserRepository struct {
	user AdminUser
}

func NewStaticUserRepository(username string, passwordHash string) *StaticUserRepository {
	return &StaticUserRepository{
		user: AdminUser{
			ID:           1,
			Username:     username,
			PasswordHash: passwordHash,
			Status:       "active",
		},
	}
}

func (r *StaticUserRepository) FindByUsername(_ context.Context, username string) (AdminUser, error) {
	if username != r.user.Username {
		return AdminUser{}, ErrInvalidCredentials
	}
	return r.user, nil
}

func (r *StaticUserRepository) TouchLastLogin(_ context.Context, _ int64) error {
	return nil
}
