package deploy

import (
	"database/sql"
	"youfun/shipyard/internal/models"
	"youfun/shipyard/pkg/types"
	"fmt"

	"github.com/google/uuid"
)

// convertAppDTOToModel converts an ApplicationDTO to an internal Application model.
func convertAppDTOToModel(dto *types.ApplicationDTO) (*models.Application, error) {
	appID, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse application ID '%s': %w", dto.ID, err)
	}
	return &models.Application{
		ID:   appID,
		Name: dto.Name,
	}, nil
}

// convertHostDTOToModel converts an SSHHostDTO to an internal SSHHost model.
func convertHostDTOToModel(dto *types.SSHHostDTO) (*models.SSHHost, error) {
	hostID, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host ID '%s': %w", dto.ID, err)
	}
	return &models.SSHHost{
		ID:         hostID,
		Name:       dto.Name,
		Addr:       dto.Addr,
		Port:       dto.Port,
		User:       dto.User,
		Password:   dto.Password,
		PrivateKey: dto.PrivateKey,
		Status:     dto.Status,
		Arch:       dto.Arch,
	}, nil
}

// convertInstanceDTOToModel converts an ApplicationInstanceDTO to an internal ApplicationInstance model.
func convertInstanceDTOToModel(dto *types.ApplicationInstanceDTO) (*models.ApplicationInstance, error) {
	instanceID, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse instance ID '%s': %w", dto.ID, err)
	}
	appID, err := uuid.Parse(dto.ApplicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse application ID '%s': %w", dto.ApplicationID, err)
	}
	hostID, err := uuid.Parse(dto.HostID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host ID '%s': %w", dto.HostID, err)
	}
	return &models.ApplicationInstance{
		ID:                 instanceID,
		ApplicationID:      appID,
		HostID:             hostID,
		Status:             dto.Status,
		ActivePort:         sql.NullInt64{Int64: dto.ActivePort, Valid: dto.ActivePort > 0},
		PreviousActivePort: sql.NullInt64{Int64: dto.PreviousActivePort, Valid: dto.PreviousActivePort > 0},
	}, nil
}
