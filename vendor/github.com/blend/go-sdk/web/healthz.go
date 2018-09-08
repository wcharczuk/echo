package web

import (
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

const (
	// VarzStarted is a common variable.
	VarzStarted = "startedUTC"
	// VarzRequests is a common variable.
	VarzRequests = "http_requests"
	// VarzRequests2xx is a common variable.
	VarzRequests2xx = "http_requests2xx"
	// VarzRequests3xx is a common variable.
	VarzRequests3xx = "http_requests3xx"
	// VarzRequests4xx is a common variable.
	VarzRequests4xx = "http_requests4xx"
	// VarzRequests5xx is a common variable.
	VarzRequests5xx = "http_requests5xx"
	// VarzErrors is a common variable.
	VarzErrors = "errors_total"
	// VarzFatals is a common variable.
	VarzFatals = "fatals_total"

	// ListenerHealthz is the uid of the healthz logger listeners.
	ListenerHealthz = "healthz"

	// ErrHealthzAppUnset is a common error.
	ErrHealthzAppUnset exception.Class = "healthz app unset"
)

// NewHealthz returns a new healthz.
func NewHealthz(monitored *App) *Healthz {
	return &Healthz{
		monitored:      monitored,
		defaultHeaders: map[string]string{},
		state:          State{},
		vars: State{
			VarzRequests:    int64(0),
			VarzRequests2xx: int64(0),
			VarzRequests3xx: int64(0),
			VarzRequests4xx: int64(0),
			VarzRequests5xx: int64(0),
			VarzErrors:      int64(0),
			VarzFatals:      int64(0),
		},
	}
}

// NewHealthzFromEnv returns a new healthz from the env.
func NewHealthzFromEnv(monitored *App) *Healthz {
	return NewHealthzFromConfig(monitored, NewHealthzConfigFromEnv())
}

// NewHealthzFromConfig returns a new healthz sidecar from a config.
func NewHealthzFromConfig(monitored *App, cfg *HealthzConfig) *Healthz {
	hz := NewHealthz(monitored)
	hz = hz.WithBindAddr(cfg.GetBindAddr())
	hz = hz.WithRecoverPanics(cfg.GetRecoverPanics())
	hz = hz.WithMaxHeaderBytes(cfg.GetMaxHeaderBytes())
	hz = hz.WithReadHeaderTimeout(cfg.GetReadHeaderTimeout())
	hz = hz.WithReadTimeout(cfg.GetReadTimeout())
	hz = hz.WithWriteTimeout(cfg.GetWriteTimeout())
	hz = hz.WithIdleTimeout(cfg.GetIdleTimeout())
	return hz
}

// Healthz is a sentinel / healthcheck sidecar that can run on a different
// port to the main app.
// It typically implements the following routes:
// 	/healthz - overall health endpoint, 200 on healthy, 5xx on not.
// 	/varz    - basic stats and metrics since start
//	/debug/vars - `pkg/expvar` output.
type Healthz struct {
	monitored  *App
	startedUTC time.Time
	bindAddr   string
	log        *logger.Logger

	defaultHeaders map[string]string
	server         *http.Server
	listener       *net.TCPListener

	maxHeaderBytes    int
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration

	state State

	varsLock sync.Mutex
	vars     State

	recoverPanics bool
	err           error
}

// Monitored returns the underlying app.
func (hz *Healthz) Monitored() *App {
	return hz.monitored
}

// Vars returns the underlying vars collection.
func (hz *Healthz) Vars() State {
	return hz.vars
}

// RecoverPanics returns if the app recovers panics.
func (hz *Healthz) RecoverPanics() bool {
	return hz.recoverPanics
}

// WithRecoverPanics sets if the app should recover panics.
func (hz *Healthz) WithRecoverPanics(value bool) *Healthz {
	hz.recoverPanics = value
	return hz
}

// MaxHeaderBytes returns the app max header bytes.
func (hz *Healthz) MaxHeaderBytes() int {
	return hz.maxHeaderBytes
}

// WithMaxHeaderBytes sets the max header bytes value and returns a reference.
func (hz *Healthz) WithMaxHeaderBytes(byteCount int) *Healthz {
	hz.maxHeaderBytes = byteCount
	return hz
}

// ReadHeaderTimeout returns the read header timeout for the server.
func (hz *Healthz) ReadHeaderTimeout() time.Duration {
	return hz.readHeaderTimeout
}

// WithReadHeaderTimeout returns the read header timeout for the server.
func (hz *Healthz) WithReadHeaderTimeout(timeout time.Duration) *Healthz {
	hz.readHeaderTimeout = timeout
	return hz
}

// ReadTimeout returns the read timeout for the server.
func (hz *Healthz) ReadTimeout() time.Duration {
	return hz.readTimeout
}

// WithReadTimeout sets the read timeout for the server and returns a reference to the app for building apps with a fluent api.
func (hz *Healthz) WithReadTimeout(timeout time.Duration) *Healthz {
	hz.readTimeout = timeout
	return hz
}

// IdleTimeout is the time before we close a connection.
func (hz *Healthz) IdleTimeout() time.Duration {
	return hz.idleTimeout
}

// WithIdleTimeout sets the idle timeout.
func (hz *Healthz) WithIdleTimeout(timeout time.Duration) *Healthz {
	hz.idleTimeout = timeout
	return hz
}

// WriteTimeout returns the write timeout for the server.
func (hz *Healthz) WriteTimeout() time.Duration {
	return hz.writeTimeout
}

// WithWriteTimeout sets the write timeout for the server and returns a reference to the app for building apps with a fluent api.
func (hz *Healthz) WithWriteTimeout(timeout time.Duration) *Healthz {
	hz.writeTimeout = timeout
	return hz
}

// WithPort sets the port for the bind address of the app, and returns a reference to the app.
func (hz *Healthz) WithPort(port int32) *Healthz {
	hz.SetPort(port)
	return hz
}

// SetPort sets the port the app listens on, typically to `:%d` which indicates listen on any interface.
func (hz *Healthz) SetPort(port int32) {
	hz.bindAddr = fmt.Sprintf(":%v", port)
}

// WithPortFromEnv sets the port from an environment variable, and returns a reference to the app.
func (hz *Healthz) WithPortFromEnv() *Healthz {
	hz.SetPortFromEnv()
	return hz
}

// SetPortFromEnv sets the port from an environment variable, and returns a reference to the app.
func (hz *Healthz) SetPortFromEnv() {
	if env.Env().Has(EnvironmentVariablePort) {
		port, err := env.Env().Int32(EnvironmentVariablePort)
		if err != nil {
			hz.err = err
		}
		hz.bindAddr = fmt.Sprintf(":%v", port)
	}
}

// BindAddr returns the address the server will bind to.
func (hz *Healthz) BindAddr() string {
	return hz.bindAddr
}

// WithBindAddr sets the address the app listens on, and returns a reference to the app.
func (hz *Healthz) WithBindAddr(bindAddr string) *Healthz {
	hz.bindAddr = bindAddr
	return hz
}

// WithBindAddrFromEnv sets the address the app listens on, and returns a reference to the app.
func (hz *Healthz) WithBindAddrFromEnv() *Healthz {
	hz.bindAddr = env.Env().String(EnvironmentVariableBindAddr)
	return hz
}

// Logger returns the diagnostics agent for the app.
func (hz *Healthz) Logger() *logger.Logger {
	return hz.log
}

// WithLogger sets the app logger agent and returns a reference to the app.
// It also sets underlying loggers in any child resources like providers and the auth manager.
func (hz *Healthz) WithLogger(log *logger.Logger) *Healthz {
	hz.log = log
	return hz
}

// Server returns the basic http.Server for the healthz host.
func (hz *Healthz) Server() *http.Server {
	hz.vars[VarzStarted] = time.Now().UTC()

	if hz.monitored.log != nil {
		hz.monitored.log.Listen(logger.HTTPResponse, ListenerHealthz, logger.NewHTTPResponseEventListener(hz.httpResponseListener))
		hz.monitored.log.Listen(logger.Error, ListenerHealthz, logger.NewErrorEventListener(hz.errorListener))
		hz.monitored.log.Listen(logger.Fatal, ListenerHealthz, logger.NewErrorEventListener(hz.errorListener))
	}

	return &http.Server{
		Addr:              hz.BindAddr(),
		Handler:           hz,
		MaxHeaderBytes:    hz.maxHeaderBytes,
		ReadTimeout:       hz.readTimeout,
		ReadHeaderTimeout: hz.readHeaderTimeout,
		WriteTimeout:      hz.writeTimeout,
		IdleTimeout:       hz.idleTimeout,
	}
}

// ServeHTTP makes the router implement the http.Handler interface.
func (hz *Healthz) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if hz.recoverPanics {
		defer hz.recover(w, r)
	}

	res := NewRawResponseWriter(w)
	res.Header().Set(HeaderContentEncoding, ContentEncodingIdentity)

	route := strings.ToLower(r.URL.Path)

	start := time.Now()
	if hz.log != nil {
		hz.log.Trigger(logger.NewHTTPResponseEvent(r).WithState(hz.state).WithRoute(route))

		defer func() {
			hz.log.Trigger(logger.NewHTTPResponseEvent(r).
				WithStatusCode(res.StatusCode()).
				WithElapsed(time.Since(start)).
				WithContentLength(res.ContentLength()).
				WithState(hz.state))
		}()
	}

	if len(hz.defaultHeaders) > 0 {
		for key, value := range hz.defaultHeaders {
			res.Header().Set(key, value)
		}
	}

	switch route {
	case "/healthz":
		hz.healthzHandler(res, r)
	case "/varz":
		hz.varzHandler(res, r)
	default:
		http.NotFound(res, r)
	}

	if err := res.Close(); err != nil && err != http.ErrBodyNotAllowed && hz.log != nil {
		hz.log.Error(err)
	}
}

func (hz *Healthz) recover(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		if hz.log != nil {
			hz.log.Fatalf("%v", rcv)
		}

		http.Error(w, fmt.Sprintf("%v", rcv), http.StatusInternalServerError)
		return
	}
}

func (hz *Healthz) healthzHandler(w ResponseWriter, r *http.Request) {
	if hz.monitored.Latch().IsRunning() {
		w.WriteHeader(http.StatusOK)
		w.Header().Set(HeaderContentType, ContentTypeText)
		fmt.Fprintf(w, "OK!\n")
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set(HeaderContentType, ContentTypeText)
	fmt.Fprintf(w, "Failure!\n")
	return
}

// /varz
// writes out the current stats
func (hz *Healthz) varzHandler(w ResponseWriter, r *http.Request) {
	hz.varsLock.Lock()
	defer hz.varsLock.Unlock()

	keys := make([]string, len(hz.vars))

	var index int
	for key := range hz.vars {
		keys[index] = fmt.Sprintf("%v", key)
		index++
	}

	sort.Strings(keys)

	w.WriteHeader(http.StatusOK)
	w.Header().Set(HeaderContentType, ContentTypeText)
	for _, key := range keys {
		fmt.Fprintf(w, "%s: %v\n", key, hz.vars[key])
	}
}

func (hz *Healthz) httpResponseListener(wre *logger.HTTPResponseEvent) {
	hz.varsLock.Lock()
	defer hz.varsLock.Unlock()

	hz.incrementVarUnsafe(VarzRequests)
	if wre.StatusCode() >= http.StatusInternalServerError {
		hz.incrementVarUnsafe(VarzRequests5xx)
	} else if wre.StatusCode() >= http.StatusBadRequest {
		hz.incrementVarUnsafe(VarzRequests4xx)
	} else if wre.StatusCode() >= http.StatusMultipleChoices {
		hz.incrementVarUnsafe(VarzRequests3xx)
	} else {
		hz.incrementVarUnsafe(VarzRequests2xx)
	}
}

func (hz *Healthz) errorListener(e *logger.ErrorEvent) {
	hz.varsLock.Lock()
	defer hz.varsLock.Unlock()

	switch e.Flag() {
	case logger.Error:
		hz.incrementVarUnsafe(VarzErrors)
		return
	case logger.Fatal:
		hz.incrementVarUnsafe(VarzFatals)
		return
	}
}

func (hz *Healthz) incrementVarUnsafe(key string) {
	if value, hasValue := hz.vars[key]; hasValue {
		if typed, isTyped := value.(int64); isTyped {
			hz.vars[key] = typed + 1
		}
	} else {
		hz.vars[key] = int64(1)
	}
}
