package mailer

import "embed"

const (
	fromName               = "Social App"
	maxRetries             = 3
	UserInvitationTemplate = "user_invitation.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) error
}
