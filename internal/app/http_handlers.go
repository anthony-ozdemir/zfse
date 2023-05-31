package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"

	"github.com/anthony-ozdemir/zfse/internal/enum"
)

func (a *Application) helperSendJSONError(w *http.ResponseWriter, jsonErrReply interface{}) {
	jsonRes, err := json.Marshal(jsonErrReply)
	if err != nil {
		zap.L().Fatal(
			"Json Marshal error.",
			zap.String("err", err.Error()),
		)
	}

	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusBadRequest)
	_, err = (*w).Write(jsonRes)
	if err != nil {
		zap.L().Fatal(
			"Unable to write JSON Error message.",
			zap.String("err", err.Error()),
		)
	}

	return
}

func (a *Application) helperSendJSONSuccess(w *http.ResponseWriter, jsonErrReply interface{}) {
	jsonRes, err := json.Marshal(jsonErrReply)
	if err != nil {
		zap.L().Fatal(
			"Json Marshal error.",
			zap.String("err", err.Error()),
		)
	}

	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusOK)
	_, err = (*w).Write(jsonRes)
	if err != nil {
		zap.L().Fatal(
			"Unable to write JSON Success message.",
			zap.String("err", err.Error()),
		)
	}

	return
}

// Helper Handlers
// GoLang HTTP Handlers normally do not crash the entire program when there is a panic inside the HTTP Handler's
// goroutine. However, this is not wanted in our case. We want to gracefully handle the panics inside the goroutine
// instead. Thus, we will let panicking HTTP Handlers crash the application.
func (a *Application) getHTTPCrashOnPanicHandler(next http.HandlerFunc) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				buf := make([]byte, 1<<20)
				n := runtime.Stack(buf, true)
				fmt.Fprintf(os.Stderr, "panic: %v\n\n%s", err, buf[:n])
				os.Exit(1)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(middle)
}

type httpReqInfo struct {
	// GET etc.
	method  string
	uri     string
	referer string
	ipaddr  string
	// response responseCode, like 200, 404
	responseCode int
	// number of bytes of the response sent
	responseSize int64
	// how long did it take to
	duration  time.Duration
	userAgent string
}

// Logs generic info about the HTTP Request.
func (a *Application) getLogHTTPRequestHandler(next http.HandlerFunc) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		ri := &httpReqInfo{
			method:    r.Method,
			uri:       r.URL.String(),
			referer:   r.Header.Get("Referer"),
			userAgent: r.Header.Get("User-Agent"),
		}

		// Design Note: We need to validate that values are non-nil, otherwise zap logging will panic
		if ri.referer == "" {
			ri.referer = "N/A"
		}

		if ri.userAgent == "" {
			ri.userAgent = "N/A"
		}

		// TODO [HP]: Implement, ensure to sanitize properly
		// ri.ipaddr = a.RequestGetRemoteAddress(r)

		// There's no rate-limit error, serve the next handler.
		next.ServeHTTP(w, r)

		// There is a chance that there is no response
		if r.Response != nil {
			ri.responseCode = r.Response.StatusCode
			ri.responseSize = r.Response.ContentLength
		} else {
			ri.responseCode = 0
			ri.responseSize = 0
		}
		ri.duration = time.Since(start)

		zap.L().Info(
			"HTTP REQ",
			zap.String("method", ri.method),
			zap.String("uri", ri.uri),
			zap.String("referer", ri.referer),
			zap.String("ipaddr", ri.ipaddr),
			zap.Int("responseCode", ri.responseCode),
			zap.Int64("responseSize", ri.responseSize),
			zap.Duration("duration", ri.duration),
			zap.String("useragent", ri.userAgent),
		)
	}

	return http.HandlerFunc(middle)

}

func (a *Application) getCommonWrapperHandler(next http.HandlerFunc) http.Handler {
	logHandler := a.getLogHTTPRequestHandler(next)
	crashOnPanicHandler := a.getHTTPCrashOnPanicHandler(logHandler.ServeHTTP)

	return crashOnPanicHandler
}

// API Methods
/*
// TODO [HP]: Implement
func (a *Application) statusGetHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		zap.L().Info("Received status request.")

		// Headers
		w.Header().Set("Vary", "*")                 // Hint uncacheable
		w.Header().Set("Cache-Control", "no-store") // No cache of any kind (private or shared)
		// Access-Control-Allow-Credentials
		// Design Note: The only valid value for this header is true (case-sensitive). If you don't need credentials,
		// omit this header entirely (rather than setting its value to false).
		// w.Header().Set("Access-Control-Allow-Credentials", "true")
		// Access-Control-Allow-Headers
		// Design Note: Can be set to '*' only when Access-Control-Allow-Credentials is not set. Otherwise requires
		// manual specification.
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		} else if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	}
}
*/

type RankingQueryReqJSON struct {
	UserQuery *string `json:"user_query"` // pointer so we can test for field absence
}

type RankingQueryRepJSON struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (a *Application) rankerQueryHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO [HP]: Implement
		zap.L().Info("Received status request.")

		// Headers
		w.Header().Set("Vary", "*")                 // Hint uncacheable
		w.Header().Set("Cache-Control", "no-store") // No cache of any kind (private or shared)
		// Access-Control-Allow-Credentials
		// Design Note: The only valid value for this header is true (case-sensitive). If you don't need credentials,
		// omit this header entirely (rather than setting its value to false).
		// w.Header().Set("Access-Control-Allow-Credentials", "true")
		// Access-Control-Allow-Headers
		// Design Note: Can be set to '*' only when Access-Control-Allow-Credentials is not set. Otherwise requires
		// manual specification.
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		} else if r.Method == http.MethodPost {
			// Try decode Json request
			decoder := json.NewDecoder(r.Body)
			jsonReq := RankingQueryReqJSON{}

			errJsonDecode := decoder.Decode(&jsonReq)
			if errJsonDecode != nil {
				jsonErr := RankingQueryRepJSON{
					Success: false,
					Error:   "json decode error",
				}
				a.helperSendJSONError(&w, jsonErr)
				return
			}

			// Check existence of mandatory fields on Json request
			if jsonReq.UserQuery == nil {
				jsonErr := RankingQueryRepJSON{
					Success: false,
					Error:   "json field missing",
				}
				a.helperSendJSONError(&w, jsonErr)
				return
			}

			// Let's check if server is ready to Rank
			// TODO [HP]: We should lock the state mutex here, so that large amount of
			// rank requests won't create a data race condition
			currentApplicationState := a.applicationStateManager.GetApplicationState()
			if currentApplicationState.task != enum.ReadyToSearch {
				jsonErr := RankingQueryRepJSON{
					Success: false,
					Error:   "index not yet ready",
				}
				a.helperSendJSONError(&w, jsonErr)
				return
			}

			// Let's start ranking
			a.Rank(*jsonReq.UserQuery)

			// Everything seems fine. Let's reply back with success
			jsonRep := RankingQueryRepJSON{
				Success: true,
			}
			a.helperSendJSONSuccess(&w, jsonRep)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	}
}
