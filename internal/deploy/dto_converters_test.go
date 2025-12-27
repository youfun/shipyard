package deploy

import (
	"youfun/shipyard/pkg/types"
	"testing"

	"github.com/google/uuid"
)

func TestConvertAppDTOToModel(t *testing.T) {
	tests := []struct {
		name    string
		dto     *types.ApplicationDTO
		wantErr bool
	}{
		{
			name: "Valid application DTO",
			dto: &types.ApplicationDTO{
				ID:   uuid.New().String(),
				Name: "test-app",
			},
			wantErr: false,
		},
		{
			name: "Invalid UUID",
			dto: &types.ApplicationDTO{
				ID:   "invalid-uuid",
				Name: "test-app",
			},
			wantErr: true,
		},
		{
			name: "Empty UUID",
			dto: &types.ApplicationDTO{
				ID:   "",
				Name: "test-app",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertAppDTOToModel(tt.dto)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertAppDTOToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result.Name != tt.dto.Name {
					t.Errorf("convertAppDTOToModel() Name = %v, want %v", result.Name, tt.dto.Name)
				}
				if result.ID.String() != tt.dto.ID {
					t.Errorf("convertAppDTOToModel() ID = %v, want %v", result.ID.String(), tt.dto.ID)
				}
			}
		})
	}
}

func TestConvertHostDTOToModel(t *testing.T) {
	validUUID := uuid.New().String()
	password := "secret"
	privateKey := "-----BEGIN RSA PRIVATE KEY-----..."

	tests := []struct {
		name    string
		dto     *types.SSHHostDTO
		wantErr bool
	}{
		{
			name: "Valid host DTO",
			dto: &types.SSHHostDTO{
				ID:       validUUID,
				Name:     "test-host",
				Addr:     "192.168.1.1",
				Port:     22,
				User:     "deploy",
				Password: &password,
				Status:   "active",
				Arch:     "amd64",
			},
			wantErr: false,
		},
		{
			name: "Invalid UUID",
			dto: &types.SSHHostDTO{
				ID:   "not-a-uuid",
				Name: "test-host",
			},
			wantErr: true,
		},
		{
			name: "Host with private key",
			dto: &types.SSHHostDTO{
				ID:         validUUID,
				Name:       "test-host",
				Addr:       "10.0.0.1",
				Port:       22,
				User:       "root",
				PrivateKey: &privateKey,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertHostDTOToModel(tt.dto)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertHostDTOToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result.Name != tt.dto.Name {
					t.Errorf("convertHostDTOToModel() Name = %v, want %v", result.Name, tt.dto.Name)
				}
				if result.Addr != tt.dto.Addr {
					t.Errorf("convertHostDTOToModel() Addr = %v, want %v", result.Addr, tt.dto.Addr)
				}
				if result.Port != tt.dto.Port {
					t.Errorf("convertHostDTOToModel() Port = %v, want %v", result.Port, tt.dto.Port)
				}
			}
		})
	}
}

func TestConvertInstanceDTOToModel(t *testing.T) {
	appID := uuid.New().String()
	hostID := uuid.New().String()
	instanceID := uuid.New().String()

	tests := []struct {
		name    string
		dto     *types.ApplicationInstanceDTO
		wantErr bool
	}{
		{
			name: "Valid instance DTO with ports",
			dto: &types.ApplicationInstanceDTO{
				ID:                 instanceID,
				ApplicationID:      appID,
				HostID:             hostID,
				Status:             "running",
				ActivePort:         4000,
				PreviousActivePort: 4001,
			},
			wantErr: false,
		},
		{
			name: "Valid instance DTO without ports",
			dto: &types.ApplicationInstanceDTO{
				ID:            instanceID,
				ApplicationID: appID,
				HostID:        hostID,
				Status:        "stopped",
				ActivePort:    0,
			},
			wantErr: false,
		},
		{
			name: "Invalid instance ID",
			dto: &types.ApplicationInstanceDTO{
				ID:            "invalid",
				ApplicationID: appID,
				HostID:        hostID,
			},
			wantErr: true,
		},
		{
			name: "Invalid application ID",
			dto: &types.ApplicationInstanceDTO{
				ID:            instanceID,
				ApplicationID: "invalid",
				HostID:        hostID,
			},
			wantErr: true,
		},
		{
			name: "Invalid host ID",
			dto: &types.ApplicationInstanceDTO{
				ID:            instanceID,
				ApplicationID: appID,
				HostID:        "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInstanceDTOToModel(tt.dto)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertInstanceDTOToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result.Status != tt.dto.Status {
					t.Errorf("convertInstanceDTOToModel() Status = %v, want %v", result.Status, tt.dto.Status)
				}
				// Check ActivePort valid flag
				if tt.dto.ActivePort > 0 && !result.ActivePort.Valid {
					t.Errorf("convertInstanceDTOToModel() ActivePort should be valid when > 0")
				}
				if tt.dto.ActivePort == 0 && result.ActivePort.Valid {
					t.Errorf("convertInstanceDTOToModel() ActivePort should be invalid when 0")
				}
			}
		})
	}
}
