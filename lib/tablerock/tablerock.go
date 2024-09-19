package tablerock

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"github.com/patrickjmcd/lake-info/lib/measurement"
)

const LakeURL = "https://www.swl-wc.usace.army.mil/pages/data/tabular/htm/tab7d.htm"
const LakeName = "tablerock"
const SheetName = "Table Rock Lake"

func GetAllRecords(url string) ([]*lakeinfov1.LakeInfoMeasurement, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var records []*lakeinfov1.LakeInfoMeasurement
	htmlParts := strings.Split(string(bytes), "<hr>")
	lines := strings.Split(htmlParts[1], "\n")
	for i, line := range lines {
		if i > 6 {
			var measurements []string
			linesSplit := strings.Split(line, " ")
			for _, v := range linesSplit {
				if v != "" {
					measurements = append(measurements, v)
				}
			}
			if len(measurements) == 0 {
				continue
			}

			record, err := measurement.ParseMeasurement(measurements, LakeName)
			if err != nil {
				slog.Error("error parsing measurement", "error", err, "measurements", measurements)
				continue
			}
			records = append(records, record)
		}
	}

	return records, nil
}

func GetLatestRecord(url string) (*lakeinfov1.LakeInfoMeasurement, error) {
	records, err := GetAllRecords(url)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no records found")
	}

	latestRecord := records[len(records)-1]
	return latestRecord, nil
}
