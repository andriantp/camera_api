package camera

import (
	"camera_api/pkg/adapter/database/model"
	"context"
)

type Protocol interface {
	FindProtocols(ctx context.Context) ([]model.Protocol, error)
	CreateProtocol(ctx context.Context, protocol *model.Protocol) error
	FindProtocolByUUID(ctx context.Context, protocolUUID string) (*model.Protocol, error)
}

func (r *repo) FindProtocols(ctx context.Context) ([]model.Protocol, error) {
	db := r.provider.Db(ctx)
	var protocols []model.Protocol

	err := db.
		Where("deleted_at IS NULL").
		Order("id ASC").
		Find(&protocols).
		Error

	if err != nil {
		return nil, err
	}

	return protocols, nil
}

func (r *repo) CreateProtocol(ctx context.Context, protocol *model.Protocol) error {
	return r.provider.Db(ctx).Create(protocol).Error
}

func (r *repo) FindProtocolByUUID(ctx context.Context, protocolUUID string) (*model.Protocol, error) {
	var protocol model.Protocol

	err := r.provider.Db(ctx).
		Where("uuid = ? AND deleted_at IS NULL", protocolUUID).
		First(&protocol).
		Error

	if err != nil {
		return nil, err
	}

	return &protocol, nil
}
