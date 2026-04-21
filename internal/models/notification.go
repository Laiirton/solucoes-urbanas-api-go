package models

import (
	"fmt"
	"strings"
)

type RegisterPushTokenRequest struct {
	Token string `json:"token"`
}

func (r *RegisterPushTokenRequest) Validate() error {
	r.Token = strings.TrimSpace(r.Token)
	if r.Token == "" {
		return fmt.Errorf("token is required")
	}

	if !strings.HasPrefix(r.Token, "ExponentPushToken[") && !strings.HasPrefix(r.Token, "ExpoPushToken[") {
		return fmt.Errorf("invalid push token format")
	}

	return nil
}
