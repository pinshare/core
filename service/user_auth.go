package service

import (
	"fmt"
	"strings"
	"time"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"github.com/pinshare/config"
	"github.com/pinshare/core/lib"
	"github.com/pinshare/spec/service"
	"github.com/ziutek/mymysql/mysql"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type userAuthService struct {
	conf   *config.Config
	logger *zap.SugaredLogger
	serviceInterface
}

func NewUserAuthService() *userAuthService {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return &userAuthService{
		logger: logger.Sugar(),
	}
}

func init() {
	addService(NewUserAuthService())
}

func (s *userAuthService) Name() string {
	return "UserAuthService"
}

func (s *userAuthService) Register(server *grpc.Service, config *config.Config) error {
	s.conf = config
	service.RegisterUserAuthService(server, s)
	return nil
}

func (s *userAuthService) Execute(ctx context.Context, request *service.UserAuthRequest) (*service.UserResponse, error) {
	db, err := s.conf.MySQL.Connect()
	if err != nil {
		return nil, err
	}

	if userId, token, ok, err := s.userExists(db, request); err != nil {
		s.logger.Error(err)
		return nil, err
	} else if ok {
		return &service.UserResponse{
			Id:             userId,
			Name:           request.GetGithubName(),
			Token:          token,
			ProfileIconUrl: request.GetGithuProfileIconUrl(),
		}, nil
	}

	if userId, token, err := s.registerUser(db, request); err != nil {
		s.logger.Error(err)
		return nil, err
	} else {
		return &service.UserResponse{
			Id:             userId,
			Name:           request.GetGithubName(),
			Token:          token,
			ProfileIconUrl: request.GetGithuProfileIconUrl(),
		}, nil
	}
}

func (s *userAuthService) userExists(db mysql.Conn, githubId string) (userId, token string, exists bool, err error) {
	var row mysql.Row
	var stmt mysql.Stmt
	stmt, _ = db.Prepare("SELECT id, user_id, token FROM users WHERE github_id = ? LIMIT 1")
	row, _, err = stmt.ExecFirst(githubId)
	if err != nil {
		return
	} else if len(row) == 0 {
		return
	}

	exists = true
	id := row.Int(0)

	// update
	stmt, _ = db.Prepare("UPDATE users SET profile_url = ? WHERE id = ? LIMIT 1")
	if _, err = stmt.Run(id); err != nil {
		return
	}

	userId = row.String(1)
	token = row.String(2)
	return
}

func (s *userAuthService) registerUser(db mysql.Conn, request *serive.UserAuthRequest) (userId, token string, err error) {
	userId = fmt.Sprintf("%s@%s", request.GetGithubName(), "localhost")
	token = lib.GenerateToken(userId)
	now := time.Now().Format("2006-01-02 15:04:05")

	sql := fmt.Sprintf("INSERT INTO %s VALUES %s",
		"(id, user_id, github_id, token, profile_url, username, created_at, updated_at)",
		"(NULL, ?, ?, ?, ?, ?, ?, ?)",
	)

	stmt, _ := db.Prepare(sql)
	_, err = stmt.Run(
		userId,
		token,
		request.GetGithuProfileIconUrl(),
		request.GetGithubName(),
		now,
		now,
	)
	return
}
