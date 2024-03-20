package model

type Campaign struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SignupLink string `json:"sign_up_link"`
}
