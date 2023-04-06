package eliona

import (
	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
	"github.com/eliona-smart-building-assistant/go-utils/common"
)

func GlutzDashboard(projectId string) (api.Dashboard, error) {
	dashboard := api.Dashboard{}
	dashboard.Name = "Glutz Doors"
	dashboard.ProjectId = projectId

	dashboard.Widgets = []api.Widget{}

	assets, _, err := client.NewClient().AssetsApi.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName("glutz_device").
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, err
	}

	for _, asset := range assets {
		widget := api.Widget{
			WidgetTypeName: "Glutz Test",
			AssetId:        asset.Id,
			Details: map[string]interface{}{
				"size":     1,
				"timespan": 7,
			},
			Data: []api.WidgetData{
				{
					ElementSequence: nullableInt32(1),
					AssetId:         asset.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "open",
						"description":         "Entry Door, Big Office, Main Building",
						"key":                 "_SETPOINT",
						"seq":                 0,
						"subtype":             "output",
					},
				},
				{
					ElementSequence: nullableInt32(1),
					AssetId:         asset.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "open",
						"description":         "Open",
						"key":                 "_CURRENT",
						"seq":                 0,
						"subtype":             "output",
					},
				},
				{
					ElementSequence: nullableInt32(2),
					AssetId:         asset.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "openable",
						"description":         "Door Status: ",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
			},
		}
		dashboard.Widgets = append(dashboard.Widgets, widget)
	}
	return dashboard, nil
}

func nullableInt32(val int32) api.NullableInt32 {
	return *api.NewNullableInt32(common.Ptr[int32](val))
}
