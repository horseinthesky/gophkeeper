package server

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gophkeeper/db/db"
	"gophkeeper/pb"
	"gophkeeper/server/crypto"
	"gophkeeper/server/validation"
)

func (s *Server) Register(ctx context.Context, in *pb.User) (*pb.Token, error) {
	violations := validateUser(in)
	if violations != nil {
		return nil, validation.InvalidArgumentError(violations)
	}

	hashedPassword, err := crypto.HashPassword(in.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	_, err = s.storage.CreateUser(
		ctx,
		db.CreateUserParams{
			Name:     in.Name,
			Passhash: hashedPassword,
		},
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}

	token, err := s.tm.CreateToken(
		in.Name,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token")
	}

	return &pb.Token{Value: token}, nil
}

func (s *Server) Login(ctx context.Context, in *pb.User) (*pb.Token, error) {
	violations := validateUser(in)
	if violations != nil {
		return nil, validation.InvalidArgumentError(violations)
	}

	dbUser, err := s.storage.GetUser(
		ctx,
		in.Name,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find user")
	}

	err = crypto.CheckPassword(in.Password, dbUser.Passhash)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "incorrect password")
	}

	token, err := s.tm.CreateToken(
		dbUser.Name,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token")
	}

	return &pb.Token{Value: token}, nil
}

func validateUser(user *pb.User) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateUsername(user.Name); err != nil {
		violations = append(violations, validation.FieldViolation("username", err))
	}

	if err := validation.ValidatePassword(user.Password); err != nil {
		violations = append(violations, validation.FieldViolation("password", err))
	}

	return violations
}
