package application

import (
	"encoding/json"

	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	"github.com/FantasyRL/go-mcp-demo/pkg/jwt"
	"github.com/FantasyRL/go-mcp-demo/pkg/utils"
	"github.com/west2-online/jwch"
)

type LoginData struct {
	Identifier  string // 学号
	Cookie      string // 教务处原始 cookie
	AccessToken string // JWT
}

// Login 教务处验密 -> 拿到 id+cookie -> upsert -> 发 JWT
func (h *Host) Login(id, pwd string) (*LoginData, error) {
	// 请求教务处
	stu := jwch.NewStudent().WithUser(id, pwd)
	identifier, rawCookies, err := stu.GetIdentifierAndCookies()
	if err != nil {
		return nil, err
	}
	// 持久化用户信息
	u, err := h.templateRepository.GetUserByID(h.ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		info, err := stu.GetInfo()
		if err != nil {
			return nil, err
		}
		_, err = h.templateRepository.CreateUserByIDAndName(h.ctx, id, info.Name)
		if err != nil {
			return nil, err
		}
	}

	// 签发 JWT
	token, err := jwt.GenerateAccessToken(id)
	if err != nil {
		return nil, err
	}

	return &LoginData{
		Identifier:  identifier,
		Cookie:      utils.ParseCookiesToString(rawCookies),
		AccessToken: token,
	}, nil
}

func (h *Host) GetUserInfo() (*jwch.StudentDetail, error) {
	loginData, e := utils.ExtractLoginData(h.ctx)
	if !e {
		return nil, errno.ParamError
	}
	stu := jwch.NewStudent().WithLoginData(loginData.ID, utils.ParseCookies(loginData.Cookie))
	info, err := stu.GetInfo()
	if err != nil {
		return nil, err
	}
	return info, nil
}

// UpdateUserSetting 更新用户设置
func (h *Host) UpdateUserSetting(userID string, settingJSON string) error {
	if userID == "" {
		return errno.ParamError
	}
	// 验证 JSON 格式
	var temp map[string]interface{}
	if err := json.Unmarshal([]byte(settingJSON), &temp); err != nil {
		return errno.NewErrNo(50001, "invalid JSON format")
	}

	return h.templateRepository.UpdateUserSetting(h.ctx, userID, settingJSON)
}
