package server

import "github.com/patrickjmcd/lake-info/dal"

type LakeInfoServer struct {
	db *dal.Client
}

func New(db *dal.Client) *LakeInfoServer {
	return &LakeInfoServer{
		db: db,
	}
}
