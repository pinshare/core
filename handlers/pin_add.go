package handlers

import (
	"time"

	"github.com/pinshare/config"
	"github.com/pinshare/spec/service"
	"github.com/pinshare/syncker"
	"github.com/ziutek/mymysql/mysql"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type addPinServer struct {
	conf   *config.Config
	logger *zap.SugaredLogger
	serviceInterface
}

func NewAddPinServer() *addPinServer {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return &addPinServer{
		logger: logger.Sugar(),
	}
}

func init() {
	addService(NewAddPinServer())
}

func (s *addPinServer) Name() string {
	return "AddPinServer"
}

func (s *addPinServer) Register(server *grpc.Server, config *config.Config) error {
	s.conf = config
	service.RegisterAddPinServer(server, s)
	return nil
}

func (s *addPinServer) Execute(ctx context.Context, request *service.AddRequest) (*service.PinResponse, error) {
	db, err := s.conf.MySQL.Connect()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	tr, err := db.Begin()
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	createPin, _ := db.Prepare("INSERT INTO pins VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}
	meta, err := tr.Do(createPin).Run(
		1,
		request.GetTitle(),
		request.GetDescription(),
		request.GetPhrase(),
		request.GetUrl(),
		now.Unix(),
		now.Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		s.logger.Error(err)
		tr.Rollback()
		return nil, err
	}

	tagIds := make([]int, 0)
	createTag, err := db.Prepare("INSERT INTO tags VALUES (NULL, ?, ?, ?)")
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}
	ti := tr.Do(createTag)
	for _, tag := range request.GetTags() {
		if id, exists := s.findByTagName(db, tag); exists {
			tagIds = append(tagIds, id)
			continue
		}
		meta, err := ti.Run(
			tag,
			now.Format("2006-01-02 15:04:05"),
			now.Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			s.logger.Error(err)
			tr.Rollback()
			return nil, err
		}
		tagIds = append(tagIds, int(meta.InsertId()))
	}

	relPinTags, err := db.Prepare("INSERT INTO rel_pin_tags VALUES (?, ?, ?, ?)")
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	pinId := meta.InsertId()
	ti = tr.Do(relPinTags)
	for _, tagId := range tagIds {
		_, err := ti.Run(
			pinId,
			tagId,
			now.Format("2006-01-02 15:04:05"),
			now.Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			s.logger.Error(err)
			tr.Rollback()
			return nil, err
		}
	}

	tr.Commit()

	resp := &service.PinResponse{
		Id:          int32(pinId),
		UserId:      request.GetUserId(),
		Title:       request.GetTitle(),
		Url:         request.GetUrl(),
		Phrase:      request.GetPhrase(),
		Timestamp:   request.GetTimestamp(),
		Description: request.GetDescription(),
		Tags:        request.GetTags(),
	}

	go syncker.SyncRow(s.conf, resp)
	return resp, nil
}

func (s *addPinServer) findByTagName(db mysql.Conn, tagName string) (tagId int, exists bool) {
	stmt, _ := db.Prepare("SELECT id FROM tags WHERE name = ? LIMIT 1")
	row, _, err := stmt.ExecFirst(tagName)
	if err != nil {
		s.logger.Info(err)
		return 0, false
	} else if len(row) == 0 {
		return 0, false
	}
	return row.Int(0), true
}
