package sync

import (
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

// HandleGetStatus godoc
// @Summary      Get GDELT sync status
// @Tags         admin
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /admin/sync/status [get]
func (h *Handler) HandleGetStatus(c *gin.Context) {
	state, err := h.svc.GetStatus(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("[sync.HandleGetStatus] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, state)
}

// HandleTriggerSync godoc
// @Summary      Trigger GDELT sync manually
// @Description  Runs a full GDELT sync cycle synchronously and returns the result
// @Tags         admin
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /admin/sync/trigger [post]
func (h *Handler) HandleTriggerSync(c *gin.Context) {
	h.svc.TriggerSync(c.Request.Context())

	state, err := h.svc.GetStatus(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("[sync.HandleTriggerSync] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, state)
}
