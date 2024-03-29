package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"github.com/xuri/excelize/v2"
	"kontrakt-server/dataloader"
	"kontrakt-server/graph/auth"
	"kontrakt-server/graph/generated"
	"kontrakt-server/graph/model"
	"kontrakt-server/prisma/db"
	"kontrakt-server/utils"
	"strings"
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
	return dataloader.For(ctx).SkillsByContractID.Load(obj.ID)
}

func (r *contractResolver) Groups(ctx context.Context, obj *db.ContractModel) ([]db.GroupModel, error) {
	return dataloader.For(ctx).GroupsByContractID.Load(obj.ID)
}

func (r *groupResolver) Contracts(ctx context.Context, obj *db.GroupModel) ([]db.ContractModel, error) {
	return dataloader.For(ctx).ContractsByGroupID.Load(obj.ID)
}

func (r *groupResolver) Students(ctx context.Context, obj *db.GroupModel) ([]db.StudentModel, error) {
	return dataloader.For(ctx).StudentsByGroupID.Load(obj.ID)
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
	//Delete the student
	student, err := r.Prisma.Student.FindUnique(db.Student.OwnerID.Equals(ownerUsername)).Delete().Exec(ctx)
	if err != nil {
		return nil, err
	}
	//Delete the user
	_, err = r.Prisma.User.FindUnique(db.User.Username.Equals(ownerUsername)).Delete().Exec(ctx)
	if err != nil {
		return nil, err
	}
	return student, nil
}

func (r *mutationResolver) UpsertOneSkillToStudent(ctx context.Context, studentOwnerUsername string, skillID int, mark model.Mark) (*db.StudentSkillModel, error) {
	return r.Prisma.StudentSkill.UpsertOne(db.StudentSkill.StudentIDSkillID(db.StudentSkill.StudentID.Equals(studentOwnerUsername), db.StudentSkill.SkillID.Equals(skillID))).Update(db.StudentSkill.Mark.Set(db.Mark(mark))).Create(
		db.StudentSkill.Mark.Set(db.Mark(mark)),
		db.StudentSkill.Skill.Link(db.Skill.ID.Equals(skillID)),
		db.StudentSkill.Student.Link(db.Student.OwnerID.Equals(studentOwnerUsername)),
	).Exec(ctx)
}

func (r *mutationResolver) CreateOneStudent(ctx context.Context, student model.StudentInput, user model.UserInput) (*db.StudentModel, error) {
	// Create the user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	username := strings.ToLower(string(student.FirstName[0]) + student.LastName)

	createdUser, err := r.Prisma.User.CreateOne(db.User.Username.Set(username), db.User.Password.Set(string(hashedPassword)), db.User.Role.Set(db.RoleSTUDENT)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	// Create the associated student
	return r.Prisma.Student.CreateOne(db.Student.Owner.Link(db.User.Username.Equals(createdUser.Username)), db.Student.FirstName.Set(strings.Title(student.FirstName)), db.Student.LastName.Set(strings.Title(student.LastName))).Exec(ctx)
}

func (r *mutationResolver) CreateOneTeacher(ctx context.Context, username string, password string, firstName string, lastName string) (*db.TeacherModel, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	createdUser, err := r.Prisma.User.CreateOne(db.User.Username.Set(username), db.User.Password.Set(string(hashedPassword)), db.User.Role.Set(db.RoleTEACHER)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.Prisma.Teacher.CreateOne(db.Teacher.Owner.Link(db.User.Username.Equals(createdUser.Username)), db.Teacher.FirstName.Set(firstName), db.Teacher.LastName.Set(lastName)).Exec(ctx)
}

type studentAndSkill struct {
	studentOwnerUsername string
	skillID              int
}

func (r *mutationResolver) GenerateSpreadsheet(ctx context.Context) (string, error) {
	f := excelize.NewFile()

	contracts, err := r.Prisma.Contract.FindMany().With(db.Contract.Skills.Fetch().With(db.Skill.StudentSkills.Fetch().With(db.StudentSkill.Student.Fetch())), db.Contract.Groups.Fetch().With(db.Group.Students.Fetch())).Exec(ctx)
	if err != nil {
		return "", err
	}

	for _, contract := range contracts {
		f.NewSheet(contract.Name)
		f.DeleteSheet(f.GetSheetName(0))
		students := make(map[string]db.StudentModel)
		for _, groupModel := range contract.Groups() {
			for _, studentModel := range groupModel.Students() {
				students[studentModel.OwnerID] = studentModel
			}
		}
		i := 2
		f.SetCellValue(contract.Name, "A1", "Élèves")
		for _, studentModel := range students {
			axis, err := excelize.CoordinatesToCellName(1, i)
			if err != nil {
				return "", err
			}
			err = f.SetCellValue(contract.Name, axis, studentModel.FirstName+" "+studentModel.LastName)
			if err != nil {
				return "", err
			}
			i++
		}
		i = 2
		for skillIndex, skillModel := range contract.Skills() {
			axis, err := excelize.CoordinatesToCellName(skillIndex+2, 1)
			if err != nil {
				return "", err
			}
			err = f.SetCellValue(contract.Name, axis, skillModel.Name)
			if err != nil {
				return "", err
			}
			studentToStudentSkill := make(map[string]db.StudentSkillModel)

			for _, studentSkillModel := range skillModel.StudentSkills() {
				studentToStudentSkill[studentSkillModel.StudentID] = studentSkillModel
			}
			for s := range students {
				studentSkillModel, exists := studentToStudentSkill[s]
				axis, err := excelize.CoordinatesToCellName(skillIndex+2, i)
				var markData utils.MarkData
				if exists {
					markData = utils.GetMarkData(studentSkillModel.Mark)
				} else {
					markData = utils.GetMarkData(db.MarkTODO)
				}
				err = f.SetCellValue(contract.Name, axis, markData.Text)
				if err != nil {
					return "", err
				}
				style, err := f.NewStyle(markData.Style)
				err = f.SetCellStyle(contract.Name, axis, axis, style)
				if err != nil {
					return "", err
				}
				i++
			}
			i = 2

		}

	}
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return "", err
	}
	toString := b64.StdEncoding.EncodeToString(buffer.Bytes())
	return "data:application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;base64," + toString, err
}

func (r *queryResolver) Contracts(ctx context.Context, groups *model.FilterGroup) ([]db.ContractModel, error) {
	var params []db.ContractWhereParam
	if groups != nil {
		params = append(params, db.Contract.Groups.Some(db.Group.ID.In(groups.IdsIn)))
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
	var param []db.StudentWhereParam
	if contractID != nil {
		param = append(param, db.Student.Groups.Some(db.Group.Contracts.Some(db.Contract.ID.EqualsIfPresent(contractID))))
	}
	return r.Prisma.Student.FindMany(param...).Exec(ctx)
}

func (r *queryResolver) Teachers(ctx context.Context) ([]db.TeacherModel, error) {
	return r.Prisma.Teacher.FindMany().Exec(ctx)
}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	user := auth.ForContext(ctx)
	return &model.User{
		Username: user.Username,
		Role:     model.Role(user.Role),
	}, nil
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
	return dataloader.For(ctx).SkillBySkillID.Load(obj.SkillID)
}

func (r *studentSkillResolver) Student(ctx context.Context, obj *db.StudentSkillModel) (*db.StudentModel, error) {
	return dataloader.For(ctx).StudentByUsername.Load(obj.StudentID)
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

func (r *userResolver) Student(ctx context.Context, obj *model.User) ([]db.StudentModel, error) {
	return r.Prisma.Student.FindMany(db.Student.OwnerID.Equals(obj.Username)).Exec(ctx)
}

func (r *userResolver) Teacher(ctx context.Context, obj *model.User) ([]db.TeacherModel, error) {
	return r.Prisma.Teacher.FindMany(db.Teacher.OwnerID.Equals(obj.Username)).Exec(ctx)
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

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type contractResolver struct{ *Resolver }
type groupResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type skillResolver struct{ *Resolver }
type studentResolver struct{ *Resolver }
type studentSkillResolver struct{ *Resolver }
type teacherResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
