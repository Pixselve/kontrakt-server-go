package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"kontrakt-server/graph/generated"
	"kontrakt-server/graph/model"
	"kontrakt-server/prisma/db"
	"kontrakt-server/utils"
	"time"

	"github.com/prisma/prisma-client-go/runtime/transaction"
	"golang.org/x/crypto/bcrypt"
)

func (r *contractResolver) End(ctx context.Context, obj *db.ContractModel) (string, error) {
	return obj.End.String(), nil
}

func (r *contractResolver) Start(ctx context.Context, obj *db.ContractModel) (string, error) {
	return obj.Start.String(), nil
}

func (r *contractResolver) Skills(ctx context.Context, obj *db.ContractModel) ([]db.SkillModel, error) {
	return r.Prisma.Skill.FindMany(db.Skill.ContractID.Equals(obj.ID)).Exec(ctx)
}

func (r *contractResolver) Groups(ctx context.Context, obj *db.ContractModel) ([]db.GroupModel, error) {
	return r.Prisma.Group.FindMany(db.Group.Contracts.Some(db.Contract.ID.Equals(obj.ID))).Exec(ctx)
}

func (r *groupResolver) Contracts(ctx context.Context, obj *db.GroupModel) ([]db.ContractModel, error) {
	return r.Prisma.Contract.FindMany(db.Contract.Groups.Some(db.Group.ID.Equals(obj.ID))).Exec(ctx)
}

func (r *groupResolver) Students(ctx context.Context, obj *db.GroupModel) ([]db.StudentModel, error) {
	return r.Prisma.Student.FindMany(db.Student.Groups.Some(db.Group.ID.Equals(obj.ID))).Exec(ctx)
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

func (r *mutationResolver) CreateOneGroup(ctx context.Context, name string, contractID *int) (*db.GroupModel, error) {
	var param []db.GroupSetParam
	if contractID != nil {
		param = append(param, db.Group.Contracts.Link(db.Contract.ID.Equals(*contractID)))
	}
	return r.Prisma.Group.CreateOne(db.Group.Name.Set(name), param...).Exec(ctx)
}

func (r *mutationResolver) UpdateOneContract(ctx context.Context, contractID int, groupIDs []int) (*db.ContractModel, error) {
	toLink, err := r.Prisma.Group.FindMany(db.Group.ID.In(groupIDs), db.Group.Not(db.Group.Contracts.Some(db.Contract.ID.Equals(contractID)))).Exec(ctx)
	if err != nil {
		return nil, err
	}
	toUnLink, err := r.Prisma.Group.FindMany(db.Group.Not(db.Group.ID.In(groupIDs)), db.Group.Contracts.Some(db.Contract.ID.Equals(contractID))).Exec(ctx)
	if err != nil {
		return nil, err
	}
	var transactions []transaction.Param
	for _, groupModel := range toUnLink {
		transactions = append(transactions, r.Prisma.Group.FindUnique(db.Group.ID.Equals(groupModel.ID)).Update(db.Group.Contracts.Unlink(db.Contract.ID.Equals(contractID))).Tx())
	}
	for _, groupModel := range toLink {
		transactions = append(transactions, r.Prisma.Group.FindUnique(db.Group.ID.Equals(groupModel.ID)).Update(db.Group.Contracts.Link(db.Contract.ID.Equals(contractID))).Tx())
	}
	err = r.Prisma.Prisma.Transaction(transactions...).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.Prisma.Contract.FindUnique(db.Contract.ID.Equals(contractID)).Exec(ctx)
}

func (r *mutationResolver) CreateOneSkill(ctx context.Context, name string, contractID int) (*db.SkillModel, error) {
	return r.Prisma.Skill.CreateOne(db.Skill.Name.Set(name), db.Skill.Contract.Link(db.Contract.ID.Equals(contractID))).Exec(ctx)
}

func (r *mutationResolver) DeleteOneSkill(ctx context.Context, id int) (*db.SkillModel, error) {
	return r.Prisma.Skill.FindUnique(db.Skill.ID.Equals(id)).Delete().Exec(ctx)
}

func (r *mutationResolver) UpdateOneSkill(ctx context.Context, skillID int, name *string) (*db.SkillModel, error) {
	return r.Prisma.Skill.FindUnique(db.Skill.ID.Equals(skillID)).Update(db.Skill.Name.SetIfPresent(name)).Exec(ctx)
}

func (r *mutationResolver) UpdateOneStudent(ctx context.Context, ownerUsername string, groupIDs []int) (*db.StudentModel, error) {
	toLink, err := r.Prisma.Group.FindMany(db.Group.ID.In(groupIDs), db.Group.Not(db.Group.Students.Some(db.Student.OwnerID.Equals(ownerUsername)))).Exec(ctx)
	if err != nil {
		return nil, err
	}
	toUnLink, err := r.Prisma.Group.FindMany(db.Group.Not(db.Group.ID.In(groupIDs)), db.Group.Students.Some(db.Student.OwnerID.Equals(ownerUsername))).Exec(ctx)
	if err != nil {
		return nil, err
	}
	var transactions []transaction.Param
	for _, groupModel := range toUnLink {
		transactions = append(transactions, r.Prisma.Group.FindUnique(db.Group.ID.Equals(groupModel.ID)).Update(db.Group.Students.Unlink(db.Student.OwnerID.Equals(ownerUsername))).Tx())
	}
	for _, groupModel := range toLink {
		transactions = append(transactions, r.Prisma.Group.FindUnique(db.Group.ID.Equals(groupModel.ID)).Update(db.Group.Students.Link(db.Student.OwnerID.Equals(ownerUsername))).Tx())
	}
	err = r.Prisma.Prisma.Transaction(transactions...).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.Prisma.Student.FindUnique(db.Student.OwnerID.Equals(ownerUsername)).Exec(ctx)
}

func (r *mutationResolver) CreateOneContract(ctx context.Context, end string, name string, hexColor string, start string, skillNames []string) (*db.ContractModel, error) {
	startTime, err := time.Parse("2006-01-02", start)
	if err != nil {
		return nil, err
	}
	endTime, err := time.Parse("2006-01-02", end)
	if err != nil {
		return nil, err
	}

	contract, err := r.Prisma.Contract.CreateOne(
		db.Contract.End.Set(endTime),
		db.Contract.Name.Set(name),
		db.Contract.HexColor.Set(hexColor),
		db.Contract.Start.Set(startTime),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}

	var skillsTransactions []transaction.Param
	for _, skillName := range skillNames {
		skillsTransactions = append(skillsTransactions, r.Prisma.Skill.CreateOne(db.Skill.Name.Set(skillName), db.Skill.Contract.Link(db.Contract.ID.Equals(contract.ID))).Tx())
	}
	if err := r.Prisma.Prisma.Transaction(skillsTransactions...).Exec(ctx); err != nil {
		return nil, err
	}
	return contract, nil
}

func (r *mutationResolver) DeleteOneContract(ctx context.Context, id int) (*db.ContractModel, error) {
	err := r.Prisma.Prisma.Transaction(r.Prisma.StudentSkill.FindMany(db.StudentSkill.Skill.Where(db.Skill.ContractID.Equals(id))).Delete().Tx(), r.Prisma.Skill.FindMany(db.Skill.ContractID.Equals(id)).Delete().Tx()).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.Prisma.Contract.FindUnique(db.Contract.ID.Equals(id)).Delete().Exec(ctx)
}

func (r *mutationResolver) DeleteOneStudent(ctx context.Context, ownerUsername string) (*db.StudentModel, error) {
	return r.Prisma.Student.FindUnique(db.Student.OwnerID.Equals(ownerUsername)).Delete().Exec(ctx)
}

func (r *mutationResolver) UpsertOneSkillToStudent(ctx context.Context, studentOwnerUsername string, skillID int, mark model.Mark) (*db.StudentSkillModel, error) {
	return r.Prisma.StudentSkill.UpsertOne(db.StudentSkill.StudentIDSkillID(db.StudentSkill.StudentID.Equals(studentOwnerUsername), db.StudentSkill.SkillID.Equals(skillID))).Update(db.StudentSkill.Mark.Set(db.Mark(mark))).Create(
		db.StudentSkill.Mark.Set(db.Mark(mark)),
		db.StudentSkill.Skill.Link(db.Skill.ID.Equals(skillID)),
		db.StudentSkill.Student.Link(db.Student.OwnerID.Equals(studentOwnerUsername)),
	).Exec(ctx)
}

func (r *queryResolver) Contracts(ctx context.Context, groupIds []int) ([]db.ContractModel, error) {
	var params []db.ContractWhereParam
	if len(groupIds) > 0 {
		params = append(params, db.Contract.Groups.Some(db.Group.ID.In(groupIds)))
	}
	return r.Prisma.Contract.FindMany(params...).Exec(ctx)
}

func (r *queryResolver) Groups(ctx context.Context) ([]db.GroupModel, error) {
	return r.Prisma.Group.FindMany().Exec(ctx)
}

func (r *queryResolver) Student(ctx context.Context, ownerUsername string) (*db.StudentModel, error) {
	return r.Prisma.Student.FindUnique(db.Student.OwnerID.Equals(ownerUsername)).Exec(ctx)
}

func (r *queryResolver) Contract(ctx context.Context, id int) (*db.ContractModel, error) {
	return r.Prisma.Contract.FindUnique(db.Contract.ID.Equals(id)).Exec(ctx)
}

func (r *queryResolver) Students(ctx context.Context, contractID *int) ([]db.StudentModel, error) {
	return r.Prisma.Student.FindMany(db.Student.Groups.Some(db.Group.Contracts.Some(db.Contract.ID.EqualsIfPresent(contractID)))).Exec(ctx)
}

func (r *queryResolver) Teachers(ctx context.Context) ([]db.TeacherModel, error) {
	return r.Prisma.Teacher.FindMany().Exec(ctx)
}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) StudentSkills(ctx context.Context, studentUsername string, contractID *int) ([]db.StudentSkillModel, error) {
	// Find existing studentSkills
	studentSkills, err := r.Prisma.StudentSkill.FindMany(db.StudentSkill.StudentID.Equals(studentUsername), db.StudentSkill.Skill.Where(db.Skill.ContractID.EqualsIfPresent(contractID))).Exec(ctx)
	if err != nil {
		return nil, err
	}
	// Find to do studentSkills
	todoSkills, err := r.Prisma.Skill.FindMany(db.Skill.StudentSkills.Every(db.StudentSkill.Not(db.StudentSkill.StudentID.Equals(studentUsername))), db.Skill.Contract.Where(db.Contract.ID.EqualsIfPresent(contractID), db.Contract.Groups.Some(db.Group.Students.Some(db.Student.OwnerID.Equals(studentUsername))))).Exec(ctx)
	if err != nil {
		return nil, err
	}
	for _, skill := range todoSkills {
		studentSkills = append(studentSkills, db.StudentSkillModel{
			InnerStudentSkill: db.InnerStudentSkill{
				SkillID:   skill.ID,
				StudentID: studentUsername,
				Mark:      db.MarkTODO,
			},
		})
	}
	return studentSkills, nil
}

func (r *skillResolver) StudentSkills(ctx context.Context, obj *db.SkillModel) ([]db.StudentSkillModel, error) {
	// Find existing studentSkills
	studentSkills, err := r.Prisma.StudentSkill.FindMany(db.StudentSkill.SkillID.Equals(obj.ID)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	// Find to do studentSkills
	todoStudents, err := r.Prisma.Student.FindMany(db.Student.Groups.Some(db.Group.Contracts.Some(db.Contract.ID.Equals(obj.ContractID))), db.Student.StudentSkills.Some(db.StudentSkill.Not(db.StudentSkill.SkillID.Equals(obj.ID)))).Exec(ctx)
	if err != nil {
		return nil, err
	}
	for _, student := range todoStudents {
		studentSkills = append(studentSkills, db.StudentSkillModel{
			InnerStudentSkill: db.InnerStudentSkill{
				SkillID:   obj.ID,
				StudentID: student.OwnerID,
				Mark:      db.MarkTODO,
			},
		})
	}
	return studentSkills, nil
}

func (r *studentResolver) Owner(ctx context.Context, obj *db.StudentModel) (*model.User, error) {
	user, err := r.Prisma.User.FindUnique(db.User.Username.Equals(obj.OwnerID)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &model.User{
		Username: user.Username,
		Role:     model.Role(user.Role),
	}, nil
}

func (r *studentResolver) OwnerUsername(ctx context.Context, obj *db.StudentModel) (string, error) {
	return obj.OwnerID, nil
}

func (r *studentResolver) StudentSkills(ctx context.Context, obj *db.StudentModel) ([]db.StudentSkillModel, error) {
	// Find existing studentSkills
	studentSkills, err := r.Prisma.StudentSkill.FindMany(db.StudentSkill.StudentID.Equals(obj.OwnerID)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	// Find to do studentSkills
	todoSkills, err := r.Prisma.Skill.FindMany(db.Skill.StudentSkills.Every(db.StudentSkill.Not(db.StudentSkill.StudentID.Equals(obj.OwnerID))), db.Skill.Contract.Where(db.Contract.Groups.Some(db.Group.Students.Some(db.Student.OwnerID.Equals(obj.OwnerID))))).Exec(ctx)
	if err != nil {
		return nil, err
	}
	for _, skill := range todoSkills {
		studentSkills = append(studentSkills, db.StudentSkillModel{
			InnerStudentSkill: db.InnerStudentSkill{
				SkillID:   skill.ID,
				StudentID: obj.OwnerID,
				Mark:      db.MarkTODO,
			},
		})
	}
	return studentSkills, nil
}

func (r *studentResolver) Groups(ctx context.Context, obj *db.StudentModel) ([]db.GroupModel, error) {
	return r.Prisma.Group.FindMany(db.Group.Students.Some(db.Student.OwnerID.Equals(obj.OwnerID))).Exec(ctx)
}

func (r *studentSkillResolver) Mark(ctx context.Context, obj *db.StudentSkillModel) (model.Mark, error) {
	return model.Mark(obj.Mark), nil
}

func (r *studentSkillResolver) Skill(ctx context.Context, obj *db.StudentSkillModel) (*db.SkillModel, error) {
	return r.Prisma.Skill.FindUnique(db.Skill.ID.Equals(obj.SkillID)).Exec(ctx)
}

func (r *studentSkillResolver) Student(ctx context.Context, obj *db.StudentSkillModel) (*db.StudentModel, error) {
	return r.Prisma.Student.FindUnique(db.Student.OwnerID.Equals(obj.StudentID)).Exec(ctx)
}

func (r *teacherResolver) Owner(ctx context.Context, obj *db.TeacherModel) (*model.User, error) {
	user, err := r.Prisma.User.FindUnique(db.User.Username.Equals(obj.OwnerID)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &model.User{
		Username: user.Username,
		Role:     model.Role(user.Role),
	}, nil
}

func (r *teacherResolver) OwnerUsername(ctx context.Context, obj *db.TeacherModel) (string, error) {
	return obj.OwnerID, nil
}

// Contract returns generated.ContractResolver implementation.
func (r *Resolver) Contract() generated.ContractResolver { return &contractResolver{r} }

// Group returns generated.GroupResolver implementation.
func (r *Resolver) Group() generated.GroupResolver { return &groupResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Skill returns generated.SkillResolver implementation.
func (r *Resolver) Skill() generated.SkillResolver { return &skillResolver{r} }

// Student returns generated.StudentResolver implementation.
func (r *Resolver) Student() generated.StudentResolver { return &studentResolver{r} }

// StudentSkill returns generated.StudentSkillResolver implementation.
func (r *Resolver) StudentSkill() generated.StudentSkillResolver { return &studentSkillResolver{r} }

// Teacher returns generated.TeacherResolver implementation.
func (r *Resolver) Teacher() generated.TeacherResolver { return &teacherResolver{r} }

type contractResolver struct{ *Resolver }
type groupResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type skillResolver struct{ *Resolver }
type studentResolver struct{ *Resolver }
type studentSkillResolver struct{ *Resolver }
type teacherResolver struct{ *Resolver }
