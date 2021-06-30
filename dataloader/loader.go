package dataloader

import (
	"context"
	"kontrakt-server/prisma/db"
	"net/http"
	"time"
)

const loadersKey = "dataloaders"

type Loaders struct {
	GroupsByContractID GroupsLoader
	SkillsByContractID SkillsLoader
	ContractsByGroupID ContractsLoader
	StudentsByGroupID  StudentsLoader
}

func Middleware(prismaClient *db.PrismaClient, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), loadersKey, &Loaders{
			GroupsLoader{
				fetch: func(contractIDs []int) ([][]db.GroupModel, []error) {
					contracts, err := prismaClient.Contract.FindMany(db.Contract.ID.In(contractIDs)).With(db.Contract.Groups.Fetch()).Exec(r.Context())
					groupsByContractID := map[int][]db.GroupModel{}
					for _, contract := range contracts {
						groupsByContractID[contract.ID] = contract.Groups()
					}
					groups := make([][]db.GroupModel, len(contractIDs))
					for i, contractID := range contractIDs {
						groups[i] = groupsByContractID[contractID]
					}
					return groups, []error{err}
				},
				wait:     1 * time.Millisecond,
				maxBatch: 100,
			},
			SkillsLoader{
				fetch: func(contractIDs []int) ([][]db.SkillModel, []error) {
					contracts, err := prismaClient.Contract.FindMany(db.Contract.ID.In(contractIDs)).With(db.Contract.Skills.Fetch()).Exec(r.Context())
					skillsByContractID := map[int][]db.SkillModel{}
					for _, contract := range contracts {
						skillsByContractID[contract.ID] = contract.Skills()
					}
					skills := make([][]db.SkillModel, len(contractIDs))
					for i, contractID := range contractIDs {
						skills[i] = skillsByContractID[contractID]
					}
					return skills, []error{err}
				},
				wait:     1 * time.Millisecond,
				maxBatch: 100,
			},
			ContractsLoader{
				fetch: func(groupIDs []int) ([][]db.ContractModel, []error) {
					groups, err := prismaClient.Group.FindMany(db.Group.ID.In(groupIDs)).With(db.Group.Contracts.Fetch()).Exec(r.Context())
					contractsByGroupID := map[int][]db.ContractModel{}
					for _, group := range groups {
						contractsByGroupID[group.ID] = group.Contracts()
					}
					contracts := make([][]db.ContractModel, len(groupIDs))
					for i, contractID := range groupIDs {
						contracts[i] = contractsByGroupID[contractID]
					}
					return contracts, []error{err}
				},
				wait:     1 * time.Millisecond,
				maxBatch: 100,
			},
			StudentsLoader{
				fetch: func(groupIDs []int) ([][]db.StudentModel, []error) {
					groups, err := prismaClient.Group.FindMany(db.Group.ID.In(groupIDs)).With(db.Group.Students.Fetch()).Exec(r.Context())
					studentsByGroupID := map[int][]db.StudentModel{}
					for _, group := range groups {
						studentsByGroupID[group.ID] = group.Students()
					}
					students := make([][]db.StudentModel, len(groupIDs))
					for i, contractID := range groupIDs {
						students[i] = studentsByGroupID[contractID]
					}
					return students, []error{err}
				},
				wait:     1 * time.Millisecond,
				maxBatch: 100,
			},
		})
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}
