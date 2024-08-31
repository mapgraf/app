package plugin

import (
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"mapgl-app/pkg/httpadapter"
	"mapgl-app/pkg/settings"
	// 	"net/http"
	//  	"fmt"
	"github.com/gorilla/mux"
	//"mapgl-app/pkg/signal"
)

// Make sure App implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. Plugin should not implement all these interfaces - only those which are
// required for a particular task.
var (
	_ backend.CallResourceHandler   = (*App)(nil)
	_ instancemgmt.InstanceDisposer = (*App)(nil)
	_ backend.CheckHealthHandler    = (*App)(nil)
)

// App is an example app backend plugin which can respond to data queries.
type App struct {
	backend.CallResourceHandler
	MapglSettings *settings.MapglAppSettings
}

// NewApp creates a new example *App instance.
func NewApp(ctx context.Context, appSettings backend.AppInstanceSettings) (instancemgmt.Instance, error) {
	var app App

	mapglSettings, err := settings.ReadMapglSettings(&appSettings)

	log.DefaultLogger.Debug("Initializing new mapgl app instance")
	if err != nil {
		log.DefaultLogger.Error("Error parsing Mapgl settings", "error", err)
		return nil, err
	}

	// Use a httpadapter (provided by the SDK) for resource calls. This allows us
	// to use a *http.ServeMux for resource calls, so we can map multiple routes
	// to CallResource without having to implement extra logic.

	r := mux.NewRouter()
	app.MapglSettings = mapglSettings
	app.registerRoutes(r)
	app.CallResourceHandler = httpadapter.New(r)
	return &app, nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created.
func (a *App) Dispose() {
	// cleanup
}

// CheckHealth handles health checks sent from Grafana to the plugin.
func (a *App) CheckHealth(_ context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "ok",
	}, nil
}

// func (a *App) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
//   u := req.GetHTTPHeaders() //("Connection")
//  log.DefaultLogger.Info(fmt.Sprintf("Request Headers: %v", u))
//  log.DefaultLogger.Info(fmt.Sprintf("sender: %v", sender))
//
//
//   return sender.Send(&backend.CallResourceResponse{
//         Status: http.StatusOK,
//         Body: []byte(fmt.Sprintf(`{"Hello, world! %d"}`, u)),
//     })
// }
