package preferences

import (
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

// HandleGet godoc
// @Summary      Get my preferences
// @Tags         preferences
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /preferences [get]
func (h *Handler) HandleGet(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	prefs, err := h.svc.Get(c.Request.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("[preferences.HandleGet] error")
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, prefs)
}

// HandleUpdate godoc
// @Summary      Update my preferences
// @Tags         preferences
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateRequest  true  "Preferences"
// @Success      200      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /preferences [put]
func (h *Handler) HandleUpdate(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	prefs, err := h.svc.Update(c.Request.Context(), userID, req)
	if err != nil {
		log.Error().Err(err).Msg("[preferences.HandleUpdate] error")
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, prefs)
}

// HandleUpdateLocation godoc
// @Summary      Update my current location
// @Tags         preferences
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateLocationRequest  true  "Location"
// @Success      200      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /preferences/location [put]
func (h *Handler) HandleUpdateLocation(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := h.svc.UpdateLocation(c.Request.Context(), userID, req.Latitude, req.Longitude); err != nil {
		log.Error().Err(err).Msg("[preferences.HandleUpdateLocation] error")
		response.InternalError(c)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "location updated"})
}
