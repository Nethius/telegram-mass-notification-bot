package addNotifications

import (
	"encoding/json"
	"net/http"
	"strconv"

	"notification_receiver/internal/model"

	"github.com/rs/zerolog"
)

type repo interface {
	GetUser(userName string) (int64, error)
}

type Handler struct {
	repo          repo
	logger        zerolog.Logger
	notifications chan model.Notification
}

func NewHandler(repo repo, logger zerolog.Logger, notifications chan model.Notification) *Handler {
	l := logger.With().Str("component", "add_notification_handler").Logger()
	return &Handler{
		repo:          repo,
		logger:        l,
		notifications: notifications,
	}
}

func (h *Handler) AddNotification(w http.ResponseWriter, r *http.Request) {
	notification := model.Notification{}
	err := json.NewDecoder(r.Body).Decode(&notification)
	if err != nil {
		h.logger.Error().Msgf("failed to decode request %v: %v", r.Body, err)
		//h.respond(w, errorMessage{Error: ErrInvalidRequest.Error()}, http.StatusBadRequest)
		return
	}

	// TODO check if users exist in bot database
	// TODO check if sender has access to send notifications to recipients

	var recipientsId []string

	for _, recipient := range notification.RecipientsId {
		id, err := h.repo.GetUser(recipient[1:])
		if err != nil {
			// TODO send response
		}
		recipientsId = append(recipientsId, strconv.FormatInt(id, 10))
	}

	notification.RecipientsId = recipientsId

	h.notifications <- notification
}
