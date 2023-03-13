/*
 * Glutz App API
 *
 * API to access and configure the Glutz
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package apiservices

import (
	"context"
	"net/http"
	"glutz/apiserver"
	"glutz/conf"
)

// DevicesApiService is a service that implements the logic for the DevicesApiServicer
// This service should implement the business logic for every endpoint for the DevicesApi API.
// Include any external packages or services that will be required by this service.
type DevicesApiService struct {
}

// NewDevicesApiService creates a default api service
func NewDevicesApiService() apiserver.DevicesApiServicer {
	return &DevicesApiService{}
}

// GetDevices - List all devices mapped to eliona assets
func (s *DevicesApiService) GetDevices(ctx context.Context, configId int64) (apiserver.ImplResponse, error) {
	devices, err := conf.GetDevices(ctx, configId)
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.Response(http.StatusOK, devices), nil
}
