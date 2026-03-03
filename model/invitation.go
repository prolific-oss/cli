package model

// Invitee represents the user being invited.
type Invitee struct {
	ID    *string `json:"id"`
	Name  *string `json:"name"`
	Email string  `json:"email"`
}

// Invitation represents an invitation to collaborate on a workspace.
type Invitation struct {
	Association string  `json:"association"`
	Invitee     Invitee `json:"invitee"`
	InvitedBy   string  `json:"invited_by"`
	Status      string  `json:"status"`
	InviteLink  string  `json:"invite_link"`
	Role        string  `json:"role"`
}

// CreateInvitation is the request body for creating invitations.
type CreateInvitation struct {
	Association string   `json:"association"`
	Emails      []string `json:"emails"`
	Role        string   `json:"role"`
}
