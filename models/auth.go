package models

import "github.com/missdeer/wego/modules/utils"

// return a user salt token
func GetUserSalt() string {
	return utils.GetRandomString(10)
}
