package auth

import (
	"context"
	"doligo_001/internal/api/auth"
	"doligo_001/internal/domain/identity"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthUseCase struct {
	userRepo identity.UserRepository
	roleRepo identity.RoleRepository
}

func NewAuthUseCase(userRepo identity.UserRepository, roleRepo identity.RoleRepository) *AuthUseCase {
	return &AuthUseCase{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (uc *AuthUseCase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	permissions, err := uc.roleRepo.GetRolePermissions(ctx, user.RoleID)
	if err != nil {
		return "", err
	}

	var permissionNames []string
	for _, p := range permissions {
		permissionNames = append(permissionNames, p.Name)
	}

	return auth.GenerateToken(user.ID, permissionNames)
}
