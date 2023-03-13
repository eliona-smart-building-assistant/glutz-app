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
	"errors"
	"net/http"
	"glutz/apiserver"
)

// CustomizationApiService is a service that implements the logic for the CustomizationApiServicer
// This service should implement the business logic for every endpoint for the CustomizationApi API.
// Include any external packages or services that will be required by this service.
type CustomizationApiService struct {
}

// NewCustomizationApiService creates a default api service
func NewCustomizationApiService() apiserver.CustomizationApiServicer {
	return &CustomizationApiService{}
}

// GetDashboardTemplateByName - Get a full dashboard template
func (s *CustomizationApiService) GetDashboardTemplateByName(ctx context.Context, dashboardTemplateName string, projectId string) (apiserver.ImplResponse, error) {
	// TODO - update GetDashboardTemplateByName with the required logic for this service method.
	// Add api_customization_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, Dashboard{}) or use other options such as http.Ok ...
	//return Response(200, Dashboard{}), nil

	//TODO: Uncomment the next line to return response Response(404, {}) or use other options such as http.Ok ...
	//return Response(404, nil),nil

	return apiserver.Response(http.StatusNotImplemented, nil), errors.New("GetDashboardTemplateByName method not implemented")
}

// GetDashboardTemplateNames - List available dashboard templates
func (s *CustomizationApiService) GetDashboardTemplateNames(ctx context.Context) (apiserver.ImplResponse, error) {
	// TODO - update GetDashboardTemplateNames with the required logic for this service method.
	// Add api_customization_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, []string{}) or use other options such as http.Ok ...
	//return Response(200, []string{}), nil

	return apiserver.Response(http.StatusNotImplemented, nil), errors.New("GetDashboardTemplateNames method not implemented")
}
