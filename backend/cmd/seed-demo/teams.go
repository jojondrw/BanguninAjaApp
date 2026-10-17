package main

import (
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
)

type teamSeed struct {
	position string
	role     string
	projects []string
}

func (s *seeder) demoTeams() map[uuid.UUID]teamSeed {
	teams := map[uuid.UUID]teamSeed{
		s.owner: {"Direktur proyek", access.RoleAdmin, []string{projectGriyaAsri, projectRuko, projectHospital, projectCluster}},
	}
	if s.manager != uuid.Nil {
		teams[s.manager] = teamSeed{"Site manager", access.RoleProjectManager, []string{projectGriyaAsri, projectHospital}}
	}
	return teams
}

func (s *seeder) seedProjectTeams() error {
	for userID, team := range s.demoTeams() {
		for _, code := range team.projects {
			projectID, known := s.ids["project:"+code]
			if !known {
				continue
			}
			keys := map[string]any{"user_id": userID, "project_id": projectID}
			_, _, err := s.ensure("project_member", keys, func() (uuid.UUID, error) {
				member, err := s.services.projects.AddMember(s.ctx, projectID, project.MemberCreateRequest{
					UserID: userID, Position: team.position, Role: team.role,
				})
				return member.ID, err
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
