package adminauth

import (
	"errors"
	"net/http"

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

// HandleLogin godoc
// @Summary      Admin login
// @Description  Authenticates an admin using ADMIN_USER / ADMIN_PASSWORD env credentials. Returns an admin-scoped JWT.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Admin credentials"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Router       /admin/auth/login [post]
func (h *Handler) HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req)
	if errors.Is(err, ErrInvalidCredentials) {
		response.Unauthorized(c, "invalid username or password")
		return
	}
	if errors.Is(err, ErrNotConfigured) {
		response.Error(c, http.StatusServiceUnavailable, response.ErrInternal, "admin authentication is not configured")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("[adminauth.HandleLogin] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, resp)
}
