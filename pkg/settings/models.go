package settings

const (
	ApiToken = "token"
	ApiPort  = "8089"
)

// SettingsDTO model
type MapglAppSettingsDTO struct {
	ApiToken string `json:"apiToken"`
	ApiPort  string `json:"apiPort"`
}

// Settings model
type MapglAppSettings struct {
	ApiToken string
	ApiPort  string
}
