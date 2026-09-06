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
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/keygen"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"github.com/longkun/ssh-blog/internal/blog"
	"github.com/longkun/ssh-blog/internal/ui"
)

const (
	defaultHost       = "0.0.0.0"
	defaultPort       = 2222
	defaultDir        = "posts"
	defaultPagesDir   = "pages"
	defaultWalineURL  = "https://waline-comments.happy365.day"
	defaultWalinePath = "messages/index.html"
)

func defaultHostKeyPath() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".ssh", "ssh-blog_hostkey")
	}
	return "ssh-blog_hostkey"
}

func main() {
	var (
		host       string
		port       int
		posts      string
		pagesDir   string
		walineURL  string
		keyPath    string
		version    bool
	)

	flag.StringVar(&host, "host", defaultHost, "address to listen on")
	flag.IntVar(&port, "port", defaultPort, "port to listen on")
	flag.StringVar(&posts, "posts", defaultDir, "directory containing markdown posts")
	flag.StringVar(&pagesDir, "pages", defaultPagesDir, "directory containing static pages (about.md, friends.md, …)")
	flag.StringVar(&walineURL, "waline", defaultWalineURL, "waline API base URL for the message board")
	flag.StringVar(&keyPath, "key", defaultHostKeyPath(), "path to SSH host key (default: ~/.ssh/ssh-blog_hostkey)")
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.Parse()

	if version {
		fmt.Println("ssh-blog 0.2.0")
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
			bubbletea.Middleware(teaHandler(absPosts, absPages, walineURL)),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	log.Printf("ssh-blog listening on %s:%d, posts=%s, pages=%s", host, port, absPosts, absPages)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.Shutdown(ctx)
}

func teaHandler(postsDir, pagesDir, walineURL string) bubbletea.Handler {
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
			waline = blog.NewWalineClient(walineURL)
		}
		return ui.New(ui.Config{Posts: posts, Pages: pages, Waline: waline}),
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
