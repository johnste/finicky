package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"finicky/window"
)

var headlessFlag = flag.Bool("headless", false, "Run without a window and print the UI's address instead (Linux)")

// instanceListener receives URLs from later launches while this instance
// runs, so links open without starting a second Finicky.
var instanceListener net.Listener

// runApp handles the URLs Finicky was launched with, then runs the GTK main
// loop on the main thread. Launched without a URL, it opens the window, since
// there's no menu bar icon to open it from on Linux.
func runApp(forceOpenWindow bool, showStatusItem bool, keepRunning bool) {
	window.Headless = *headlessFlag
	window.OnClose = windowDidClose
	window.OnOpenURL = func(url string) { handleURL(url, nil, false) }

	// Remove the instance socket when stopped from a terminal or by the
	// session, not only when exiting normally.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		platformCleanup()
		os.Exit(0)
	}()

	urls := flag.Args()
	go func() {
		for _, url := range urls {
			handleURL(url, nil, false)
		}
		if len(urls) == 0 || forceOpenWindow {
			queueWindowDisplay(true)
		}
	}()

	window.Run()
}

func setStatusItemError(hasError bool) {}

// platformPreflight makes this the single running instance, or hands this
// launch's URLs (or a request to show the window) to the instance that
// already runs and returns handled = true. In devMode, or when headless, it
// always runs as its own instance so it can't pick up the user's real one.
func platformPreflight(devMode bool) (handled bool) {
	if devMode || *headlessFlag {
		return false
	}

	path := instanceSocketPath()
	ln, err := net.Listen("unix", path)
	if err != nil {
		// The socket exists: either another instance is running, or it's
		// left over from one that crashed.
		if conn, dialErr := net.DialTimeout("unix", path, time.Second); dialErr == nil {
			defer conn.Close()
			forwardToInstance(conn)
			return true
		}
		os.Remove(path)
		if ln, err = net.Listen("unix", path); err != nil {
			slog.Warn("Failed to listen for other launches; running as a separate instance", "error", err)
			return false
		}
	}
	os.Chmod(path, 0600)
	instanceListener = ln
	go acceptInstanceMessages(ln)
	return false
}

func platformCleanup() {
	if instanceListener != nil {
		instanceListener.Close()
	}
}

func instanceSocketPath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		return filepath.Join(os.TempDir(), fmt.Sprintf("finicky-%d.sock", os.Getuid()))
	}
	return filepath.Join(dir, "finicky.sock")
}

// forwardToInstance sends one line per URL ("open <url>"), or "show" when
// launched without URLs or with -window.
func forwardToInstance(conn net.Conn) {
	urls := flag.Args()
	for _, url := range urls {
		fmt.Fprintf(conn, "open %s\n", url)
	}
	if len(urls) == 0 || flag.Lookup("window").Value.String() == "true" {
		fmt.Fprintln(conn, "show")
	}
	slog.Debug("Handed off to running instance", "urls", len(urls))
}

func acceptInstanceMessages(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go func() {
			defer conn.Close()
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				line := scanner.Text()
				switch {
				case line == "show":
					queueWindowDisplay(true)
				case strings.HasPrefix(line, "open "):
					handleURL(strings.TrimPrefix(line, "open "), nil, false)
				}
			}
		}()
	}
}
