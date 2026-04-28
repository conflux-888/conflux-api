package devicetoken

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
// @Summary      Register a device push token
// @Tags         device-tokens
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "Token"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /users/me/device-tokens [post]
func (h *Handler) HandleRegister(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := h.svc.Register(c.Request.Context(), userID, req); err != nil {
		log.Error().Err(err).Msg("[devicetoken.HandleRegister] error")
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "registered"})
}

// HandleUnregister godoc
// @Summary      Unregister a device push token
// @Tags         device-tokens
// @Produce      json
// @Param        token  path  string  true  "Device token"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /users/me/device-tokens/{token} [delete]
func (h *Handler) HandleUnregister(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	token := c.Param("token")
	if err := h.svc.Unregister(c.Request.Context(), userID, token); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "device token not found")
			return
		}
		log.Error().Err(err).Msg("[devicetoken.HandleUnregister] error")
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "unregistered"})
}
