package service

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"minecraft-manager/model"
	jwtpkg "minecraft-manager/pkg/jwt"
	"minecraft-manager/pkg/redis"
)

type AuthService struct {
	DB        *gorm.DB
	JWTSecret string
}

func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{DB: db, JWTSecret: jwtSecret}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	var user model.User
	if err := s.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invalid username or password")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	token, err := jwtpkg.GenerateToken(user.ID, user.Username, user.Role, s.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	return &LoginResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}, nil
}

func (s *AuthService) Logout(token string) error {
	claims, err := jwtpkg.ParseToken(token, s.JWTSecret)
	if err != nil {
		// Token is already invalid; that's fine for logout
		return nil
	}

	ttl := jwtpkg.GetRemainingTTL(claims)
	if ttl > 0 {
		return redis.BlacklistToken(token, ttl)
	}
	return nil
}

func (s *AuthService) GetUser(userID uint) (*model.User, error) {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// HashPassword generates a bcrypt hash. Used for seeding and user creation.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

