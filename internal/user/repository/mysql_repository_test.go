package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/radyatamaa/dating-apps-api/internal/domain"
	"github.com/radyatamaa/dating-apps-api/internal/user"
	"github.com/radyatamaa/dating-apps-api/pkg/helper"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"regexp"
	"testing"
	"github.com/bxcodec/faker"
)

type MysqlRepositoryTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock
}

func (t *MysqlRepositoryTestSuite) SetupSuite() {
	var (
		db  *sql.DB
		err error
	)

	// Create a new sqlmock database connection and mock object
	if db, t.mock, err = sqlmock.New(); err != nil {
		t.Fail(err.Error())
	}

	// Expect the `SELECT VERSION()` query to be called and return a mock result
	t.mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.23"))

	// Open a GORM DB connection using the mocked SQL connection
	if t.DB, err = gorm.Open(mysql.New(mysql.Config{Conn: db}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}); err != nil {
		t.Fail(err.Error())
	}
}

type fields struct {
	zapLogger zaplogger.Logger
	db        *gorm.DB
}

func (t *MysqlRepositoryTestSuite) TestNewMysqlRepository() {
	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name string
		args args
		want user.MysqlRepository
	}{
		{
			name: "success",
			args: args{
				db: t.DB,
			},
			want: NewMysqlRepository(t.DB,nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func() {
			t.Equalf(tt.want, NewMysqlRepository(tt.args.db,nil), "NewMysqlRepository(%v)", tt.args.db)
		})
	}
}

func (t *MysqlRepositoryTestSuite) TestFetchWithFilter() {
	type args struct {
		ctx       context.Context
		limit     int
		offset    int
		order     string
		fields    []string
		associate []string
		filter    []string
		model     interface{}
		args      []interface{}
	}
	tests := []struct {
		name    string
		fields  func(args *args) fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				row, fieldsDomain := helper.GetValueAndColumnStructToDriverValue(mockDomain)

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE users.id = (?) ORDER BY users.id ASC LIMIT 10")).
					WithArgs(1).WillReturnRows(
					sqlmock.NewRows(fieldsDomain).AddRow(row...))

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE users.id = (?) LIMIT 10")).
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				return fields
			},
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
				order:  "users.id ASC",
				fields: []string{
					"*",
				},
				associate: []string{},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "success associate",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				row, fieldsDomain := helper.GetValueAndColumnStructToDriverValue(mockDomain)

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` INNER JOIN profile on profile.user_id = users.id WHERE users.id = (?) ORDER BY users.id ASC LIMIT 10")).
					WithArgs(1).WillReturnRows(
					sqlmock.NewRows(fieldsDomain).AddRow(row...))

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` INNER JOIN profile on profile.user_id = users.id WHERE users.id = (?) LIMIT 10")).
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				return fields
			},
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
				order:  "users.id ASC",
				fields: []string{
					"*",
				},
				associate: []string{
					"INNER JOIN profile on profile.user_id = users.id",
				},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "get query context deadline exceeded error",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}

				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE users.id = (?) ORDER BY users.id ASC LIMIT 10")).
					WithArgs(1).WillReturnError(errors.New("context deadline exceeded"))

				return fields
			},
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
				order:  "users.id ASC",
				fields: []string{
					"*",
				},
				associate: []string{},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, "context deadline exceeded")
			},
		},
		{
			name: "count context deadline exceeded error",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				row, fieldsDomain := helper.GetValueAndColumnStructToDriverValue(mockDomain)
				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE users.id = (?) ORDER BY users.id ASC LIMIT 10")).
					WithArgs(1).WillReturnRows(
					sqlmock.NewRows(fieldsDomain).AddRow(row...))

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE users.id = (?) LIMIT 10")).
					WithArgs(1).WillReturnError(context.DeadlineExceeded)

				return fields
			},
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
				order:  "users.id ASC",
				fields: []string{
					"*",
				},
				associate: []string{},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, "context deadline exceeded")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name,  func() {
			fields := tt.fields(&tt.args)
			c := mysqlRepository{
				zapLogger: fields.zapLogger,
				db:        fields.db,
			}
			_,err := c.FetchWithFilter(tt.args.ctx, tt.args.limit, tt.args.offset, tt.args.order, tt.args.fields, tt.args.associate, tt.args.filter, tt.args.model, tt.args.args...)
			tt.wantErr(t.T(), err,
				fmt.Sprintf("FetchWithFilter(%v, %v, %v, %v, %v, %v, %v, %v, %v)",
					tt.args.ctx, tt.args.limit, tt.args.offset, tt.args.order, tt.args.fields, tt.args.associate, tt.args.filter, tt.args.model, tt.args.args))
		})
	}
}

func (t *MysqlRepositoryTestSuite) TestFetchWithFilterAndPagination() {
	type args struct {
		ctx       context.Context
		limit     int
		offset    int
		order     string
		fields    []string
		associate []string
		filter    []string
		model     interface{}
		args      []interface{}
	}
	tests := []struct {
		name    string
		fields  func(args *args) fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				row, fieldsDomain := helper.GetValueAndColumnStructToDriverValue(mockDomain)

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE users.id = ? ORDER BY users.id ASC LIMIT 10")).
					WithArgs(1).WillReturnRows(
					sqlmock.NewRows(fieldsDomain).AddRow(row...))

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE users.id = ? LIMIT 10")).
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				return fields
			},
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
				order:  "users.id ASC",
				fields: []string{
					"*",
				},
				associate: []string{},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "success associate",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				row, fieldsDomain := helper.GetValueAndColumnStructToDriverValue(mockDomain)

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` INNER JOIN profile on profile.user_id = users.id WHERE users.id = ? ORDER BY users.id ASC LIMIT 10")).
					WithArgs(1).WillReturnRows(
					sqlmock.NewRows(fieldsDomain).AddRow(row...))

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` INNER JOIN profile on profile.user_id = users.id WHERE users.id = ? LIMIT 10")).
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				return fields
			},
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
				order:  "users.id ASC",
				fields: []string{
					"*",
				},
				associate: []string{
					"INNER JOIN profile on profile.user_id = users.id",
				},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "get query context deadline exceeded error",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}

				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE users.id = ? ORDER BY users.id ASC LIMIT 10")).
					WithArgs(1).WillReturnError(errors.New("context deadline exceeded"))

				return fields
			},
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
				order:  "users.id ASC",
				fields: []string{
					"*",
				},
				associate: []string{},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, "context deadline exceeded")
			},
		},
		{
			name: "count context deadline exceeded error",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				row, fieldsDomain := helper.GetValueAndColumnStructToDriverValue(mockDomain)
				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE users.id = ? ORDER BY users.id ASC LIMIT 10")).
					WithArgs(1).WillReturnRows(
					sqlmock.NewRows(fieldsDomain).AddRow(row...))

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE users.id = ? LIMIT 10")).
					WithArgs(1).WillReturnError(context.DeadlineExceeded)

				return fields
			},
			args: args{
				ctx:    context.TODO(),
				limit:  10,
				offset: 0,
				order:  "users.id ASC",
				fields: []string{
					"*",
				},
				associate: []string{},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, "context deadline exceeded")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name,  func() {
			fields := tt.fields(&tt.args)
			c := mysqlRepository{
				zapLogger: fields.zapLogger,
				db:        fields.db,
			}
			_,err := c.FetchWithFilterAndPagination(tt.args.ctx, tt.args.limit, tt.args.offset, tt.args.order, tt.args.fields, tt.args.associate, tt.args.filter, tt.args.model, tt.args.args...)
			tt.wantErr(t.T(), err,
				fmt.Sprintf("FetchWithFilterAndPagination(%v, %v, %v, %v, %v, %v, %v, %v, %v)",
					tt.args.ctx, tt.args.limit, tt.args.offset, tt.args.order, tt.args.fields, tt.args.associate, tt.args.filter, tt.args.model, tt.args.args))
		})
	}
}

func (t *MysqlRepositoryTestSuite) TestSingleWithFilter() {
	type args struct {
		ctx       context.Context
		fields    []string
		associate []string
		filter    []string
		model     interface{}
		args      []interface{}
	}
	tests := []struct {
		name    string
		fields  func(args *args) fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				row, fieldsDomain := helper.GetValueAndColumnStructToDriverValue(mockDomain)

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE users.id = ? ORDER BY `users`.`id` LIMIT 1")).
					WithArgs(1).WillReturnRows(
					sqlmock.NewRows(fieldsDomain).AddRow(row...))


				return fields
			},
			args: args{
				ctx:    context.TODO(),
				fields: []string{
					"*",
				},
				associate: []string{},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "success associate",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				row, fieldsDomain := helper.GetValueAndColumnStructToDriverValue(mockDomain)

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` INNER JOIN profile on profile.user_id = users.id WHERE users.id = ? ORDER BY `users`.`id` LIMIT 1")).
					WithArgs(1).WillReturnRows(
					sqlmock.NewRows(fieldsDomain).AddRow(row...))

				return fields
			},
			args: args{
				ctx:    context.TODO(),
				fields: []string{
					"*",
				},
				associate: []string{
					"INNER JOIN profile on profile.user_id = users.id",
				},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "get query context deadline exceeded error",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}

				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				mockDB.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE users.id = ? ORDER BY `users`.`id` LIMIT 1")).
					WithArgs(1).WillReturnError(errors.New("context deadline exceeded"))

				return fields
			},
			args: args{
				ctx:    context.TODO(),
				fields: []string{
					"*",
				},
				associate: []string{},
				filter: []string{
					"users.id = ?",
				},
				model: &domain.User{},
				args: []interface{}{
					1,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, "context deadline exceeded")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name,  func() {
			fields := tt.fields(&tt.args)
			c := mysqlRepository{
				zapLogger: fields.zapLogger,
				db:        fields.db,
			}
			err := c.SingleWithFilter(tt.args.ctx, tt.args.fields, tt.args.associate, tt.args.filter, tt.args.model, tt.args.args...)
			tt.wantErr(t.T(), err,
				fmt.Sprintf("SingleWithFilter(%v, %v, %v, %v, %v, %v)", tt.args.ctx, tt.args.fields, tt.args.associate, tt.args.filter, tt.args.model, tt.args.args))
		})
	}
}

func (t *MysqlRepositoryTestSuite) TestUpdate() {
	type args struct {
		ctx  context.Context
		data domain.User
	}
	tests := []struct {
		name    string
		fields  func(args *args) fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}
				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				args.data = mockDomain

				mockDB.ExpectBegin()
				mockDB.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `password_hash`=?, `email`=?, `premium_expires_at`=?, `created_at`=?, `updated_at`=? WHERE `id` = ?")).
					WithArgs(mockDomain.PasswordHash, mockDomain.Email, mockDomain.PremiumExpiresAt, mockDomain.CreatedAt, mockDomain.UpdatedAt, mockDomain.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()

				return fields
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: assert.NoError,
		},
		{
			name: "context deadline exceeded error",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}

				mockDB := t.mock

				mockDomain := domain.User{}
				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				args.data = mockDomain

				mockDB.ExpectBegin()
				mockDB.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `password_hash`=?, `email`=?, `premium_expires_at`=?, `created_at`=?, `updated_at`=? WHERE `id` = ?")).
					WithArgs(mockDomain.PasswordHash, mockDomain.Email, mockDomain.PremiumExpiresAt, mockDomain.CreatedAt, mockDomain.UpdatedAt, mockDomain.ID).
					WillReturnError(errors.New("context deadline exceeded"))
				mockDB.ExpectCommit()

				return fields
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func() {
			fields := tt.fields(&tt.args)
			c := mysqlRepository{
				zapLogger: fields.zapLogger,
				db:        fields.db,
			}
			err := c.Update(tt.args.ctx, tt.args.data)
			tt.wantErr(t.T(), err,
				fmt.Sprintf("Update(%v, %v)", tt.args.ctx, tt.args.data))
		})
	}
}

func (t *MysqlRepositoryTestSuite) TestUpdateSelectedField() {
	type args struct {
		ctx    context.Context
		fields []string
		values map[string]interface{}
		id     int
	}
	tests := []struct {
		name    string
		fields  func(args *args) fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				args.values = map[string]interface{}{
					"email": 						  mockDomain.Email,
				}
				args.id = mockDomain.ID
				mockDB.ExpectBegin()
				mockDB.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `email`=? WHERE id =?")).
					WithArgs(args.values["email"], mockDomain.ID).WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()


				return fields
			},
			args: args{
				ctx: context.TODO(),
				fields: []string{
					"email",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "get query context deadline exceeded error",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}

				mockDB := t.mock

				mockDomain := domain.User{}

				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				args.values = map[string]interface{}{
					"email": 						  mockDomain.Email,
				}
				args.id = mockDomain.ID
				mockDB.ExpectBegin()
				mockDB.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `email`=? WHERE id =?")).
					WithArgs(args.values["email"], mockDomain.ID).WillReturnError(errors.New("context deadline exceeded"))
				mockDB.ExpectCommit()

				return fields
			},
			args: args{
				ctx: context.TODO(),
				fields: []string{
					"email",
				},
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name,  func() {
			fields := tt.fields(&tt.args)
			c := mysqlRepository{
				zapLogger: fields.zapLogger,
				db:        fields.db,
			}
			err := c.UpdateSelectedField(tt.args.ctx, tt.args.fields, tt.args.values, tt.args.id)
			tt.wantErr(t.T(), err,
				fmt.Sprintf("UpdateSelectedField(%v, %v, %v, %v)", tt.args.ctx, tt.args.fields, tt.args.values, tt.args.id))
		})
	}
}

func (t *MysqlRepositoryTestSuite) TestStore() {
	type args struct {
		ctx  context.Context
		data domain.User
	}
	tests := []struct {
		name    string
		fields  func(args *args) fields
		args    args
		want    domain.User
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}
				mockDB := t.mock

				mockDomain := domain.User{}
				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				args.data = mockDomain

				mockDB.ExpectBegin()
				mockDB.ExpectExec(regexp.QuoteMeta("INSERT INTO `users` (`password_hash`,`email`,`premium_expires_at`,`created_at`,`updated_at`,`id`) VALUES (?,?,?,?,?,?)")).
					WithArgs(sqlmock.AnyArg(), mockDomain.Email,sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), mockDomain.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockDB.ExpectCommit()

				return fields
			},
			args: args{
				ctx: context.TODO(),
			},
			want: domain.User{},
			wantErr: assert.NoError,
		},
		{
			name: "context deadline exceeded error",
			fields: func(args *args) fields {
				fields := fields{
					zapLogger: zaplogger.NewZapLogger("", ""),
					db:        t.DB,
				}

				mockDB := t.mock

				mockDomain := domain.User{}
				err := faker.FakeData(&mockDomain)
				t.NoError(err)

				args.data = mockDomain

				mockDB.ExpectBegin()
				mockDB.ExpectExec(regexp.QuoteMeta("INSERT INTO `users` (`password_hash`,`email`,`premium_expires_at`,`created_at`,`updated_at`,`id`) VALUES (?,?,?,?,?,?)")).
					WithArgs(sqlmock.AnyArg(), mockDomain.Email,sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), mockDomain.ID).
					WillReturnError(errors.New("context deadline exceeded"))
				mockDB.ExpectCommit()

				return fields
			},
			args: args{
				ctx: context.TODO(),
			},
			want: domain.User{},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func() {
			fields := tt.fields(&tt.args)
			c := mysqlRepository{
				zapLogger: fields.zapLogger,
				db:        fields.db,
			}
			_, err := c.Store(tt.args.ctx, tt.args.data)
			tt.wantErr(t.T(), err,
				fmt.Sprintf("Store(%v, %v)", tt.args.ctx, tt.args.data))
		})
	}
}

func TestUserMysqlRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MysqlRepositoryTestSuite))
}
