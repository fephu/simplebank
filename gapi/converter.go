package gapi

import (
	db "github.com/elkelk/simplebank/db/sqlc"
	"github.com/elkelk/simplebank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreateAt:          timestamppb.New(user.CreatedAt),
	}
}
