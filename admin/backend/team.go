package backend

import (
	"context"
	"sync"

	"github.com/DanCreative/veracode-admin-plus/admin"
	"github.com/DanCreative/veracode-go/veracode"
)

var _ admin.UserEntityRepository[admin.Team] = &TeamRepository{}

type TeamRepository struct {
	getClient  func() (*veracode.Client, error)
	murw       sync.RWMutex
	localTeams []admin.Team
}

func NewTeamRepository(getClientFunc func() (*veracode.Client, error)) *TeamRepository {
	return &TeamRepository{
		getClient:  getClientFunc,
		localTeams: make([]admin.Team, 0),
	}
}

func (r *TeamRepository) List(ctx context.Context, options admin.PageOptions, shouldRefresh bool) ([]admin.Team, error) {
	client, err := r.getClient()
	if err != nil {
		return nil, err
	}

	if shouldRefresh || len(r.localTeams) < 1 {
		vteams, _, err := client.Identity.ListTeams(ctx, veracode.ListTeamOptions{Size: options.Size, Page: options.Page})
		if err != nil {
			return nil, err
		}

		dteams := make([]admin.Team, 0, len(vteams))

		for _, vteam := range vteams {
			dteams = append(dteams, admin.Team{
				TeamId:       vteam.TeamId,
				TeamLegacyId: vteam.TeamLegacyId,
				TeamName:     vteam.TeamName,
				Relationship: vteam.Relationship.Name,
			})
		}
		r.murw.Lock()
		r.localTeams = dteams
		r.murw.Unlock()

		return dteams, nil
	} else {
		r.murw.RLock()
		defer r.murw.RUnlock()
		return r.localTeams, nil
	}
}
