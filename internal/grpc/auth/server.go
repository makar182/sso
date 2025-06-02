package auth

import (
	"context"
	ssov1 "github.com/makar182/protos/gen/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const emptyValue = 0

type Auth interface {
	Login(ctx context.Context, email string, password string, appId int) (string, error)
	Logout(ctx context.Context, token string) (bool, error)
	Register(ctx context.Context, email string, password string) (int64, error)
	IsAdmin(ctx context.Context, userId int64) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func RegisterServerAPI(srv *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(srv, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" || req.GetAppId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "email, password and app_id must be provided")
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.LogoutResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token must be provided")
	}
	isLoggedOut, err := s.auth.Logout(ctx, req.GetToken())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to logout: %v", err)
	}
	return &ssov1.LogoutResponse{
		IsLoggedOut: isLoggedOut,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password must be provided")
	}

	userId, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register: %v", err)
	}

	res := &ssov1.RegisterResponse{
		UserId: userId,
	}
	return res, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if req.GetUserId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "userId must be provided")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check admin status: %v", err)
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}
