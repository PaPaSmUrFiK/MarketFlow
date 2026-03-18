package appgrpc

import (
	"fmt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/lib/sl"
	admingrpc "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/transport/grpc/admin"
	authgrpc "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/transport/grpc/auth"
	usergrpc "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/transport/grpc/user"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"strconv"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger,
	authService authgrpc.Service,
	userService usergrpc.Service,
	adminService admingrpc.Service,
	port int,
) *App {
	gRPCServer := grpc.NewServer()
	authgrpc.Register(gRPCServer, authService)
	usergrpc.Register(gRPCServer, userService)
	admingrpc.Register(gRPCServer, adminService)
	return &App{log, gRPCServer, port}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "appgrpc.Run"
	log := a.log.With(slog.String("operation", op),
		slog.String("port", strconv.Itoa(a.port)),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return sl.Err(op, err)
	}

	log.Info("starting gRPC server ", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return sl.Err(op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "appgrpc.Stop"

	a.log.With(slog.String("operation", op)).Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
