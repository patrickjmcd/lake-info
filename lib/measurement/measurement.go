package measurement

import (
	"fmt"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"time"
)

func parseDatetime(dateStr string, timeStr string) (*time.Time, error) {
	addDate := false
	if timeStr == "2400" {
		timeStr = "0000"
		addDate = true
	}

	// Concatenate the date and time strings
	combinedString := dateStr + timeStr

	// Define the layout that corresponds to the date and time strings
	layout := "02Jan20061504"

	// Parse the combined string to a time.Time value
	result, err := time.Parse(layout, combinedString)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if addDate {
		result = result.AddDate(0, 0, 1)
	}

	return &result, nil
}

func ParseMeasurement(measurement []string, lakeName string) (*lakeinfov1.LakeInfoMeasurement, error) {
	if len(measurement) != 8 {
		return nil, ErrInvalidMeasurement
	}

	timestamp, err := parseDatetime(measurement[0], measurement[1])
	if err != nil {
		return nil, err
	}

	level, err := strconv.ParseFloat(measurement[2], 64)
	if err != nil {
		return nil, err
	}

	generation, err := strconv.ParseFloat(measurement[4], 64)
	if err != nil {
		return nil, err
	}

	turbineReleaseRate, err := strconv.ParseFloat(measurement[5], 64)
	if err != nil {
		return nil, err
	}

	spillwayReleaseRate, err := strconv.ParseFloat(measurement[6], 64)
	if err != nil {
		return nil, err
	}

	totalReleaseRate, err := strconv.ParseFloat(measurement[7], 64)
	if err != nil {
		return nil, err
	}

	return &lakeinfov1.LakeInfoMeasurement{
		MeasuredAt:          timestamppb.New(*timestamp),
		LakeName:            lakeName,
		Level:               level,
		Generation:          generation,
		TurbineReleaseRate:  turbineReleaseRate,
		SpillwayReleaseRate: spillwayReleaseRate,
		TotalReleaseRate:    totalReleaseRate,
		CreatedAt:           timestamppb.Now(),
	}, nil

}
