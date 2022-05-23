package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/turbot/steampipe-plugin-sdk/v3/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v3/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v3/plugin/transform"
)

//// TABLE DEFINITION

type threatIntelSetInfo = struct {
	guardduty.GetThreatIntelSetOutput
	ThreatIntelSetID string
	DetectorID       string
}

func tableAwsGuardDutyThreatIntelSet(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_guardduty_threat_intel_set",
		Description: "AWS GuardDuty ThreatIntelSet",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"detector_id", "threat_intel_set_id"}),
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: isNotFoundError([]string{"InvalidInputException", "BadRequestException"}),
			},
			Hydrate: getGuardDutyThreatIntelSet,
		},
		List: &plugin.ListConfig{
			ParentHydrate: listGuardDutyDetectors,
			Hydrate:       listGuardDutyThreatIntelSets,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "detector_id", Require: plugin.Optional},
			},
		},
		GetMatrixItem: BuildRegionList,
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "name",
				Description: "A ThreatIntelSet name displayed in all findings that are generated by activity that involves IP addresses included in this ThreatIntelSet.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getGuardDutyThreatIntelSet,
			},
			{
				Name:        "threat_intel_set_id",
				Description: "The ID of the ThreatIntelSet.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ThreatIntelSetID"),
			},
			{
				Name:        "detector_id",
				Description: "The ID of the detector.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getGuardDutyThreatIntelSet,
				Transform:   transform.FromField("DetectorID"),
			},
			{
				Name:        "format",
				Description: "The format of the threatIntelSet.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getGuardDutyThreatIntelSet,
			},
			{
				Name:        "location",
				Description: "The URI of the file that contains the ThreatIntelSet.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getGuardDutyThreatIntelSet,
			},
			{
				Name:        "status",
				Description: "The status of threatIntelSet file uploaded.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getGuardDutyThreatIntelSet,
			},
			// Standard columns
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Hydrate:     getGuardDutyThreatIntelSet,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "tags",
				Description: resourceInterfaceDescription("tags"),
				Type:        proto.ColumnType_JSON,
				Hydrate:     getGuardDutyThreatIntelSet,
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAwsGuardDutyThreatIntelSetAkas,
				Transform:   transform.FromValue(),
			},
		}),
	}
}

//// LIST FUNCTION

func listGuardDutyThreatIntelSets(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Get details of detector
	detectorID := h.Item.(detectorInfo).DetectorID

	// Create session
	svc, err := GuardDutyService(ctx, d)
	if err != nil {
		return nil, err
	}
	equalQuals := d.KeyColumnQuals

	// Minimize the API call with the given detector_id
	if equalQuals["detector_id"] != nil {
		if equalQuals["detector_id"].GetStringValue() != "" {
			if equalQuals["detector_id"].GetStringValue() != "" && equalQuals["detector_id"].GetStringValue() != detectorID {
				return nil, nil
			}
		} else if len(getListValues(equalQuals["detector_id"].GetListValue())) > 0 {
			if !strings.Contains(fmt.Sprint(getListValues(equalQuals["detector_id"].GetListValue())), detectorID) {
				return nil, nil
			}
		}
	}

	input := &guardduty.ListThreatIntelSetsInput{
		DetectorId: &detectorID,
		MaxResults: aws.Int64(50),
	}

	// List call
	err = svc.ListThreatIntelSetsPages(
		input,
		func(page *guardduty.ListThreatIntelSetsOutput, isLast bool) bool {
			for _, result := range page.ThreatIntelSetIds {
				d.StreamLeafListItem(ctx, threatIntelSetInfo{
					ThreatIntelSetID: *result,
					DetectorID:       detectorID,
				})

				// Context may get cancelled due to manual cancellation or if the limit has been reached
				if d.QueryStatus.RowsRemaining(ctx) == 0 {
					return false
				}
			}
			return !isLast
		},
	)

	return nil, err
}

//// HYDRATE FUNCTIONS

func getGuardDutyThreatIntelSet(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getGuardDutyThreatIntelSet")

	// Create Session
	svc, err := GuardDutyService(ctx, d)
	if err != nil {
		return nil, err
	}

	var id string
	var detectorID string
	if h.Item != nil {
		detectorID = h.Item.(threatIntelSetInfo).DetectorID
		id = h.Item.(threatIntelSetInfo).ThreatIntelSetID
	} else {
		detectorID = d.KeyColumnQuals["detector_id"].GetStringValue()
		id = d.KeyColumnQuals["threat_intel_set_id"].GetStringValue()
	}

	params := &guardduty.GetThreatIntelSetInput{
		DetectorId:       &detectorID,
		ThreatIntelSetId: &id,
	}

	op, err := svc.GetThreatIntelSet(params)
	if err != nil {
		logger.Debug("getGuardDutyThreatIntelSet", "ERROR", err)
		return nil, err
	}

	return threatIntelSetInfo{*op, id, detectorID}, nil
}

//// TRANSFORM FUNCTIONS

func getAwsGuardDutyThreatIntelSetAkas(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getAwsGuardDutyThreatIntelSetAkas")
	data := h.Item.(threatIntelSetInfo)
	region := d.KeyColumnQualString(matrixKeyRegion)

	getCommonColumnsCached := plugin.HydrateFunc(getCommonColumns).WithCache()
	c, err := getCommonColumnsCached(ctx, d, h)
	if err != nil {
		return nil, err
	}
	commonColumnData := c.(*awsCommonColumnData)
	aka := "arn:" + commonColumnData.Partition + ":guardduty:" + region + ":" + commonColumnData.AccountId + ":detector" + "/" + data.DetectorID + "/threatintelset/" + data.ThreatIntelSetID

	return []string{aka}, nil
}
