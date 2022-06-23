package command_parser

import (
	"configuration_parser/internal/repository"
	"fmt"
	"strings"
)

type repo interface {
	InsertUser(userId int64, userName string) error
	GetUser(userName string) (int64, error)
	AddNotificationAccess(userId int64, userNameWithAccess string) error
	RemoveNotificationAccess(userId int64, userNameWithAccess string) error
}

type Service struct {
	repo repo
}

func NewService(repo repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Start(userId int64, userName string) string {
	err := s.repo.InsertUser(userId, userName)
	if err != nil {
		// TODO log
		// TODO check isActive
		switch err {
		case repository.ErrAlreadyExists:
			return "You already logged in."
		default:
			return "An internal error has occurred"
		}
	}
	return "Login successful."
}

func (s *Service) GrantAccess(userId int64, request string) string {
	tokens := strings.Split(request, " ")
	// TODO parse multiply usernames
	if len(tokens) > 2 || tokens[1][0] != '@' {
		return "Incorrect use of the command!\n\n" +
			"You must specify only the user you want to give access to - /grant_access @username\n\n"
	}
	userNameWithAccess := tokens[1][1:] //remove @ from username

	_, err := s.repo.GetUser(userNameWithAccess)
	// TODO check isActive
	if err != nil {
		return fmt.Sprintf("@%s not logged in.", userNameWithAccess)
	}

	err = s.repo.AddNotificationAccess(userId, userNameWithAccess)
	if err != nil {
		// TODO log
		switch err {
		case repository.ErrAlreadyExists:
			return fmt.Sprintf("@%s already has access to send you notifications.", userNameWithAccess)
		default:
			return "An internal error has occurred"
		}
	}

	return fmt.Sprintf("@%s can now send you notifications", userNameWithAccess)
}

func (s *Service) RemoveAccess(userId int64, request string) string {
	tokens := strings.Split(request, " ")
	// TODO parse multiply usernames
	if len(tokens) > 2 || tokens[1][0] != '@' {
		return "Incorrect use of the command!\n\n" +
			"You must specify only the user you want to deny access to - /remove_access @username\n\n"
	}
	userNameWithAccess := tokens[1][1:] //remove @ from username

	_, err := s.repo.GetUser(userNameWithAccess)
	// TODO check isActive
	if err != nil {
		return fmt.Sprintf("@%s not logged in.", userNameWithAccess)
	}

	err = s.repo.RemoveNotificationAccess(userId, userNameWithAccess)
	if err != nil {
		// TODO log
		switch err {
		case repository.ErrNotExists:
			return fmt.Sprintf("@%s does not have access to send you notifications.", userNameWithAccess)
		default:
			return "An internal error has occurred"
		}
	}

	return fmt.Sprintf("@%s can no longer send you notifications", userNameWithAccess)
}
