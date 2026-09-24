package usecase

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/ran-jita/task-api/internal/entity"
	"github.com/ran-jita/task-api/internal/repository"
)

type authUsecase struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

func NewAuthUsecase(userRepo repository.UserRepository, jwtSecret []byte) AuthUsecase {
	return &authUsecase{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (u *authUsecase) Register(ctx context.Context, input RegisterInput) (*entity.User, error) {
	existing, err := u.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyUsed
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hashed),
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *authUsecase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	user, err := u.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredential
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredential
	}

	token, err := u.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{Token: token, User: *user}, nil
}

func (u *authUsecase) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": timeNowAddHours(24),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(u.jwtSecret)
}
