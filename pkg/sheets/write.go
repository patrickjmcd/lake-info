package sheets

import (
	"context"
	"fmt"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/sheets/v4"
)

func (c *Client) FindNextRow(ctx context.Context, lakeName LakeName) (int, error) {
	startRow := 2
	readRange := fmt.Sprintf("%s!A%d:H", lakeName, startRow)
	resp, err := c.gSheets.Service.Spreadsheets.Values.Get(c.spreadsheetId, readRange).Do()
	if err != nil {
		log.Error().Err(err).Msg("unable to retrieve data from sheet")
		return 0, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	if len(resp.Values) == 0 {
		return startRow, nil
	}
	return startRow + len(resp.Values), nil
}

func makeMeasurementRow(measurement *lakeinfov1.LakeInfoMeasurement) []interface{} {
	return []interface{}{
		measurement.MeasuredAt.AsTime().Local().Format("01/02/2006 15:04:05"),
		measurement.Level,
		measurement.Temperature,
		measurement.Generation,
		measurement.TurbineReleaseRate,
		measurement.SpillwayReleaseRate,
		measurement.TotalReleaseRate,
		measurement.CreatedAt.AsTime().Local().Format("01/02/2006 15:04:05"),
	}
}

func (c *Client) WriteMeasurement(ctx context.Context, lakeName LakeName, measurement *lakeinfov1.LakeInfoMeasurement) error {
	nextRow, err := c.FindNextRow(ctx, lakeName)
	if err != nil {
		return err
	}
	writeRange := fmt.Sprintf("A%d", nextRow)
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, makeMeasurementRow(measurement))
	_, err = c.gSheets.Service.Spreadsheets.Values.Update(c.spreadsheetId, writeRange, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Error().Err(err).Msg("unable to write data to sheet")
		return fmt.Errorf("unable to write data to sheet: %v", err)
	}
	return nil
}

func (c *Client) WriteMeasurements(ctx context.Context, lakeName LakeName, measurements []*lakeinfov1.LakeInfoMeasurement) error {
	nextRow, err := c.FindNextRow(ctx, lakeName)
	if err != nil {
		return err
	}
	writeRange := fmt.Sprintf("A%d", nextRow)
	var vr sheets.ValueRange
	for _, m := range measurements {
		vr.Values = append(vr.Values, makeMeasurementRow(m))
	}
	_, err = c.gSheets.Service.Spreadsheets.Values.Update(c.spreadsheetId, writeRange, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Error().Err(err).Msg("unable to write data to sheet")
		return fmt.Errorf("unable to write data to sheet: %v", err)
	}
	return nil
}
