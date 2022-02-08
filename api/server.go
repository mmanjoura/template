package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/mmanjoura/template"
	"github.com/mmanjoura/template/race"
	appUser "github.com/mmanjoura/template/user"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// Generic HTTP metrics.
var (
	requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_http_request_count",
		Help: "Total number of requests by route",
	}, []string{"method", "path"})

	requestSeconds = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_http_request_seconds",
		Help: "Total amount of request time by route, in seconds",
	}, []string{"method", "path"})

	errorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_http_error_count",
		Help: "Total number of errors by error code",
	}, []string{"code"})
)

// ShutdownTimeout is the time given for outstanding requests to finish before shutdown.
const ShutdownTimeout = 1 * time.Second

// Server represents an HTTP server. It is meant to wrap all HTTP functionality
// used by the application so that dependent packages (such as cmd/appd) do not
// need to reference the "net/http" package at all.
type Server struct {
	ln     net.Listener
	server *http.Server
	router *mux.Router
	sc     *securecookie.SecureCookie

	// Bind address & domain for the server's listener.
	// If domain is specified, server is run on TLS using acme/autocert.
	Addr   string
	Domain string

	// Keys used for secure cookie encryption.
	HashKey  string
	BlockKey string

	// GitHub OAuth settings.
	GitHubClientID     string
	GitHubClientSecret string

	// Servics used by the various HTTP routes.
	AuthService appUser.AuthService
	// ShopService           user.ShopService
	// ShopMembershipService user.ShopMembershipService
	// EventService          user.EventService
	RaceService race.RaceService
}

// NewServer returns a new instance of Server.
func NewServer() *Server {
	// Create a new server that wraps the net/http server & add a gorilla router.
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter(),
	}

	// Report panics to external service.
	s.router.Use(reportPanic)
	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	// Setup error handling routes.
	s.router.NotFoundHandler = http.HandlerFunc(s.handleNotFound)

	s.router.HandleFunc("/cards", s.FindRacecards).Methods("GET")
	s.router.HandleFunc("/results", s.FindResults).Methods("GET")
	s.router.HandleFunc("/detailInfos", s.FindDetailInfo).Methods("GET")

	return s
}

// UseTLS returns true if the cert & key file are specified.
func (s *Server) UseTLS() bool {
	return s.Domain != ""
}

// Scheme returns the URL scheme for the server.
func (s *Server) Scheme() string {
	if s.UseTLS() {
		return "https"
	}
	return "http"
}

// Port returns the TCP port for the running server.
// This is useful in tests where we allocate a random port by using ":0".
func (s *Server) Port() int {
	if s.ln == nil {
		return 0
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

// URL returns the local base URL of the running server.
func (s *Server) URL() string {
	scheme, port := s.Scheme(), s.Port()

	// Use localhost unless a domain is specified.
	domain := "localhost"
	if s.Domain != "" {
		domain = s.Domain
	}

	// Return without port if using standard ports.
	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return fmt.Sprintf("%s://%s", s.Scheme(), domain)
	}
	return fmt.Sprintf("%s://%s:%d", s.Scheme(), domain, s.Port())
}

// Open validates the server options and begins listening on the bind address.
func (s *Server) Open() (err error) {
	// Initialize our secure cookie with our encryption keys.
	if err := s.openSecureCookie(); err != nil {
		return err
	}

	// Validate GitHub OAuth settings.
	if s.GitHubClientID == "" {
		return fmt.Errorf("github client id required")
	} else if s.GitHubClientSecret == "" {
		return fmt.Errorf("github client secret required")
	}

	// Open a listener on our bind address.
	if s.Domain != "" {
		s.ln = autocert.NewListener(s.Domain)
	} else {
		if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
			return err
		}
	}

	// Begin serving requests on the listener. We use Serve() instead of
	// ListenAndServe() because it allows us to check for listen errors (such
	// as trying to use an already open port) synchronously.
	go s.server.Serve(s.ln)

	return nil
}

// openSecureCookie validates & decodes the block & hash key and initializes
// our secure cookie implementation.
func (s *Server) openSecureCookie() error {
	// Ensure hash & block key are set.
	if s.HashKey == "" {
		return fmt.Errorf("hash key required")
	} else if s.BlockKey == "" {
		return fmt.Errorf("block key required")
	}

	// Decode from hex to byte slices.
	hashKey, err := hex.DecodeString(s.HashKey)
	if err != nil {
		return fmt.Errorf("invalid hash key")
	}
	blockKey, err := hex.DecodeString(s.BlockKey)
	if err != nil {
		return fmt.Errorf("invalid block key")
	}

	// Initialize cookie management & encode our cookie data as JSON.
	s.sc = securecookie.New(hashKey, blockKey)
	s.sc.SetSerializer(securecookie.JSONEncoder{})

	return nil
}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// OAuth2Config returns the GitHub OAuth2 configuration.
func (s *Server) OAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.GitHubClientID,
		ClientSecret: s.GitHubClientSecret,
		Scopes:       []string{},
		Endpoint:     github.Endpoint,
	}
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	// Override method for forms passing "_method" value.
	if r.Method == http.MethodPost {
		switch v := r.PostFormValue("_method"); v {
		case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete:
			r.Method = v
		}
	}

	// Override content-type for certain extensions.
	// This allows us to easily cURL API endpoints with a ".json" or ".csv"
	// extension instead of having to explicitly set Content-type & Accept headers.
	// The extensions are removed so they don't appear in the routes.
	switch ext := path.Ext(r.URL.Path); ext {
	case ".json":
		r.Header.Set("Accept", "application/json")
		r.Header.Set("Content-type", "application/json")
		r.URL.Path = strings.TrimSuffix(r.URL.Path, ext)
	case ".csv":
		r.Header.Set("Accept", "text/csv")
		r.URL.Path = strings.TrimSuffix(r.URL.Path, ext)
	}

	// Delegate remaining HTTP handling to the gorilla router.
	s.router.ServeHTTP(w, r)
}

// reportPanic is middleware for catching panics and reporting them.
func reportPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				template.ReportPanic(err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// handleNotFound handles requests to routes that don't exist.
func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	// tmpl := html.ErrorTemplate{
	// 	StatusCode: http.StatusNotFound,
	// 	Header:     "Your page cannot be found.",
	// 	Message:    "Sorry, it looks like we can't find what you're looking for.",
	// }
	// tmpl.Render(r.Context(), w)
}

// handleIndex handles the "GET /" route. It displays a dashboard with the
// user's shops, recently updated membership values, & a chart.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// If user is not logged in & application is built with a home page,
	// return the home page. Otherwise redirect to login.
	// if app.UserIDFromContext(r.Context()) == 0 {
	// 	if buf := assets.IndexHTML; len(buf) != 0 {
	// 		w.Write(buf)
	// 		return
	// 	}

	// 	http.Redirect(w, r, "/login", http.StatusFound)
	// 	return
	// }

	// var err error
	// var tmpl html.IndexTemplate

	// Fetch all shops the user is a member of.
	// If user is not a member of any shops, redirect to shop list which
	// includes a description of how to start.
	// if tmpl.shops, _, err = s.ShopService.FindShops(r.Context(), app.ShopFilter{}); err != nil {
	// 	Error(w, r, err)
	// 	return
	// } else if len(tmpl.shops) == 0 {
	// 	http.Redirect(w, r, "/shops", http.StatusFound)
	// 	return
	// }

	// Fetch recently updated members.
	// if tmpl.Memberships, _, err = s.ShopMembershipService.FindShopMemberships(r.Context(), app.ShopMembershipFilter{
	// 	Limit:  20,
	// 	SortBy: "updated_at_desc",
	// }); err != nil {
	// 	Error(w, r, err)
	// 	return
	// }

	// Fetch historical average app values.
	// interval := time.Minute
	// end := time.Now().Truncate(interval).Add(interval)
	// start := end.Add(-1 * time.Hour)
	// if tmpl.AverageshopValueReport, err = s.ShopService.AverageShopValueReport(r.Context(), start, end, interval); err != nil {
	// 	Error(w, r, err)
	// 	return
	// }

	// Render the template to the response.
	// tmpl.Render(r.Context(), w)
}

// handleSettings handles the "GET /settings" route.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	// var tmpl html.SettingsTemplate
	// tmpl.Render(r.Context(), w)
}

// handleVersion displays the deployed version.
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(template.Version))
}

// handleVersion displays the deployed commit.
func (s *Server) handleCommit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(template.Commit))
}

// session returns session data from the secure cookie.
func (s *Server) session(r *http.Request) (Session, error) {
	// Read session data from cookie.
	// If it returns an error then simply return an empty session.
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return Session{}, nil
	}

	// Decode session data into a Session object & return.
	var session Session
	if err := s.UnmarshalSession(cookie.Value, &session); err != nil {
		return Session{}, err
	}
	return session, nil
}

// setSession creates a secure cookie with session data.
func (s *Server) setSession(w http.ResponseWriter, session Session) error {
	// Encode session data to JSON.
	buf, err := s.MarshalSession(session)
	if err != nil {
		return err
	}

	// Write cookie to HTTP response.
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    buf,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		Secure:   s.UseTLS(),
		HttpOnly: true,
	})
	return nil
}

// MarshalSession encodes session data to string.
// This is exported to allow the unit tests to generate fake sessions.
func (s *Server) MarshalSession(session Session) (string, error) {
	return s.sc.Encode(SessionCookieName, session)
}

// UnmarshalSession decodes session data into a Session object.
// This is exported to allow the unit tests to generate fake sessions.
func (s *Server) UnmarshalSession(data string, session *Session) error {
	return s.sc.Decode(SessionCookieName, data, &session)
}

// ListenAndServeTLSRedirect runs an HTTP server on port 80 to redirect users
// to the TLS-enabled port 443 server.
func ListenAndServeTLSRedirect(domain string) error {
	return http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+domain, http.StatusFound)
	}))
}

// ListenAndServeDebug runs an HTTP server with /debug endpoints (e.g. pprof, vars).
func ListenAndServeDebug() error {
	h := http.NewServeMux()
	h.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":6060", h)
}

// Error prints & optionally logs an error message.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	// Extract error code & message.
	code, message := template.ErrorCode(err), template.ErrorMessage(err)

	// Track metrics by code.
	errorCount.WithLabelValues(code).Inc()

	// Log & report internal errors.
	if code == template.EINTERNAL {
		template.ReportError(r.Context(), err, r)
		LogError(r, err)
	}

	// Print user message to response based on reqeust accept header.
	switch r.Header.Get("Accept") {
	case "application/json":
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(ErrorStatusCode(code))
		json.NewEncoder(w).Encode(&ErrorResponse{Error: message})

	default:
		w.WriteHeader(ErrorStatusCode(code))
		// tmpl := html.ErrorTemplate{
		// 	StatusCode: ErrorStatusCode(code),
		// 	Header:     "An error has occurred.",
		// 	Message:    message,
		// }
		// tmpl.Render(r.Context(), w)
	}
}

// ErrorStatusCode returns the associated HTTP status code for a app error code.
func ErrorStatusCode(code string) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

// FromErrorStatusCode returns the associated app code for an HTTP status code.
func FromErrorStatusCode(code int) string {
	for k, v := range codes {
		if v == code {
			return k
		}
	}
	return template.EINTERNAL
}

// lookup of application error codes to HTTP status codes.
var codes = map[string]int{
	template.ECONFLICT:       http.StatusConflict,
	template.EINVALID:        http.StatusBadRequest,
	template.ENOTFOUND:       http.StatusNotFound,
	template.ENOTIMPLEMENTED: http.StatusNotImplemented,
	template.EUNAUTHORIZED:   http.StatusUnauthorized,
	template.EINTERNAL:       http.StatusInternalServerError,
}

// LogError logs an error with the HTTP route information.
func LogError(r *http.Request, err error) {
	log.Printf("[http] error: %s %s: %s", r.Method, r.URL.Path, err)
}

// ErrorResponse represents a JSON structure for error output.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SessionCookieName is the name of the cookie used to store the session.
const SessionCookieName = "session"

// Session represents session data stored in a secure cookie.
type Session struct {
	UserID      int    `json:"userID"`
	RedirectURL string `json:"redirectURL"`
	State       string `json:"state"`
}

// SetFlash sets the flash cookie for the next request to read.
func SetFlash(w http.ResponseWriter, s string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    s,
		Path:     "/",
		HttpOnly: true,
	})
}
