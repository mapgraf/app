package settings

import (
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func ReadMapglSettings(dsInstanceSettings *backend.AppInstanceSettings) (*MapglAppSettings, error) {
	mapglSettingsDTO := &MapglAppSettingsDTO{}

	 err := json.Unmarshal(dsInstanceSettings.JSONData, &mapglSettingsDTO)
        if err != nil {
            return nil, err
        }

    if apiToken, exists := dsInstanceSettings.DecryptedSecureJSONData["apiToken"]; exists {
        mapglSettingsDTO.ApiToken = apiToken
      }


	if mapglSettingsDTO.ApiToken == "" {
		mapglSettingsDTO.ApiToken = ApiToken
	}
	if mapglSettingsDTO.ApiPort == "" {
    		mapglSettingsDTO.ApiPort = ApiPort
    	}

	mapglSettings := &MapglAppSettings{
		ApiToken: mapglSettingsDTO.ApiToken,
		ApiPort:  mapglSettingsDTO.ApiPort,
	}

	return mapglSettings, nil
}
