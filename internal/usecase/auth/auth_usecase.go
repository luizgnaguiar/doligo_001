// Package auth contains the use case for handling authentication logic.
// It orchestrates the process of user login by coordinating with the user
// repository and handling password verification and token generation.
package auth

import (
	"context"
	"time"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"doligo_001/internal/domain/identity"
	apiMiddleware "doligo_001/internal/api/middleware"
	"doligo_001/internal/usecase"
)

// ErrInvalidCredentials is returned when the email or password is incorrect.
var ErrInvalidCredentials = errors.New("invalid credentials")

// AuthUsecase implements the business logic for authentication.
type AuthUsecase struct {
	userRepo     identity.UserRepository
	jwtSecret    []byte
	jwtTTL       time.Duration
	auditService usecase.AuditService
}

// NewAuthUsecase creates a new AuthUsecase.
func NewAuthUsecase(userRepo identity.UserRepository, jwtSecret []byte, jwtTTL time.Duration, auditService usecase.AuditService) *AuthUsecase {
	return &AuthUsecase{
		userRepo:     userRepo,
		jwtSecret:    jwtSecret,
		jwtTTL:       jwtTTL,
		auditService: auditService,
	}
}

// Login authenticates a user and returns a JWT token.
func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// In a real app, you might want to return a generic error to avoid user enumeration attacks.
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	permissions := []string{}
	for _, role := range user.Roles {
		for _, p := range role.Permissions {
			permissions = append(permissions, p.Name)
		}
	}
	
	now := time.Now()
	claims := &apiMiddleware.Claims{
		UserID:      user.ID,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(uc.jwtTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(uc.jwtSecret)
	if err != nil {
		return "", err
	}

	corrID, _ := apiMiddleware.FromContext(ctx)
	uc.auditService.Log(ctx, user.ID, "identity", user.ID.String(), "LOGIN", nil, nil, corrID)

	return tokenString, nil
}
