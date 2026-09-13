package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/keygen"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"github.com/longkun/ssh-blog/internal/blog"
	"github.com/longkun/ssh-blog/internal/config"
	"github.com/longkun/ssh-blog/internal/ui"
)

const (
	binaryName = "ssh-blog"
)

func defaultHostKeyPath(homeOverride string) string {
	if homeOverride == "" {
		if home, err := os.UserHomeDir(); err == nil {
			homeOverride = home
		}
	}
	if homeOverride == "" {
		return "ssh-blog_hostkey"
	}
	return filepath.Join(homeOverride, ".ssh", "ssh-blog_hostkey")
}

func main() {
	var (
		configPath string
		host       string
		port       int
		posts      string
		pagesDir   string
		walineURL  string
		keyPath    string
		version    bool
	)
	flag.StringVar(&configPath, "config", "", "path to config.yaml (default: ./config.yaml or ~/.config/ssh-blog/config.yaml)")
	flag.StringVar(&host, "host", "", "address to listen on (overrides config)")
	flag.IntVar(&port, "port", 0, "port to listen on (overrides config)")
	flag.StringVar(&posts, "posts", "", "directory containing markdown posts (overrides config)")
	flag.StringVar(&pagesDir, "pages", "", "directory containing static pages (overrides config)")
	flag.StringVar(&walineURL, "waline", "", "waline API base URL (overrides config; empty disables)")
	flag.StringVar(&keyPath, "key", "", "path to SSH host key (overrides config)")
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if host == "" {
		host = cfg.Server.DefaultHost
	}
	if port == 0 {
		port = cfg.Server.DefaultPort
	}
	if posts == "" {
		posts = cfg.Server.DefaultPostsDir
	}
	if pagesDir == "" {
		pagesDir = cfg.Server.DefaultPagesDir
	}
	if keyPath == "" {
		keyPath = defaultHostKeyPath("")
	}
	if walineURL == "" {
		walineURL = cfg.Waline.BaseURL
	}

	if version {
		fmt.Println(cfg.Site.Version)
		return
	}

	absPosts, err := filepath.Abs(posts)
	if err != nil {
		log.Fatalf("resolve posts dir: %v", err)
	}
	if _, err := os.Stat(absPosts); err != nil {
		log.Fatalf("posts dir %q: %v", absPosts, err)
	}
	absPages, err := filepath.Abs(pagesDir)
	if err != nil {
		log.Fatalf("resolve pages dir: %v", err)
	}

	hostKey, err := loadOrGenerateHostKey(keyPath)
	if err != nil {
		log.Fatalf("host key: %v", err)
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, fmt.Sprintf("%d", port))),
		wish.WithHostKeyPEM(hostKey),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler(cfg, absPosts, absPages, walineURL)),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	log.Printf(cfg.Server.ListenBanner, host, port, absPosts, absPages)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println(cfg.Server.ShutdownMessage)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	_ = s.Shutdown(ctx)
}

func teaHandler(cfg *config.Config, postsDir, pagesDir, walineURL string) bubbletea.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		posts, err := blog.LoadDir(postsDir)
		if err != nil {
			log.Printf("load posts for %s: %v", s.User(), err)
		}
		pages, err := blog.LoadPages(pagesDir)
		if err != nil {
			log.Printf("load pages for %s: %v", s.User(), err)
		}
		var waline *blog.WalineClient
		if walineURL != "" {
			waline = blog.NewWalineClient(blog.WalineOptions{
				BaseURL:  walineURL,
				Timeout:  cfg.Waline.Timeout,
				PageSize: cfg.Waline.PageSize,
			})
		}
		return ui.New(ui.Config{
			Cfg:    cfg,
			Posts:  posts,
			Pages:  pages,
			Waline: waline,
		}),
			[]tea.ProgramOption{tea.WithAltScreen()}
	}
}

func loadOrGenerateHostKey(path string) ([]byte, error) {
	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			return data, nil
		}
		log.Printf("generating new host key at %s", path)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, err
		}
		kp, err := keygen.New(path, keygen.WithKeyType(keygen.Ed25519), keygen.WithWrite())
		if err != nil {
			return nil, err
		}
		if err := kp.WriteKeys(); err != nil {
			return nil, err
		}
		return kp.RawPrivateKey(), nil
	}
	kp, err := keygen.New("", keygen.WithKeyType(keygen.Ed25519))
	if err != nil {
		return nil, err
	}
	return kp.RawPrivateKey(), nil
}
