package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"kontrakt-server/graph/generated"
	"kontrakt-server/graph/model"
	"kontrakt-server/prisma/db"
)

func (r *contractResolver) Skills(ctx context.Context, obj *model.Contract) ([]*model.Skill, error) {
	skills, err := r.Prisma.Skill.FindMany(db.Skill.ContractID.Equals(obj.ID)).Exec(ctx)
	if err != nil {
		return []*model.Skill{}, nil
	}
	var result []*model.Skill
	for _, skill := range skills {
		result = append(result, &model.Skill{
			ContractID:    skill.ContractID,
			ID:            skill.ID,
			Name:          skill.Name,
		})
	}
	return result, nil
}

func (r *queryResolver) Contracts(ctx context.Context) ([]*model.Contract, error) {
	contracts, err := r.Prisma.Contract.FindMany().Exec(ctx)
	if err != nil {
		return nil, err
	}
	var result []*model.Contract
	for _, contract := range contracts {
		result = append(result, &model.Contract{
			Archived: contract.Archived,
			End:      contract.End.String(),
			ID:       contract.ID,
			Name:     contract.Name,
			HexColor: contract.HexColor,
			Start:    contract.Start.String(),
		})
	}
	return result, err
}

// Contract returns generated.ContractResolver implementation.
func (r *Resolver) Contract() generated.ContractResolver { return &contractResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type contractResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
