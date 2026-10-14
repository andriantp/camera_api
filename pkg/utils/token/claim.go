package token

import "camera_api/pkg/adapter/database/model"

func ClaimsFromUserSession(data *model.UserSession) *Claims {
	return &Claims{
		UserUUID: data.UserUUID,
		UserName: data.UserName,
		RoleUUID: data.RoleUUID,
		RoleName: data.RoleName,
	}
}

func ClaimsFromUserToken(data *model.UserToken) *Claims {
	return &Claims{
		UserUUID: data.UserUUID,
		UserName: data.UserName,
		RoleUUID: data.RoleUUID,
		RoleName: data.RoleName,
	}
}
