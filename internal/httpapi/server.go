package httpapi

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/ledgerly/internal/auth"
	"github.com/example/ledgerly/internal/config"
	"github.com/example/ledgerly/internal/logger"
	"github.com/example/ledgerly/internal/middleware"
	"github.com/example/ledgerly/internal/service"
	"github.com/example/ledgerly/internal/store"
)

// Server wires together all layers and provides an http.Handler.
type Server struct {
	accounts     *service.AccountService
	categories   *service.CategoryService
	transactions *service.TransactionService
	transfers    *service.TransferService
	budgets      *service.BudgetService
	report       *service.ReportService
	auth         *auth.Authenticator
	rateLimiter  *middleware.RateLimiter
	log          *logger.Logger
	cfg          *config.Config
	stores       []interface{ Close() error }
}

// NewServer creates a new Server from stores.
func NewServer(
	accStore *store.MemoryAccountStore,
	catStore *store.MemoryCategoryStore,
	txStore *store.MemoryTransactionStore,
	trStore *store.MemoryTransferStore,
	budStore *store.MemoryBudgetStore,
	log *logger.Logger,
	cfg *config.Config,
) *Server {
	accSvc := service.NewAccountService(accStore)
	catSvc := service.NewCategoryService(catStore)
	txSvc := service.NewTransactionService(txStore, accStore)
	trSvc := service.NewTransferService(trStore, accStore)
	budSvc := service.NewBudgetService(budStore, catStore)
	repSvc := service.NewReportService(txStore, catStore, accStore)
	authSvc := auth.NewAuthenticator()
	rl := middleware.NewRateLimiter(cfg.RateLimit, cfg.RateWindow)

	return &Server{
		accounts:     accSvc,
		categories:   catSvc,
		transactions: txSvc,
		transfers:    trSvc,
		budgets:      budSvc,
		report:       repSvc,
		auth:         authSvc,
		rateLimiter:  rl,
		log:          log,
		cfg:          cfg,
		stores:       []interface{ Close() error }{accStore, catStore, txStore, trStore, budStore},
	}
}

// Handler builds the HTTP handler with middleware.
func (s *Server) Handler() http.Handler {
	accH := NewAccountHandler(s.accounts, s.cfg.MaxBody)
	catH := NewCategoryHandler(s.categories, s.cfg.MaxBody)
	txH := NewTransactionHandler(s.transactions, s.transfers, s.report, s.cfg.MaxBody)
	budH := NewBudgetHandler(s.budgets, s.cfg.MaxBody)
	router := NewRouter(accH, catH, txH, budH)

	var root http.Handler = router.Handler(s.auth)
	root = middleware.RequestID(root)
	root = middleware.Logging(root, s.log)
	root = middleware.Recovery(root, s.log)
	root = s.rateLimiter.Limit(root)
	root = middleware.Timeout(root, s.cfg.Timeout)
	return root
}

// Close shuts down all stores.
func (s *Server) Close() error {
	for _, store := range s.stores {
		if err := store.Close(); err != nil {
			s.log.Warn("store close error", "err", err.Error())
		}
	}
	return nil
}

// RegisterToken creates a token for a user and returns it.
// This is a helper used by cmd/ledgerly to bootstrap an auth token.
func (s *Server) RegisterToken(userID string) (string, error) {
	token, _, err := s.auth.Register(nil, userID)
	return token, err
}

// Listen starts the HTTP server and blocks until os signals.
func (s *Server) Listen() error {
	srv := &http.Server{
		Addr:    s.cfg.Addr,
		Handler: s.Handler(),
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s.log.Info("server start", "addr", s.cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("server error", "err", err.Error())
		}
	}()
	<-quit
	s.log.Info("server shutting down")
	return srv.Close()
}
