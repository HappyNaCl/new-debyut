package dto

import "github.com/HappyNaCl/new-debyut/internal/genproto/users"

type MeResponse struct {
	Id                  string `json:"id"`
	FirstName           string `json:"firstName"`
	Surname             string `json:"surname"`
	DateOfBirth         string `json:"dateOfBirth"`
	PhoneNumber         string `json:"phoneNumber"`
	Gender              string `json:"gender"`
	SubscribeNewsletter bool   `json:"subscribeNewsletter"`
	AvatarUrl           string `json:"avatarUrl"`
	Role                string `json:"role"`
}

func NewMeResponse(res *users.GetMeResponse, role string) *MeResponse {
	dateOfBirth := ""
	if res.GetDateOfBirth() != nil {
		dateOfBirth = res.GetDateOfBirth().AsTime().Format("2006-01-02")
	}

	return &MeResponse{
		Id:                  res.GetId(),
		FirstName:           res.GetFirstName(),
		Surname:             res.GetSurname(),
		DateOfBirth:         dateOfBirth,
		PhoneNumber:         res.GetPhoneNumber(),
		Gender:              res.GetGender(),
		SubscribeNewsletter: res.GetSubscribeNewsletter(),
		AvatarUrl:           res.GetAvatarUrl(),
		Role:                role,
	}
}
