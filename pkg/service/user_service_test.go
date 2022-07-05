package service

import (
	"context"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/kroksys/user-service-example/pkg/db"
	"github.com/kroksys/user-service-example/pkg/models"
	"github.com/kroksys/user-service-example/pkg/pb/v1"
	"google.golang.org/grpc"
)

func init() {
	// Use real database connection provided by docker-compose.
	testDatabaseConnectionString := "user:userpw@tcp(localhost:3306)/users?parseTime=true"
	if os.Getenv("USERSERVICE_TEST_CONNECTION_STRING") != "" {
		testDatabaseConnectionString = os.Getenv("USERSERVICE_TEST_CONNECTION_STRING")
	}
	err := db.Connect(testDatabaseConnectionString)
	if err != nil {
		log.Fatalf("user_service_test.go: could not connect to testing database: %v", err)
	}

	// Cleanup database
	err = db.DB.Migrator().DropTable(&models.User{})
	if err != nil {
		log.Fatalf("user_service_test.go: could not drop user table: %v", err)
	}
	err = db.DB.Migrator().CreateTable(&models.User{})
	if err != nil {
		log.Fatalf("user_service_test.go: could not create user table: %v", err)
	}
}

func TestAddUser(t *testing.T) {
	s := UserService{}
	testEmail := "john.doe@email.com"

	// Try to add user without email address
	req := &pb.AddUserRequest{}
	_, err := s.AddUser(context.Background(), req)
	if err == nil {
		t.Errorf("TestAddUser(%s) adding empty email expected to receive an error. Got: nil", testEmail)
	}

	// Add first valid user
	req.Email = testEmail
	resp, err := s.AddUser(context.Background(), req)
	if err != nil {
		t.Errorf("TestAddUser got unexpected error: %v", err)
	}
	if resp.Email != testEmail {
		t.Errorf("TestAddUser(%s)=%v, wanted %v", testEmail, resp.Email, testEmail)
	}
	if resp.Id == "" || resp.Id == "00000000-0000-0000-0000-000000000000" {
		t.Errorf("TestAddUser(%s) responsed with empty Id", testEmail)
	}

	// Try to add the same email again - because email is unique it should return an error.
	_, err = s.AddUser(context.Background(), req)
	if err == nil {
		t.Errorf("TestAddUser(%s) adding email second time expected to receive an 'Duplicate entry' error. Got: nil", testEmail)
	}
}

func TestModifyUser(t *testing.T) {
	s := UserService{}
	testEmail := "john2.doe@email.com"

	// Try to modify user without id
	req := &pb.ModifyUserRequest{}
	_, err := s.ModifyUser(context.Background(), req)
	if err == nil {
		t.Errorf("TestModifyUser: user id was not provided and ModifyUser should return err. Got: nil")
	}

	// Creating user for further updates
	createReq := &pb.AddUserRequest{Email: testEmail}
	userRec, err := s.AddUser(context.Background(), createReq)
	if err != nil {
		t.Fatalf("TestModifyUser: failed to add user for further update testing: %v", err)
	}

	// Just a placeholder for data that should be updated
	u := models.User{
		FirstName: "AnotherJohn",
		LastName:  "Dow",
		Nickname:  "Piggin",
		Password:  "the secret",
		Email:     "anotherjohn@email.com",
		Country:   "US",
	}

	req.Id = userRec.Id
	req.FirstName = stringPtr(u.FirstName)
	req.LastName = stringPtr(u.LastName)
	req.Nickname = stringPtr(u.Nickname)
	req.Password = stringPtr(u.Password)
	req.Email = stringPtr(u.Email)
	req.Country = stringPtr(u.Country)
	userResp, err := s.ModifyUser(context.Background(), req)
	if err != nil {
		t.Errorf("TestModifyUser: failed to modify user: %v", err)
	}
	if userResp.FirstName != u.FirstName ||
		userResp.LastName != u.LastName ||
		userResp.Nickname != u.Nickname ||
		userResp.Password != u.Password ||
		userResp.Email != u.Email ||
		userResp.Country != u.Country {
		t.Errorf("TestModifyUser: data did not update. Got: %v Expected: %v", userResp, u)
	}

	// Try to enter invalid UUID
	req.Id = "asdvd-asdv-asd-asddd"
	_, err = s.ModifyUser(context.Background(), req)
	if err == nil {
		t.Errorf("TestModifyUser updating user with invalid uuid should return error. Got: nil")
	}
}

func TestRemoveUser(t *testing.T) {
	s := UserService{}
	testEmail := "john3.doe@email.com"

	// Creating user for removal
	createReq := &pb.AddUserRequest{Email: testEmail}
	userRec, err := s.AddUser(context.Background(), createReq)
	if err != nil {
		t.Fatalf("TestRemoveUser: failed to add user for further removal: %v", err)
	}

	// Remove user without ID
	_, err = s.RemoveUser(context.Background(), &pb.RemoveUserRequest{})
	if err == nil {
		t.Errorf("TestRemoveUser remove user without id should return error. Got: nil")
	}

	// Try to enter invalid UUID
	_, err = s.RemoveUser(context.Background(), &pb.RemoveUserRequest{Id: "asdvd-asdv-asd-asddd"})
	if err == nil {
		t.Errorf("TestRemoveUser remove user with invalid uuid should return error. Got: nil")
	}

	// Final valid removal
	_, err = s.RemoveUser(context.Background(), &pb.RemoveUserRequest{Id: userRec.Id})
	if err != nil {
		t.Errorf("TestRemoveUser: failed to remove user: %v", err)
	}
}

func TestListUsers(t *testing.T) {
	s := UserService{}
	populateDatabase(s)
	resp, err := s.ListUsers(context.Background(), &pb.ListUsersRequest{})
	if err != nil {
		t.Errorf("TestListUsers: failed to list users: %v", err)
	}
	totalLen := len(resp.Users)

	// List with limit 1
	resp, err = s.ListUsers(context.Background(), &pb.ListUsersRequest{Limit: intPtr(1)})
	if err != nil {
		t.Errorf("TestListUsers: failed to list users with limit: %v", err)
	}
	if len(resp.Users) != 1 {
		t.Errorf("TestListUsers: failed to list users with limit 1: response got %d records", len(resp.Users))
	}

	// List with out of limit offset
	resp, err = s.ListUsers(context.Background(), &pb.ListUsersRequest{Offset: intPtr(int32(1))})
	if err != nil {
		t.Errorf("TestListUsers: failed to list users with offset %d: %v", totalLen, err)
	}
	if len(resp.Users) != 1 {
		t.Errorf("TestListUsers: with larger offset than data should return at lest one result. Got %d records", len(resp.Users))
	}

	// List by country AU ... should have 2 records
	resp, err = s.ListUsers(context.Background(), &pb.ListUsersRequest{Country: "AU"})
	if err != nil {
		t.Errorf("TestListUsers: failed to list users with Country AU: %v", err)
	}
	if len(resp.Users) != 2 {
		t.Errorf("TestListUsers: list filtering by Country AU should have only 2 records. Got %d records.", len(resp.Users))
	}
}

func TestWatch(t *testing.T) {
	// Start a server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcServer, err := StartGrpcServer(ctx, grpcAddr)
	if err != nil {
		t.Fatalf("Error starting GRPC server: %v\n", err)
	}
	defer grpcServer.GracefulStop()

	// Connection for the client
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("TestServer could not create grpc client: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	// Send commands to server [Add, Modify and Remove]
	go func() {
		time.Sleep(time.Millisecond * 200)
		userRec, err := client.AddUser(context.Background(), &pb.AddUserRequest{Email: "another@email.com"})
		if err != nil {
			grpcServer.GracefulStop()
			log.Fatalf("TestWatch could not create a user, %v", err)
		}
		client.ModifyUser(context.Background(), &pb.ModifyUserRequest{
			Id:        userRec.Id,
			FirstName: stringPtr("John"),
		})
		client.RemoveUser(context.Background(), &pb.RemoveUserRequest{
			Id: userRec.Id,
		})
	}()

	// Get stream handler
	stream, err := client.Watch(context.Background(), &pb.WatchRequest{})
	if err != nil {
		t.Fatalf("TestWatch could not watch users: %v", err)
	}

	// Handle stream
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("TestServer error while reading stream: %v", err)
		}

		if msg.Method == pb.WatchResponse_DELETE {
			break
		}
	}
}

func populateDatabase(s UserService) {
	s.AddUser(context.Background(), &pb.AddUserRequest{Email: "john1@email.com", Country: "AU"})
	s.AddUser(context.Background(), &pb.AddUserRequest{Email: "john2@email.com", Country: "AU"})
	s.AddUser(context.Background(), &pb.AddUserRequest{Email: "john3@email.com", Country: "DE"})
	s.AddUser(context.Background(), &pb.AddUserRequest{Email: "john4@email.com", Country: "DE"})
	s.AddUser(context.Background(), &pb.AddUserRequest{Email: "john5@email.com", Country: "NZ"})
	s.AddUser(context.Background(), &pb.AddUserRequest{Email: "john6@email.com", Country: "NZ"})
	s.AddUser(context.Background(), &pb.AddUserRequest{Email: "john7@email.com", Country: "IT"})
}

// Converts string to string ponter
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int32) *int32 {
	return &i
}
