package addNotifications

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"notification_receiver/internal/model"

	"github.com/rs/zerolog"
)

type errorMessage struct {
	Error string `json:"errorMessage"`
}

type responseMessage struct {
	Message       string   `json:"message"`
	Authorized    []string `json:"authorizedUsers"`
	NotAuthorized []string `json:"notAuthorizedUsers"`
}

type repo interface {
	GetUser(userName string) (int64, error)
}

type publisher interface {
	Publish(notification model.Notification) error
}

type Handler struct {
	repo      repo
	logger    zerolog.Logger
	publisher publisher
}

func NewHandler(repo repo, logger zerolog.Logger, publisher publisher) *Handler {
	l := logger.With().Str("component", "add_notification_handler").Logger()
	return &Handler{
		repo:      repo,
		logger:    l,
		publisher: publisher,
	}
}

func (h *Handler) respond(w http.ResponseWriter, data interface{}, code int) {
	w.WriteHeader(code)
	if data != nil {
		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error().Msgf("failed to write response: %v", err)
		}
	}
}

func (h *Handler) AddNotification(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msgf("received a new notification: %s", r.Body)
	notification := model.Notification{}
	err := json.NewDecoder(r.Body).Decode(&notification)
	if err != nil {
		msg := fmt.Sprintf("failed to decode request %v: %v", r.Body, err)
		h.logger.Error().Msgf(msg)
		h.respond(w, errorMessage{Error: msg}, http.StatusBadRequest)
		return
	}

	// TODO check if users exist in bot database
	// TODO check if sender has access to send notifications to recipients

	var existingRecipientsId []string
	var existingRecipientsUserName []string
	var nonExistentRecipients []string

	for _, recipient := range notification.RecipientsId {
		id, err := h.repo.GetUser(recipient[1:])
		if err != nil {
			nonExistentRecipients = append(nonExistentRecipients, recipient)
			continue
		}
		existingRecipientsId = append(existingRecipientsId, strconv.FormatInt(id, 10))
		existingRecipientsUserName = append(existingRecipientsUserName, recipient)
	}

	notification.RecipientsId = existingRecipientsId

	err = h.publisher.Publish(notification)
	if err != nil {
		h.respond(w, errorMessage{Error: err.Error()}, http.StatusInternalServerError)
	}

	var response responseMessage
	response.Authorized = existingRecipientsUserName
	response.NotAuthorized = nonExistentRecipients
	if len(response.NotAuthorized) > 0 {
		response.Message = "Some users are not authorized in the telegram bot"
	} else {
		response.Message = "Notifications successfully added to the queue!"
	}
	h.respond(w, response, http.StatusOK)
}
