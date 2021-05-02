package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"kontrakt-server/graph/generated"
	"kontrakt-server/graph/model"
	"kontrakt-server/prisma/db"
	"kontrakt-server/utils"
)

func (r *contractResolver) Skills(ctx context.Context, obj *model.Contract) ([]*model.Skill, error) {
	skills, err := r.Prisma.Skill.FindMany(db.Skill.ContractID.Equals(obj.ID)).Exec(ctx)
	if err != nil {
		return []*model.Skill{}, nil
	}
	var result []*model.Skill
	for _, skill := range skills {
		result = append(result, &model.Skill{
			ContractID: skill.ContractID,
			ID:         skill.ID,
			Name:       skill.Name,
		})
	}
	return result, nil
}

func (r *mutationResolver) Login(ctx context.Context, username string, password string) (*model.AuthPayload, error) {
	user, err := r.Prisma.User.FindUnique(db.User.Username.Equals(username)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("bad password")
	}
	token, err := utils.GetToken(user.Username)
	if err != nil {
		return nil, err
	}
	return &model.AuthPayload{
		Token: token,
		User: &model.User{
			Username: user.Username,
			Role:     model.Role(user.Role),
		},
	}, nil

}

func (r *queryResolver) Contracts(ctx context.Context, groupIds []int) ([]*model.Contract, error) {
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

func (r *queryResolver) Groups(ctx context.Context) ([]*model.Group, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Marks(ctx context.Context) ([]model.Mark, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Student(ctx context.Context, ownerUsername string) (*model.Student, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Contract(ctx context.Context, id int) (*model.Contract, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Students(ctx context.Context) ([]*model.Student, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Teachers(ctx context.Context) ([]*model.Teacher, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

// Contract returns generated.ContractResolver implementation.
func (r *Resolver) Contract() generated.ContractResolver { return &contractResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type contractResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
