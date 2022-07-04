package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/kroksys/user-service-example/pkg/db"
	"github.com/kroksys/user-service-example/pkg/models"
	"github.com/kroksys/user-service-example/pkg/pb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Protobuf generated user service implementation
type UserService struct {
	pb.UnimplementedUserServiceServer
}

func (UserService) AddUser(ctx context.Context, in *pb.AddUserRequest) (*pb.UserResponse, error) {
	if in.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "AddUser: email must not be empty")
	}
	userRec := models.User{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Nickname:  in.Nickname,
		Password:  in.Password,
		Email:     in.Email,
		Country:   in.Country,
	}
	err := db.CreateUser(&userRec)
	if err != nil {
		return nil, err
	}
	return userRec.ToUserResponse(), nil
}

func (UserService) ModifyUser(ctx context.Context, in *pb.ModifyUserRequest) (*pb.UserResponse, error) {
	// Checking id and creating user record from it
	if in.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "ModifyUser: id must not be empty")
	}
	id, err := uuid.Parse(in.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	userRec := &models.User{ID: id}

	// Because each of provided fields from ModifyUserRequest are optinal
	// it is possible to update only provided data.
	updateMap := map[string]interface{}{}
	if in.FirstName != nil {
		updateMap["first_name"] = in.GetFirstName()
	}
	if in.LastName != nil {
		updateMap["last_name"] = in.GetLastName()
	}
	if in.Nickname != nil {
		updateMap["nickname"] = in.GetNickname()
	}
	if in.Password != nil {
		updateMap["password"] = in.GetPassword()
	}
	if in.Email != nil {
		updateMap["email"] = in.GetEmail()
	}
	if in.Country != nil {
		updateMap["country"] = in.GetCountry()
	}

	// Update and handle error if exists
	err = db.UpdateUserByMap(userRec, updateMap)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return userRec.ToUserResponse(), nil
}

func (UserService) RemoveUser(ctx context.Context, in *pb.RemoveUserRequest) (*pb.RemoveUserResponse, error) {
	if in.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "RemoveUser: id must not be empty")
	}
	uid, err := uuid.Parse(in.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	err = db.DeleteUser(uid)
	if err != nil {
		status.Errorf(codes.Internal, err.Error())
	}

	return &pb.RemoveUserResponse{}, nil
}

func (UserService) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := db.ListUsers(int(in.GetLimit()), int(in.GetOffset()), in.GetCountry())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	result := []*pb.UserResponse{}
	for _, u := range users {
		result = append(result, u.ToUserResponse())
	}
	return &pb.ListUsersResponse{Users: result}, nil
}
