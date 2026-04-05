package user

import (
	"errors"
	"net/http"

	"github.com/conflux-888/conflux-api/internal/common/middleware"
	"github.com/conflux-888/conflux-api/internal/common/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// HandleRegister godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "Registration data"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      409      {object}  map[string]interface{}
// @Router       /auth/register [post]
func (h *Handler) HandleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	profile, err := h.svc.Register(c.Request.Context(), req)
	if errors.Is(err, ErrEmailTaken) {
		response.Conflict(c, "email already taken")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("[user.HandleRegister] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusCreated, profile)
}

// HandleLogin godoc
// @Summary      Login and get access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Login credentials"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Router       /auth/login [post]
func (h *Handler) HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req)
	if errors.Is(err, ErrInvalidPassword) {
		response.Unauthorized(c, "invalid email or password")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("[user.HandleLogin] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// HandleGetMe godoc
// @Summary      Get my profile
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /users/me [get]
func (h *Handler) HandleGetMe(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	profile, err := h.svc.GetProfile(c.Request.Context(), userID)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "user not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("[user.HandleGetMe] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, profile)
}

// HandleUpdateMe godoc
// @Summary      Update my profile
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateProfileRequest  true  "Profile data"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /users/me [put]
func (h *Handler) HandleUpdateMe(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	profile, err := h.svc.UpdateProfile(c.Request.Context(), userID, req)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "user not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("[user.HandleUpdateMe] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, profile)
}
