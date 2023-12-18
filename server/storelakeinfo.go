package server

import (
	"context"
	"github.com/bufbuild/connect-go"
	lakeinfov1 "github.com/patrickjmcd/lake-info/gen/lakeinfo/v1"
	"log"
)

func (s *LakeInfoServer) StoreLakeInfo(
	ctx context.Context,
	req *connect.Request[lakeinfov1.StoreLakeInfoRequest],
) (*connect.Response[lakeinfov1.StoreLakeInfoResponse], error) {
	log.Println("Request headers: ", req.Header())

	err := s.db.StoreLakeInfo(ctx, req.Msg.Measurements)
	if err != nil {
		return nil, err
	}

	res := connect.NewResponse(&lakeinfov1.StoreLakeInfoResponse{})
	res.Header().Set("LakeInfo-Version", "v1")
	return res, nil
}
