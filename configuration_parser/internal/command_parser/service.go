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

func (s *Service) Start(userId int64, userName string) (string, error) {
	if len(userName) < 0 {
		return MissingUserName, nil
	}

	err := s.repo.InsertUser(userId, userName)
	if err != nil {
		// TODO check isActive
		switch err {
		case repository.ErrAlreadyExists:
			return AlreadyLoggedIn, nil
		default:
			return InternalError, fmt.Errorf("failed to insert user in database, %v", err)
		}
	}
	return LoginSuccessful, nil
}

func (s *Service) GrantAccess(userId int64, request string) (string, error) {
	tokens := strings.Split(request, " ")
	// TODO parse multiply usernames
	if len(tokens) != 2 || tokens[1][0] != '@' {
		return fmt.Sprintf(IncorrectUsageOfCommand, "give", "/grant_access"), nil
	}
	userNameWithAccess := tokens[1][1:] //remove @ from username

	_, err := s.repo.GetUser(userNameWithAccess)
	// TODO check isActive
	if err != nil {
		return fmt.Sprintf(NotLoggedIn, userNameWithAccess), nil
	}

	err = s.repo.AddNotificationAccess(userId, userNameWithAccess)
	if err != nil {
		switch err {
		case repository.ErrAlreadyExists:
			return fmt.Sprintf(AlreadyHasAccess, userNameWithAccess), nil
		default:
			return InternalError, fmt.Errorf("failed to grant access to the user %v, %v",
				userNameWithAccess, err)
		}
	}

	return fmt.Sprintf(CanSendNotifications, userNameWithAccess), nil
}

func (s *Service) RemoveAccess(userId int64, request string) (string, error) {
	tokens := strings.Split(request, " ")
	// TODO parse multiply usernames
	if len(tokens) > 2 || tokens[1][0] != '@' {
		return fmt.Sprintf(IncorrectUsageOfCommand, "deny", "/remove_access"), nil
	}
	userNameWithAccess := tokens[1][1:] //remove @ from username

	_, err := s.repo.GetUser(userNameWithAccess)
	// TODO check isActive
	if err != nil {
		return fmt.Sprintf(NotLoggedIn, userNameWithAccess), nil
	}

	err = s.repo.RemoveNotificationAccess(userId, userNameWithAccess)
	if err != nil {
		switch err {
		case repository.ErrNotExists:
			return fmt.Sprintf(HaveNotAccess, userNameWithAccess), nil
		default:
			return InternalError, fmt.Errorf("failed to remove access from user %v, %v",
				userNameWithAccess, err)
		}
	}

	return fmt.Sprintf(CantSendNotifications, userNameWithAccess), nil
}
