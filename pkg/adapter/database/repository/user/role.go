package user

import (
	"context"

	"camera_api/pkg/adapter/database/model"
)

type Role interface {
	FindRoleByName(ctx context.Context, roleName string) (*model.Role, error)
	FindRoleByUUID(ctx context.Context, roleUUID string) (*model.Role, error)
	FindRoles(ctx context.Context, actorRoleLevel int) ([]model.Role, error)
}

func (r *repo) FindRoles(ctx context.Context, actorRoleLevel int) ([]model.Role, error) {
	db := r.provider.Db(ctx)
	var roles []model.Role

	err := db.
		Where("deleted_at IS NULL").
		Where("id > ?", actorRoleLevel).
		Order("id ASC").
		Find(&roles).
		Error

	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *repo) FindRoleByName(ctx context.Context, roleName string) (*model.Role, error) {
	db := r.provider.Db(ctx)
	var role model.Role

	err := db.
		Where("name = ? AND deleted_at IS NULL", roleName).
		First(&role).
		Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *repo) FindRoleByUUID(ctx context.Context, roleUUID string) (*model.Role, error) {
	db := r.provider.Db(ctx)
	var role model.Role

	err := db.
		Where("uuid = ? AND deleted_at IS NULL", roleUUID).
		First(&role).
		Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}
