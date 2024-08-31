package settings

const (
	ApiToken     = "token"
	ApiPort = "8089"
)

// ZabbixDatasourceSettingsDTO model
type MapglAppSettingsDTO struct {
	ApiToken string `json:"apiToken"`
	ApiPort     string    `json:"apiPort"`
}

// ZabbixDatasourceSettings model
type MapglAppSettings struct {
	ApiToken string
	ApiPort  string
}
