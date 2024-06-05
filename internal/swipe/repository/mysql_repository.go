package repository

import (
	"context"
	"github.com/radyatamaa/dating-apps-api/internal/swipe"
	"gorm.io/gorm/clause"
	"strings"

	"github.com/radyatamaa/dating-apps-api/internal/domain"
	"github.com/radyatamaa/dating-apps-api/pkg/database/paginator"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"
	"gorm.io/gorm"
)

type mysqlRepository struct {
	zapLogger zaplogger.Logger
	db        *gorm.DB
}

func NewMysqlRepository(db *gorm.DB, zapLogger zaplogger.Logger) swipe.MysqlRepository {
	return &mysqlRepository{
		db:        db,
		zapLogger: zapLogger,
	}
}

func (c mysqlRepository) DB() *gorm.DB {
	return c.db
}

func (c mysqlRepository) FetchWithFilterAndPagination(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (*paginator.Paginator, error) {
	p := paginator.NewPaginator(c.db, offset, limit, model)
	if err := p.FindWithFilter(ctx, order, fields, associate, filter, args...).Select(strings.Join(fields, ",")).Error; err != nil {
		return p, err
	}
	return p, nil
}

func (c mysqlRepository) FetchWithFilter(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (interface{}, error) {
	p := paginator.NewPaginator(c.db, offset, limit, model)
	if err := p.FindWithFilter(ctx, order, fields, associate, filter, args).Select(strings.Join(fields, ",")).Error; err != nil {
		return nil, err
	}
	return model, nil
}

func (c mysqlRepository) SingleWithFilter(ctx context.Context, fields, associate, filter []string, model interface{}, args ...interface{}) error {

	db := c.db.WithContext(ctx)

	if len(fields) > 0 {
		db = db.Select(strings.Join(fields, ","))
	}
	if len(associate) > 0 {
		for _, v := range associate {
			db.Joins(v)
		}
	}

	if len(filter) > 0 && len(args) == len(filter) {
		for i := range filter {
			db = db.Where(filter[i], args[i])
		}
	}

	if err := db.First(model).Error; err != nil {
		return err
	}
	return nil
}

func (c mysqlRepository) Update(ctx context.Context, data domain.Swipe) error {

	err := c.db.WithContext(ctx).Updates(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func (c mysqlRepository) UpdateSelectedField(ctx context.Context, field []string, values map[string]interface{}, id int) error {

	return c.db.WithContext(ctx).Table(domain.Swipe{}.TableName()).Select(field).Where("id =?", id).Updates(values).Error
}

func (c mysqlRepository) Store(ctx context.Context, data domain.Swipe) (domain.Swipe, error) {

	err := c.db.WithContext(ctx).Create(&data).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func (c mysqlRepository) Delete(ctx context.Context, id int) (int, error) {

	err := c.db.WithContext(ctx).Exec("delete from "+domain.Swipe{}.TableName()+" where id =?", id).Error
	if err != nil {
		return id, err
	}
	return id, nil
}

func (c mysqlRepository) SoftDelete(ctx context.Context, id int) (int, error) {
	var data domain.Swipe

	err := c.db.WithContext(ctx).Where("id = ?", id).Delete(&data).Error
	if err != nil {
		return id, err
	}
	return id, nil
}

func (c mysqlRepository) UpdateSelectedFieldWithTx(ctx context.Context, tx *gorm.DB, field []string, values map[string]interface{}, id int) error {

	return tx.WithContext(ctx).Table(domain.Swipe{}.TableName()).Select(field).Where("id =?", id).Updates(values).Error
}

func (c mysqlRepository) StoreWithTx(ctx context.Context, tx *gorm.DB, data domain.Swipe) (int, error) {

	err := tx.WithContext(ctx).Create(&data).Error
	if err != nil {
		return data.ID, err
	}
	return data.ID, nil
}

func (c mysqlRepository) UpsertWithTx(ctx context.Context, tx *gorm.DB, onConflictField []string, data ...domain.Swipe) error {
	var columns []clause.Column

	for i := range onConflictField {
		columns = append(columns, clause.Column{
			Name: onConflictField[i],
		})
	}

	return tx.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   columns,
		DoUpdates: clause.AssignmentColumns([]string{"swipe_type", "updated_at"}),
	}).Create(&data).Error
}

func (c mysqlRepository) Upsert(ctx context.Context, onConflictField []string, data ...domain.Swipe) error {
	var columns []clause.Column

	for i := range onConflictField {
		columns = append(columns, clause.Column{
			Name: onConflictField[i],
		})
	}

	return c.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   columns,
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "profile_id","swipe_type", "created_at", "updated_at"}),
	}).Create(&data).Error
}