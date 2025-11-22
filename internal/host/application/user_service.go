package application

import (
	"context"
	"time"

	"github.com/FantasyRL/go-mcp-demo/api/model/model"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/db"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/query"
	"github.com/golang-jwt/jwt/v5"
	"github.com/west2-online/jwch"
)

var jwtSecret = []byte("change_me_in_prod") // 可放到 config

type UserService struct{}

func NewUserService() *UserService { return &UserService{} }

type LoginResp struct {
	Identifier  string `json:"identifier"` // 学号
	Cookie      string `json:"cookie"`     // 教务处原始 cookie
	AccessToken string `json:"accessToken"`
}

// LoginByJWC 完整链路：教务处验密 -> 拿到 id+cookie -> upsert -> 发 JWT
func (s *UserService) LoginByJWC(ctx context.Context, stuNo, pwd string) (*LoginResp, error) {
	// 1. 请求教务处
	cli := jwch.NewStudent()
	if err := cli.Login(); err != nil {
		return nil, err
	}
	identifier := cli.Identifier
	cookie := cli.Password

	// 2. 不存在就插入（事务版）
	var user model.User
	err := db.Transaction[*query.Query](ctx, func(ctx context.Context) error {
		q := db.NewDBWithQuery(db.RawDB(), query.Use).Get(ctx)
		var err error
		_, err = q.Users.
			WithContext(ctx).
			Where(q.Users.Name.Eq(identifier)).
			FirstOrCreate()
		return err
	})
	if err != nil {
		return nil, err
	}

	// 3. 签发 JWT（payload 只放 uid 和过期时间）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": user.ID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	ss, _ := token.SignedString(jwtSecret)

	return &LoginResp{
		Identifier:  identifier,
		Cookie:      cookie,
		AccessToken: ss,
	}, nil
}
