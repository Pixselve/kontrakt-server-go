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

func (r *queryResolver) Students(ctx context.Context) ([]db.StudentModel, error) {
	return r.Prisma.Student.FindMany().Exec(ctx)
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
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type skillResolver struct{ *Resolver }
type studentResolver struct{ *Resolver }
type studentSkillResolver struct{ *Resolver }
type teacherResolver struct{ *Resolver }
