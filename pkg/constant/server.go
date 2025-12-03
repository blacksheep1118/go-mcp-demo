package constant

import "time"

const (
	AccessTokenTTL = time.Hour * 24 * 7
)

// DefaultUserSettingJSON 默认用户设置
const DefaultUserSettingJSON = `{
	"theme": "light",
	"language": "zh-CN",
	"notification": {
		"email": true,
		"push": true
	},
	"preferences": {
		"auto_save": true,
		"show_week_number": true
	}
}`
