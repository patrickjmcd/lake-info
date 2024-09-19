package sheets

import (
	"context"
	"fmt"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"time"
)

func tryParsingDateString(s string) (time.Time, error) {
	layouts := []string{
		"01/02/2006",
		"1/2/2006",
		"1/2/06",
		"1/2",
		"2006-01-02",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date string: %s", s)
}

func mapRowToMeasurement(ctx context.Context, row []interface{}) (*lakeinfov1.LakeInfoMeasurement, error) {
	log.Trace().Interface("row", row).Msg("row")
	if row == nil || len(row) < 2 {
		log.Warn().Msg("empty row in sheet")
		return nil, fmt.Errorf("empty row in sheet")
	}

	measurement := &lakeinfov1.LakeInfoMeasurement{}

	if measuredAtDateStr, ok := row[0].(string); ok {
		if measuredAtDateStr != "" {
			measuredAt, err := tryParsingDateString(measuredAtDateStr)
			if err != nil {
				log.Error().Err(err).Msg("unable to parse measuredAt date")
				return nil, fmt.Errorf("unable to parse measuredAt date: %v", err)
			}
			measurement.MeasuredAt = timestamppb.New(measuredAt)
		}
	} else {
		log.Error().Msgf("expected measuredAt type string, but got %T", row[1])
		return nil, fmt.Errorf("expected measuredAt type string, but got %T", row[1])
	}

	if levelStr, ok := row[1].(string); ok {
		if levelStr != "" {
			level, err := strconv.ParseFloat(levelStr, 64)
			if err != nil {
				log.Error().Err(err).Msg("unable to parse level")
				return nil, fmt.Errorf("unable to parse level: %v", err)
			}
			measurement.Level = level
		}
	}

	if tempStr, ok := row[2].(string); ok {
		if tempStr != "" {
			temp, err := strconv.ParseFloat(tempStr, 64)
			if err != nil {
				log.Error().Err(err).Msg("unable to parse temp")
				return nil, fmt.Errorf("unable to parse temp: %v", err)
			}
			measurement.Temperature = temp
		}
	}

	if generationStr, ok := row[3].(string); ok {
		if generationStr != "" {
			generation, err := strconv.ParseFloat(generationStr, 64)
			if err != nil {
				log.Error().Err(err).Msg("unable to parse generation")
				return nil, fmt.Errorf("unable to parse generation: %v", err)
			}
			measurement.Generation = generation
		}
	}

	if turbineReleaseStr, ok := row[4].(string); ok {
		if turbineReleaseStr != "" {
			turbineRelease, err := strconv.ParseFloat(turbineReleaseStr, 64)
			if err != nil {
				log.Error().Err(err).Msg("unable to parse turbineRelease")
				return nil, fmt.Errorf("unable to parse turbineRelease: %v", err)
			}
			measurement.TurbineReleaseRate = turbineRelease
		}
	}

	if spillwayReleaseStr, ok := row[5].(string); ok {
		if spillwayReleaseStr != "" {
			spillwayRelease, err := strconv.ParseFloat(spillwayReleaseStr, 64)
			if err != nil {
				log.Error().Err(err).Msg("unable to parse spillwayRelease")
				return nil, fmt.Errorf("unable to parse spillwayRelease: %v", err)
			}
			measurement.SpillwayReleaseRate = spillwayRelease
		}
	}

	if totalReleaseStr, ok := row[6].(string); ok {
		if totalReleaseStr != "" {
			totalRelease, err := strconv.ParseFloat(totalReleaseStr, 64)
			if err != nil {
				log.Error().Err(err).Msg("unable to parse totalRelease")
				return nil, fmt.Errorf("unable to parse totalRelease: %v", err)
			}
			measurement.TotalReleaseRate = totalRelease
		}
	}

	if insertedAtStr, ok := row[7].(string); ok {
		if insertedAtStr != "" {
			insertedAt, err := tryParsingDateString(insertedAtStr)
			if err != nil {
				log.Error().Err(err).Msg("unable to parse insertedAt date")
				return nil, fmt.Errorf("unable to parse insertedAt date: %v", err)
			}
			measurement.CreatedAt = timestamppb.New(insertedAt)
		}
	}

	return measurement, nil
}

func (c *Client) ReadMeasurements(ctx context.Context, lakeName LakeName) ([]*lakeinfov1.LakeInfoMeasurement, error) {

	readRange := fmt.Sprintf("%s!A2:H", lakeName)
	resp, err := c.gSheets.Service.Spreadsheets.Values.Get(c.spreadsheetId, readRange).Do()
	if err != nil {
		log.Error().Err(err).Msg("unable to retrieve data from sheet")
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	var measurements []*lakeinfov1.LakeInfoMeasurement

	for _, row := range resp.Values {
		log.Trace().Interface("row", row).Msg("row")
		if row == nil || len(row) < 2 {
			log.Warn().Msgf("empty row in sheet %s\n", c.sheetName)
			continue
		}

		measurement, err := mapRowToMeasurement(ctx, row)
		if err != nil {
			log.Error().Err(err).Msg("unable to map row to expert")
			continue
		}
		measurements = append(measurements, measurement)
	}
	return measurements, nil
}
