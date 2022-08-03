package command_parser

var (
	MissingUserName = "You must specify a username for your profile in the telegram settings.\n\n" +
		"Open Telegram -> Settings -> Edit profile -> Enter username"
	AlreadyLoggedIn         = "You already logged in."
	InternalError           = "An internal error has occurred"
	LoginSuccessful         = "Login successful."
	IncorrectUsageOfCommand = "Incorrect use of the command!\n\n" +
		"You must specify only the user you want to %v access to - %v @username\n\n"
	NotLoggedIn           = "@%v not logged in."
	AlreadyHasAccess      = "@%v already has access to send you notifications."
	CanSendNotifications  = "@%v can now send you notifications"
	HaveNotAccess         = "@%v does not have access to send you notifications."
	CantSendNotifications = "@%s can no longer send you notifications"
)
