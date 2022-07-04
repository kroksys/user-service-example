package service

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/kroksys/user-service-example/pkg/db"
	"github.com/kroksys/user-service-example/pkg/models"
	"github.com/kroksys/user-service-example/pkg/pb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Protobuf generated user service implementation
type UserService struct {
	pb.UnimplementedUserServiceServer
	Redis *redis.Client
}

func (s UserService) AddUser(ctx context.Context, in *pb.AddUserRequest) (*pb.UserResponse, error) {
	log.Println("UserService:AddUser")
	if in.Email == "" {
		log.Println("UserService:AddUser empty email address provided")
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
		log.Printf("UserService:AddUser error creating user %s\n", err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := userRec.ToUserResponse()

	// Publish changes to redis pubsub
	if s.Redis != nil {
		data, err := proto.Marshal(&pb.WatchResponse{
			Method: pb.WatchResponse_CREATE,
			User:   resp,
		})
		if err != nil {
			log.Printf("AddUser failed to marshal data for publishing user changes %v\n", err)
		} else {
			if err := s.Redis.Publish(ctx, "users", data).Err(); err != nil {
				log.Printf("AddUser publish user changes err: %v\n", err)
			}
		}
	}

	return userRec.ToUserResponse(), nil
}

func (s UserService) ModifyUser(ctx context.Context, in *pb.ModifyUserRequest) (*pb.UserResponse, error) {
	log.Println("UserService:ModifyUser")
	// Checking id and creating user record from it
	if in.Id == "" {
		log.Println("UserService:ModifyUser empty id provided")
		return nil, status.Errorf(codes.InvalidArgument, "ModifyUser: id must not be empty")
	}
	id, err := uuid.Parse(in.Id)
	if err != nil {
		log.Printf("UserService:ModifyUser could not parse id to uuid %s\n", err.Error())
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
		log.Printf("UserService:ModifyUser error updating user %s\n", err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := userRec.ToUserResponse()

	// Publish changes to redis pubsub
	if s.Redis != nil {
		data, err := proto.Marshal(&pb.WatchResponse{
			Method: pb.WatchResponse_UPDATE,
			User:   resp,
		})
		if err != nil {
			log.Printf("ModifyUser failed to marshal data for publishing user changes %v\n", err)
		} else {
			if err := s.Redis.Publish(ctx, "users", data).Err(); err != nil {
				log.Printf("ModifyUser publish user changes err: %v\n", err)
			}
		}
	}

	return resp, nil
}

func (s UserService) RemoveUser(ctx context.Context, in *pb.RemoveUserRequest) (*pb.RemoveUserResponse, error) {
	log.Println("UserService:RemoveUser")
	if in.Id == "" {
		log.Println("UserService:RemoveUser empty id provided")
		return nil, status.Errorf(codes.InvalidArgument, "RemoveUser: id must not be empty")
	}
	uid, err := uuid.Parse(in.Id)
	if err != nil {
		log.Printf("UserService:RemoveUser could not parse id to uuid %s\n", err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	err = db.DeleteUser(uid)
	if err != nil {
		log.Printf("UserService:RemoveUser error deleting user %s\n", err.Error())
		status.Errorf(codes.Internal, err.Error())
	}

	// Publish changes to redis pubsub
	if s.Redis != nil {
		data, err := proto.Marshal(&pb.WatchResponse{
			Method: pb.WatchResponse_DELETE,
			User: &pb.UserResponse{
				Id: in.Id,
			},
		})
		if err != nil {
			log.Printf("RemoveUser failed to marshal data for publishing user changes %v\n", err)
		} else {
			if err := s.Redis.Publish(ctx, "users", data).Err(); err != nil {
				log.Printf("RemoveUser publish user changes err: %v\n", err)
			}
		}
	}

	return &pb.RemoveUserResponse{}, nil
}

func (UserService) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	log.Println("UserService:ListUsers")
	users, err := db.ListUsers(int(in.GetLimit()), int(in.GetOffset()), in.GetCountry())
	if err != nil {
		log.Printf("UserService:ListUsers error listing user %s\n", err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	result := []*pb.UserResponse{}
	for _, u := range users {
		result = append(result, u.ToUserResponse())
	}
	return &pb.ListUsersResponse{Users: result}, nil
}

func (s UserService) Watch(in *pb.WatchRequest, stream pb.UserService_WatchServer) error {
	pubsub := s.Redis.Subscribe(stream.Context(), "users")
	defer pubsub.Close()

	var err error
	for {
		select {
		// Client closed stream
		case <-stream.Context().Done():
			return nil
		// Notify Client about User changes
		case data := <-pubsub.Channel():
			resp := &pb.WatchResponse{}
			err = proto.Unmarshal([]byte(data.Payload), resp)
			if err != nil {
				return err
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
		}
	}
}
