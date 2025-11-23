package infra

import (
	"context"
	"errors"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/db"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/query"
	"gorm.io/gorm"
)

type TemplateRepository struct {
	db *db.DB[*query.Query]
}

func NewTemplateRepository(db *db.DB[*query.Query]) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) CreateUserByIDAndName(ctx context.Context, id string, name string) (*model.Users, error) {
	d := r.db.Get(ctx)
	user := &model.Users{
		ID:   id,
		Name: name,
	}
	// 由于user是指针类型，Create方法会自动填充user的其他字段（如时间戳）
	err := d.WithContext(ctx).Users.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *TemplateRepository) GetUserByID(ctx context.Context, id string) (*model.Users, error) {
	d := r.db.Get(ctx)
	user, err := d.WithContext(ctx).Users.Where(d.Users.ID.Eq(id)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return user, nil
}
