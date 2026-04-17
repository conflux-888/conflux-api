package adminauth

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/conflux-888/conflux-api/internal/common/jwt"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

const tokenExpiry = 1 * time.Hour

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrNotConfigured      = errors.New("admin authentication not configured")
)

type Service struct {
	user         string
	passwordHash []byte
	jwtSecret    string
}

func NewService(user string, passwordHash []byte, jwtSecret string) *Service {
	return &Service{
		user:         user,
		passwordHash: passwordHash,
		jwtSecret:    jwtSecret,
	}
}

// Configured reports whether admin credentials are set. When false, Login always returns ErrNotConfigured.
func (s *Service) Configured() bool {
	return s.user != "" && len(s.passwordHash) > 0
}

func (s *Service) Login(_ context.Context, req LoginRequest) (*LoginResponse, error) {
	if !s.Configured() {
		log.Warn().Msg("[adminauth.Login] attempted login but admin creds not configured")
		return nil, ErrNotConfigured
	}

	// Constant-time username compare to avoid user enumeration via timing.
	if subtle.ConstantTimeCompare([]byte(req.Username), []byte(s.user)) != 1 {
		log.Warn().Str("username", req.Username).Msg("[adminauth.Login] bad username")
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(s.passwordHash, []byte(req.Password)); err != nil {
		log.Warn().Str("username", req.Username).Msg("[adminauth.Login] bad password")
		return nil, ErrInvalidCredentials
	}

	token, err := jwt.GenerateAdminToken(s.user, s.jwtSecret, tokenExpiry)
	if err != nil {
		log.Error().Err(err).Msg("[adminauth.Login] failed to generate token")
		return nil, err
	}

	log.Info().Str("username", s.user).Msg("[adminauth.Login] admin authenticated")
	return &LoginResponse{
		AccessToken: token,
		ExpiresIn:   int(tokenExpiry.Seconds()),
	}, nil
}
