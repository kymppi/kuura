package kuura

import (
	"context"

	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/models"
)

type M2MService struct {
	db *db_gen.Queries
}

func NewM2MService(generatedQueries *db_gen.Queries) *M2MService {
	return &M2MService{
		db: generatedQueries,
	}
}

func (s *M2MService) CreateRoleTemplate(ctx context.Context, name string, roles []string) error {
	err := s.db.CreateM2MRoleTemplate(ctx, db_gen.CreateM2MRoleTemplateParams{
		ID:    name,
		Roles: roles,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *M2MService) GetRoleTemplates(ctx context.Context) ([]*models.M2MRoleTemplate, error) {
	data, err := s.db.GetM2MRoleTemplates(ctx)

	if err != nil {
		return nil, err
	}

	var result []*models.M2MRoleTemplate
	for _, row := range data {
		result = append(result, &models.M2MRoleTemplate{
			Id:    row.ID,
			Roles: row.Roles,
		})
	}

	return result, nil
}
