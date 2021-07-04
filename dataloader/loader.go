package dataloader

import (
	"context"
	"kontrakt-server/prisma/db"
	"net/http"
	"time"
)

const loadersKey = "dataloaders"

type Loaders struct {
	GroupsByContractID     GroupsLoader
	SkillsByContractID     SkillsLoader
	ContractsByGroupID     ContractsLoader
	StudentsByGroupID      StudentsLoader
	StudentSkillsBySkillID StudentSkillsLoader
	SkillBySkillID         SkillLoader
	StudentByUsername      StudentLoader
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
			StudentSkillsLoader{
				fetch: func(skillIDs []int) ([][]db.StudentSkillModel, []error) {
					skills, err := prismaClient.Skill.FindMany(db.Skill.ID.In(skillIDs)).With(db.Skill.StudentSkills.Fetch()).Exec(r.Context())
					studentSkillsBySkillIDs := map[int][]db.StudentSkillModel{}
					for _, skill := range skills {
						studentSkillsBySkillIDs[skill.ID] = skill.StudentSkills()
					}
					studentSkills := make([][]db.StudentSkillModel, len(skillIDs))
					for i, contractID := range skillIDs {
						studentSkills[i] = studentSkillsBySkillIDs[contractID]
					}
					return studentSkills, []error{err}
				},
				wait:     1 * time.Millisecond,
				maxBatch: 100,
			},
			SkillLoader{
				fetch: func(skillIDs []int) ([]*db.SkillModel, []error) {
					skillsToSort, err := prismaClient.Skill.FindMany(db.Skill.ID.In(skillIDs)).Exec(r.Context())
					if err != nil {
						return []*db.SkillModel{}, []error{err}
					}
					skillByID := map[int]*db.SkillModel{}
					for i, skill := range skillsToSort {
						skillByID[skill.ID] = &skillsToSort[i]
					}
					skills := make([]*db.SkillModel, len(skillIDs))
					for i, skillID := range skillIDs {
						skills[i] = skillByID[skillID]
					}
					return skills, []error{err}
				},
				wait:     1 * time.Millisecond,
				maxBatch: 100,
			},
			StudentLoader{
				fetch: func(usernames []string) ([]*db.StudentModel, []error) {
					studentsToSort, err := prismaClient.Student.FindMany(db.Student.OwnerID.In(usernames)).Exec(r.Context())
					if err != nil {
						return []*db.StudentModel{}, []error{err}
					}
					studentByUsername := map[string]*db.StudentModel{}
					for i, student := range studentsToSort {
						studentByUsername[student.OwnerID] = &studentsToSort[i]
					}
					students := make([]*db.StudentModel, len(usernames))
					for i, username := range usernames {
						students[i] = studentByUsername[username]
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
