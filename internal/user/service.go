package user

import (
	"context"
	"errors"
	"time"

	"github.com/conflux-888/conflux-api/internal/common/jwt"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken      = errors.New("email already taken")
	ErrInvalidPassword = errors.New("invalid email or password")
)

const (
	bcryptCost  = 12
	tokenExpiry = 1 * time.Hour
)

type Service struct {
	repo      *Repository
	jwtSecret string
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*ProfileResponse, error) {
	existing, _ := s.repo.FindByEmail(ctx, req.Email)
	if existing != nil {
		log.Warn().Str("email", req.Email).Msg("[user.Register] email already taken")
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		log.Error().Err(err).Msg("[user.Register] failed to hash password")
		return nil, err
	}

	u := &User{
		Email:        req.Email,
		PasswordHash: string(hash),
		DisplayName:  req.DisplayName,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[user.Register] failed to create user")
		return nil, err
	}

	log.Info().Str("email", req.Email).Str("user_id", u.ID.Hex()).Msg("[user.Register] user registered")
	return u.ToProfileResponse(), nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if errors.Is(err, ErrNotFound) {
		log.Warn().Str("email", req.Email).Msg("[user.Login] user not found")
		return nil, ErrInvalidPassword
	}
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("[user.Login] failed to find user")
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		log.Warn().Str("email", req.Email).Msg("[user.Login] invalid password")
		return nil, ErrInvalidPassword
	}

	token, err := jwt.GenerateToken(u.ID.Hex(), u.Email, s.jwtSecret, tokenExpiry)
	if err != nil {
		log.Error().Err(err).Str("user_id", u.ID.Hex()).Msg("[user.Login] failed to generate token")
		return nil, err
	}

	log.Info().Str("user_id", u.ID.Hex()).Msg("[user.Login] login successful")
	return &LoginResponse{
		AccessToken: token,
		ExpiresIn:   int(tokenExpiry.Seconds()),
	}, nil
}

func (s *Service) GetProfile(ctx context.Context, userID string) (*ProfileResponse, error) {
	id, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		log.Warn().Str("user_id", userID).Msg("[user.GetProfile] invalid user id")
		return nil, ErrNotFound
	}

	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("[user.GetProfile] failed to find user")
		return nil, err
	}

	return u.ToProfileResponse(), nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*ProfileResponse, error) {
	id, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		log.Warn().Str("user_id", userID).Msg("[user.UpdateProfile] invalid user id")
		return nil, ErrNotFound
	}

	u, err := s.repo.Update(ctx, id, bson.M{
		"display_name": req.DisplayName,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("[user.UpdateProfile] failed to update user")
		return nil, err
	}

	log.Info().Str("user_id", userID).Msg("[user.UpdateProfile] profile updated")
	return u.ToProfileResponse(), nil
}
