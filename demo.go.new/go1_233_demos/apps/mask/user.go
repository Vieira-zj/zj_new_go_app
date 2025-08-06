package mask

import "strings"

type Sensitive interface {
	MaskSensitive() any
}

// impl sensitive

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

func (u User) MaskSensitive() any {
	return User{
		Name:     u.Name,
		Password: "******",
		Phone:    u.maskPhone(u.Phone),
		Email:    u.maskEmail(u.Email),
	}
}

func (User) maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-3:]
}

func (User) maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	if len(username) <= 2 {
		return email
	}

	return username[:1] + "***" + username[len(username)-1:] + "@" + parts[1]
}
