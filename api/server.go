package api

import (
	"block_chain/core"
	"github.com/go-kit/log"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type ServerConfig struct {
	Logger     log.Logger
	ListenAddr string
}

type Server struct {
	ServerConfig
	bc *core.Blockchain
}

func NewServer(cfg ServerConfig, bc *core.Blockchain) *Server {
	return &Server{
		ServerConfig: cfg,
		bc:           bc,
	}
}

func (s *Server) Start() error {
	e := echo.New()

	e.GET("/blocks/:hashorid", s.handleGetBlock)

	s.Logger.Log("msg", "Starting server", "addr", s.ListenAddr)

	if err := e.Start(s.ListenAddr); err != nil {
		s.Logger.Log("error", err)
		return err
	}
	return nil
}

func (s *Server) handleGetBlock(e echo.Context) error {
	hashOrID := e.Param("hashorid")

	height, err := strconv.Atoi(hashOrID)
	if err == nil {
		block, err := s.bc.GetBlock(uint32(height))
		if err != nil {
			return err
		}
		return e.JSON(http.StatusOK, block)
	}

	//panic("need to")
	return e.JSON(http.StatusOK, map[string]any{"msg": "it works!"})
}
