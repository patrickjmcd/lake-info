package server

import (
	"context"
	"github.com/bufbuild/connect-go"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"log"
)

func (s *LakeInfoServer) GetLakeInfo(
	ctx context.Context,
	req *connect.Request[lakeinfov1.GetLakeInfoRequest],
) (*connect.Response[lakeinfov1.GetLakeInfoResponse], error) {

	measurements, err := s.db.GetLakeInfo(ctx, req.Msg.LakeName, req.Msg.StartTime.AsTime(), req.Msg.EndTime.AsTime(), req.Msg.Latest)
	if err != nil {
		log.Println("Error getting lake info: ", err)
		return nil, err
	}

	res := connect.NewResponse(&lakeinfov1.GetLakeInfoResponse{
		Measurements: measurements,
	})
	res.Header().Set("LakeInfo-Version", "v1")
	return res, nil
}
