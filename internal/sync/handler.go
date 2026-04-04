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

func (h *Handler) HandleGetStatus(c *gin.Context) {
	state, err := h.svc.GetStatus(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("[sync.HandleGetStatus] unexpected error")
		response.InternalError(c)
		return
	}

	response.Success(c, http.StatusOK, state)
}

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
