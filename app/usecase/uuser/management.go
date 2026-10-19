package uuser

import (
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/logger"
	"camera_api/pkg/utils/password"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type Management interface {
	GetUsers(ctx context.Context, actor request.UserActor, limit, offset int) (*response.UserList, error)
	GetRoles(ctx context.Context, actor request.UserActor) (*response.RoleList, error)
	CreateUser(ctx context.Context, actor request.UserActor, body request.CreateUser) (*string, error)
	GetUserProfile(ctx context.Context, actor request.UserActor, userUUID string) (*response.UserSummary, error)
	UpdateUserProfile(ctx context.Context, actor request.UserActor, userUUID string, body request.UpdateProfile) error
	ResetPassword(ctx context.Context, actor request.UserActor, userUUID string, body request.ResetPassword) error
	UpdateUserStatus(ctx context.Context, actor request.UserActor, userUUID string, body request.UpdateUserStatus) error
	DeleteUser(ctx context.Context, actor request.UserActor, userUUID string) error
}

func (u *usecase) GetUsers(ctx context.Context, actor request.UserActor, limit, offset int) (*response.UserList, error) {
	actorRole, err := u.repoUser.FindRoleByName(ctx, actor.Role)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Role not found",
		}
	}

	users, err := u.repoUser.FindUsers(ctx, int(actorRole.ID), limit, offset)
	if err != nil {
		logger.Error("GetUsers-FindUsers", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	result := make([]response.UserSummary, 0, len(users))

	for _, user := range users {
		item := response.UserSummary{
			UUID:      user.UUID,
			Username:  user.Username,
			Role:      user.Role,
			LastLogin: user.LastLogin,
			LastSeen:  user.LastSeen,
			Active:    user.Active,
			CanUpdate: int64(user.RoleID) > actorRole.ID,
		}

		if user.LastActivity != nil {
			item.LastActivity = &response.LastActivity{
				At:       user.LastActivity,
				Activity: user.LastActivityType,
				Message:  user.LastActivityMessage,
			}
		}

		result = append(result, item)
	}

	return &response.UserList{
		Total:    users[0].Total,
		Returned: len(result),
		Users:    result,
	}, nil
}

func (u *usecase) GetRoles(ctx context.Context, actor request.UserActor) (*response.RoleList, error) {
	actorRole, err := u.repoUser.FindRoleByName(ctx, actor.Role)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Role not found",
		}
	}

	roles, err := u.repoUser.FindRoles(ctx, int(actorRole.ID))
	if err != nil {
		logger.Error("GetRoles-FindRoles", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	result := make([]response.UUIDName, 0, len(roles))

	for _, role := range roles {
		result = append(result, response.UUIDName{
			UUID: role.UUID.String(),
			Name: role.Name,
		})
	}

	return &response.RoleList{
		Total: int64(len(result)),
		Roles: result,
	}, nil
}

func (u *usecase) CreateUser(ctx context.Context, actor request.UserActor, body request.CreateUser) (*string, error) {
	_, err := u.repoUser.FindUserByUserName(ctx, body.Username)
	if err == nil {
		return nil, response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "User already exists",
		}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to check user",
		}
	}

	// Find actor role.
	actorRole, err := u.repoUser.FindRoleByName(ctx, actor.Role)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Actor role not found",
		}
	}
	if err != nil {
		logger.Error("CreateUser-FindActorRole", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Find target role.
	targetRole, err := u.repoUser.FindRoleByUUID(ctx, body.Role)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Role not found",
		}
	}
	if err != nil {
		logger.Error("CreateUser-FindTargetRole", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Actor can only create users with lower role hierarchy.
	if targetRole.ID <= actorRole.ID {
		return nil, response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Cannot create user with this role",
		}
	}

	// Hash password.
	hashedPassword, err := password.CreatePassword(body.Password)
	if err != nil {
		logger.Error("CreateUser-CreatePassword", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create password",
		}
	}

	now := time.Now().UTC()

	user := model.User{
		Username:    body.Username,
		Password:    hashedPassword,
		RoleID:      targetRole.UUID.String(),
		Active:      true,
		LoginFailed: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	var targetUser string
	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		if err := u.repoUser.CreateUser(ctxWithTx, &user); err != nil {
			return err
		}

		// Update actor session.
		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		targetUser = user.UUID.String()
		msg := fmt.Sprintf("Created user %s with role %s", user.Username, targetRole.Name)
		activity := model.UserActivity{
			ActorUserID:  actor.UserID,
			TargetUserID: &targetUser,
			SessionID:    &actor.SessionID,
			IPAddress:    actor.IPAddress,
			UserAgent:    actor.UserAgent,
			Activity:     "CREATE_USER",
			Message:      &msg,
			CreatedAt:    now,
		}
		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("CreateUser-Transaction", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return &targetUser, nil
}

func (u *usecase) GetUserProfile(ctx context.Context, actor request.UserActor, userUUID string) (*response.UserSummary, error) {
	actorRole, err := u.repoUser.FindRoleByName(ctx, actor.Role)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Role not found",
		}
	}

	user, err := u.repoUser.FindUserSummaryByUserUUID(ctx, userUUID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "User Not Found",
		}
	}
	if err != nil {
		logger.Error("GetUserProfile-FindUserSummaryByUserUUID", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return &response.UserSummary{
		UUID:      user.UUID,
		Username:  user.Username,
		Role:      user.Role,
		LastLogin: user.LastLogin,
		LastSeen:  user.LastSeen,
		Active:    user.Active,
		CanUpdate: int64(user.RoleID) > actorRole.ID,
	}, nil
}

func (u *usecase) UpdateUserProfile(ctx context.Context, actor request.UserActor, userUUID string, body request.UpdateProfile) error {
	// Find actor role.
	actorRole, err := u.repoUser.FindRoleByName(ctx, actor.Role)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Actor role not found",
		}
	}
	if err != nil {
		logger.Error("UpdateUserProfile-FindActorRole", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Find target user.
	targetUser, err := u.repoUser.FindUserByUserUUIDForUpdate(ctx, userUUID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "User Not Found",
		}
	}
	if err != nil {
		logger.Error("UpdateUserProfile-FindUser", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Find target role.
	targetRole, err := u.repoUser.FindRoleByUUID(ctx, targetUser.RoleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Target role not found",
		}
	}
	if err != nil {
		logger.Error("UpdateUserProfile-FindTargetRole", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Actor can only update users with lower privilege.
	if targetRole.ID <= actorRole.ID {
		return response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Can not update this user",
		}
	}

	// If role is being changed, validate the new role.
	var newRole *model.Role

	if body.Role != nil {
		newRole, err = u.repoUser.FindRoleByUUID(ctx, *body.Role)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Role not found",
			}
		}
		if err != nil {
			logger.Error("UpdateUserProfile-FindNewRole", err.Error())

			return response.APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Internal server error",
			}
		}

		// Actor cannot assign a role equal to or higher than themselves.
		if newRole.ID <= actorRole.ID {
			return response.APIError{
				StatusCode: http.StatusForbidden,
				Message:    "Cannot assign this role",
			}
		}
	}

	now := time.Now().UTC()

	oldUsername := targetUser.Username
	oldRoleID := targetUser.RoleID

	msg := "Updated"
	if body.Username != nil {
		targetUser.Username = *body.Username
		msg += fmt.Sprintf(" username to %s", targetUser.Username)
	}

	if newRole != nil {
		targetUser.RoleID = newRole.UUID.String()
		msg += fmt.Sprintf(" role to %s", newRole.Name)
	}

	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		if err := u.repoUser.UpdateUser(ctxWithTx, targetUser); err != nil {
			return err
		}

		// Update actor session.
		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		targetUserUUID := targetUser.UUID.String()
		activity := model.UserActivity{
			ActorUserID:  actor.UserID,
			TargetUserID: &targetUserUUID,
			SessionID:    &actor.SessionID,
			IPAddress:    actor.IPAddress,
			UserAgent:    actor.UserAgent,
			Activity:     "UPDATE_USER",
			Message:      &msg,
			CreatedAt:    now,
		}

		// Optional detail for role/username changes.
		_ = oldUsername
		_ = oldRoleID

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("UpdateUserProfile-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}

func (u *usecase) ResetPassword(ctx context.Context, actor request.UserActor, userUUID string, body request.ResetPassword) error {
	// Find actor role.
	actorRole, err := u.repoUser.FindRoleByName(ctx, actor.Role)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Actor role not found",
		}
	}
	if err != nil {
		logger.Error("ResetPassword-FindActorRole", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Find target user.
	targetUser, err := u.repoUser.FindUserByUserUUIDForUpdate(ctx, userUUID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "User Not Found",
		}
	}
	if err != nil {
		logger.Error("ResetPassword-FindUser", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Find target role.
	targetRole, err := u.repoUser.FindRoleByUUID(ctx, targetUser.RoleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Target role not found",
		}
	}
	if err != nil {
		logger.Error("ResetPassword-FindTargetRole", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Actor can only manage users with lower privilege.
	if targetRole.ID <= actorRole.ID {
		return response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Cannot reset password for this user",
		}
	}

	// Hash new password.
	hashedPassword, err := password.CreatePassword(body.Password)
	if err != nil {
		logger.Error("ResetPassword-CreatePassword", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create password",
		}
	}

	now := time.Now().UTC()

	targetUser.Password = hashedPassword
	targetUser.LoginFailed = 0
	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		if err := u.repoUser.UpdateUser(ctxWithTx, targetUser); err != nil {
			return err
		}

		// Update actor session.
		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		targetUserUUID := targetUser.UUID.String()
		msg := fmt.Sprintf("Reset password for user %s", targetUser.Username)
		activity := model.UserActivity{
			ActorUserID:  actor.UserID,
			TargetUserID: &targetUserUUID,
			SessionID:    &actor.SessionID,
			IPAddress:    actor.IPAddress,
			UserAgent:    actor.UserAgent,
			Activity:     "RESET_PASSWORD",
			Message:      &msg,
			CreatedAt:    now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("ResetPassword-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}

func (u *usecase) UpdateUserStatus(ctx context.Context, actor request.UserActor, userUUID string, body request.UpdateUserStatus) error {
	// Find actor role.
	actorRole, err := u.repoUser.FindRoleByName(ctx, actor.Role)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Actor role not found",
		}
	}
	if err != nil {
		logger.Error("UpdateUserStatus-FindActorRole", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Find target user.
	targetUser, err := u.repoUser.FindUserByUserUUIDForUpdate(ctx, userUUID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "User Not Found",
		}
	}
	if err != nil {
		logger.Error("UpdateUserStatus-FindUser", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Find target role.
	targetRole, err := u.repoUser.FindRoleByUUID(ctx, targetUser.RoleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Target role not found",
		}
	}
	if err != nil {
		logger.Error("UpdateUserStatus-FindTargetRole", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Actor can only manage users with lower privilege.
	if targetRole.ID <= actorRole.ID {
		return response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Cannot update this user",
		}
	}

	// Avoid unnecessary update.
	if targetUser.Active == *body.Active {
		return response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "User status is already up to date",
		}
	}

	now := time.Now().UTC()

	targetUser.Active = *body.Active
	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		// Update target status user.
		if err := u.repoUser.UpdateStatus(ctxWithTx, targetUser.UUID.String(), targetUser.Active); err != nil {
			return err
		}
		// Update actor session last seen.
		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		targetUserUUID := targetUser.UUID.String()

		action := "Activated"
		activity := "ACTIVATE_USER"

		if !*body.Active {
			action = "Deactivated"
			activity = "DEACTIVATE_USER"
		}

		msg := fmt.Sprintf("%s user %s", action, targetUser.Username)
		userActivity := model.UserActivity{
			ActorUserID:  actor.UserID,
			TargetUserID: &targetUserUUID,
			SessionID:    &actor.SessionID,
			IPAddress:    actor.IPAddress,
			UserAgent:    actor.UserAgent,
			Activity:     activity,
			Message:      &msg,
			CreatedAt:    now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &userActivity)
	})
	if err != nil {
		logger.Error("UpdateUserStatus-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}

func (u *usecase) DeleteUser(ctx context.Context, actor request.UserActor, userUUID string) error {
	// Find actor role.
	actorRole, err := u.repoUser.FindRoleByName(ctx, actor.Role)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Actor role not found",
		}
	}
	if err != nil {
		logger.Error("DeleteUser-FindActorRole", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Find target user.
	targetUser, err := u.repoUser.FindUserByUserUUIDForUpdate(ctx, userUUID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "User Not Found",
		}
	}
	if err != nil {
		logger.Error("DeleteUser-FindUserByUserUUIDForUpdate", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Find target role.
	targetRole, err := u.repoUser.FindRoleByUUID(ctx, targetUser.RoleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Target role not found",
		}
	}
	if err != nil {
		logger.Error("DeleteUser-FindTargetRole", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	// Actor can only delete users with lower privilege.
	if targetRole.ID <= actorRole.ID {
		return response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Cannot delete this user",
		}
	}

	now := time.Now().UTC()

	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		// Soft delete target user.
		if err := u.repoUser.DeleteUser(ctxWithTx, targetUser.UUID.String()); err != nil {
			return err
		}

		// Update actor session last seen.
		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		targetUserUUID := targetUser.UUID.String()
		msg := fmt.Sprintf("Deleted user %s", targetUser.Username)
		activity := model.UserActivity{
			ActorUserID:  actor.UserID,
			TargetUserID: &targetUserUUID,
			SessionID:    &actor.SessionID,
			IPAddress:    actor.IPAddress,
			UserAgent:    actor.UserAgent,
			Activity:     "DELETE_USER",
			Message:      &msg,
			CreatedAt:    now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})
	if err != nil {
		logger.Error("DeleteUser-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}
