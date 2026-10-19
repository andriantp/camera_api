package ucamera

import (
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/logger"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Protocol interface {
	GetProtocols(ctx context.Context) (*response.ProtocolList, error)
	CreateProtocol(ctx context.Context, actor request.UserActor, req request.CreateProtocol) (*string, error)
}

func (u *usecase) GetProtocols(ctx context.Context) (*response.ProtocolList, error) {
	protocols, err := u.repoCamera.FindProtocols(ctx)
	if err != nil {
		logger.Error("GetProtocols-FindProtocols", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	result := make([]response.UUIDName, 0, len(protocols))

	for _, protocol := range protocols {
		result = append(result, response.UUIDName{
			UUID: protocol.UUID.String(),
			Name: protocol.Name,
		})
	}

	return &response.ProtocolList{
		Total:     int64(len(result)),
		Protocols: result,
	}, nil
}

func (u *usecase) CreateProtocol(ctx context.Context, actor request.UserActor, req request.CreateProtocol) (*string, error) {
	var protocolUUID string

	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		protocol := model.Protocol{
			Name: strings.ToUpper(req.Name),
		}

		if err := u.repoCamera.CreateProtocol(ctxWithTx, &protocol); err != nil {
			logger.Error("CreateProtocol-CreateProtocol", err.Error())
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		protocolUUID = protocol.UUID.String()

		msg := fmt.Sprintf("Created protocol %s", protocol.Name)
		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "CREATE_PROTOCOL",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("CreateProtocol-Transaction", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return &protocolUUID, nil
}
