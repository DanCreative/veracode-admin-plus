package backend

import (
	"context"
	"sync"

	"github.com/DanCreative/veracode-admin-plus/admin"
	"github.com/DanCreative/veracode-go/veracode"
)

var _ admin.UserEntityRepository[admin.Role] = &RoleRepository{}

type RoleRepository struct {
	murw       sync.RWMutex
	localRoles []admin.Role
	getClient  func() (*veracode.Client, error)
}

func NewRoleRepository(getClientFunc func() (*veracode.Client, error)) *RoleRepository {
	return &RoleRepository{
		getClient:  getClientFunc,
		localRoles: make([]admin.Role, 0),
	}
}

func (r *RoleRepository) List(ctx context.Context, options admin.PageOptions, shouldRefresh bool) ([]admin.Role, error) {
	client, err := r.getClient()
	if err != nil {
		return nil, err
	}

	if shouldRefresh || len(r.localRoles) < 1 {
		vroles, _, err := client.Identity.ListRoles(ctx, veracode.PageOptions(options))
		if err != nil {
			return nil, err
		}

		droles := make([]admin.Role, 0, len(vroles))

		for _, vrole := range vroles {
			droles = append(droles, admin.Role{
				RoleId:          vrole.RoleId,
				RoleName:        vrole.RoleName,
				RoleDescription: vrole.RoleDescription,
				IsApi:           vrole.IsApi,
				IsScanType:      vrole.IsScanType,
			})
		}
		r.murw.Lock()
		r.localRoles = droles
		r.murw.Unlock()

		return droles, nil
	} else {
		r.murw.RLock()
		defer r.murw.RUnlock()
		return r.localRoles, nil
	}
}
