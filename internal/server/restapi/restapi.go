package restapi

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	chiprometheus "github.com/766b/chi-prometheus"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swaggest/swgui/v3cdn"

	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/client"
	"github.com/thomasbuchinger/timerec/internal/server"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
)

//go:embed openapi.yaml
var openapiContent []byte

func Run(mgr *server.TimerecServer) {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(chiprometheus.NewMiddleware("timerec", 10, 50, 100, 1000, 5000))

	r.Use(middleware.Timeout(60 * time.Second))

	mountUtils(r)
	mountTextApi(r, mgr)
	mountUserApi(r, mgr)
	mountActivityApi(r, mgr)
	mountJobApi(r, mgr)

	mgr.Logger.Infof("Started Webserver on %s", mgr.BindAddress)
	err := http.ListenAndServe(mgr.BindAddress, r)
	if err != nil {
		mgr.Logger.Errorf("cannot start Server: %v", err)
		os.Exit(1)
	}
}

func mountUserApi(r *chi.Mux, mgr *server.TimerecServer) {
	api := chi.NewRouter()
	api.Use(middleware.Logger)
	api.Use(middleware.AllowContentType("application/json"))
	api.Use(middleware.SetHeader("Content-Type", "application/json"))

	api.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "user")

		resp, err := mgr.GetUser(r.Context(), server.SearchUserParams{Name: name})
		ObjectToJsonBytes(r.Context(), rw, resp, err)
	})
	api.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		params := server.SearchUserParams{
			Name:     chi.URLParam(r, "user"),
			Inactive: false,
		}

		resp, err := mgr.CreateUserIfMissing(r.Context(), params)
		ObjectToJsonBytes(r.Context(), rw, resp, err)

	})

	r.Mount("/user/{user}", api)
}

func mountActivityApi(r *chi.Mux, mgr *server.TimerecServer) {
	api := chi.NewRouter()
	api.Use(middleware.Logger)
	api.Use(middleware.AllowContentType("application/json"))
	api.Use(middleware.SetHeader("Content-Type", "application/json"))

	api.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "user")

		resp, err := mgr.GetActivity(r.Context(), server.GetUserParams{UserName: name})
		ObjectToJsonBytes(r.Context(), rw, resp, err)
	})
	api.Post("/", func(rw http.ResponseWriter, r *http.Request) {
		params := server.StartActivityParams{}
		err := json.NewDecoder(r.Body).Decode(&params)
		params.UserName = chi.URLParam(r, "user")
		if err != nil {
			http.Error(rw, http.StatusText(400), 400)
			return
		}

		resp, err := mgr.StartActivity(r.Context(), params)
		ObjectToJsonBytes(r.Context(), rw, resp, err)
	})
	api.Patch("/", func(rw http.ResponseWriter, r *http.Request) {
		params := server.ExtendActivityParams{}
		err := json.NewDecoder(r.Body).Decode(&params)
		params.UserName = chi.URLParam(r, "user")
		if err != nil {
			http.Error(rw, http.StatusText(400), 400)
			return
		}

		resp, err := mgr.ExtendActivity(r.Context(), params)
		ObjectToJsonBytes(r.Context(), rw, resp, err)
	})
	api.Delete("/", func(rw http.ResponseWriter, r *http.Request) {
		params := server.FinishActivityParams{}
		err := json.NewDecoder(r.Body).Decode(&params)
		params.UserName = chi.URLParam(r, "user")
		if err != nil {
			http.Error(rw, http.StatusText(400), 400)
			return
		}

		resp, err := mgr.FinishActivity(r.Context(), params)
		ObjectToJsonBytes(r.Context(), rw, resp, err)
	})

	r.Mount("/user/{user}/activity", api)
}

func mountJobApi(r *chi.Mux, mgr *server.TimerecServer) {
	api := chi.NewRouter()
	api.Use(middleware.Logger)
	api.Use(middleware.AllowContentType("application/json"))
	api.Use(middleware.SetHeader("Content-Type", "application/json"))

	api.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		default_after, _ := time.ParseDuration("-24h")
		default_before, _ := time.ParseDuration("0m")
		params := server.SearchJobParams{
			Name:          r.URL.Query().Get("name"),
			Owner:         chi.URLParam(r, "user"),
			StartedAfter:  default_after,
			StartedBefore: default_before,
		}

		resp, err := mgr.GetJob(r.Context(), params)
		if !resp.Success {
			rw.WriteHeader(404)
		}
		ObjectToJsonBytes(r.Context(), rw, resp, err)
	})
	api.Post("/{name}", func(rw http.ResponseWriter, r *http.Request) {
		params := server.SearchJobParams{
			Name:  chi.URLParam(r, "name"),
			Owner: chi.URLParam(r, "user")}

		resp, err := mgr.CreateJobIfMissing(r.Context(), params)
		ObjectToJsonBytes(r.Context(), rw, resp, err)
	})
	api.Put("/{name}", func(rw http.ResponseWriter, r *http.Request) {
		params := server.UpdateJobParams{}
		err := json.NewDecoder(r.Body).Decode(&params)
		params.Name = chi.URLParam(r, "name")
		params.Owner = chi.URLParam(r, "user")
		if err != nil {
			http.Error(rw, http.StatusText(400), 400)
			return
		}

		resp, err := mgr.UpdateJob(r.Context(), params)
		ObjectToJsonBytes(r.Context(), rw, resp, err)
	})
	api.Delete("/{name}", func(rw http.ResponseWriter, r *http.Request) {
		params := server.CompleteJobParams{}
		err := json.NewDecoder(r.Body).Decode(&params)
		params.Name = chi.URLParam(r, "name")
		params.Owner = chi.URLParam(r, "user")
		if err != nil {
			http.Error(rw, http.StatusText(400), 400)
			return
		}

		resp, err := mgr.CompleteJob(r.Context(), params)
		ObjectToJsonBytes(r.Context(), rw, resp, err)
	})

	r.Mount("/user/{user}/jobs", api)
}

func ObjectToJsonBytes(ctx context.Context, rw http.ResponseWriter, obj interface{}, err error) {
	reqid := ctx.Value(middleware.RequestIDKey).(string)
	rw.Header().Add(middleware.RequestIDHeader, reqid)

	// return an Error if
	if err != nil {
		var respErr server.ResponseError
		isResponseError := errors.As(err, &respErr)

		errbytes, jsonerr := json.Marshal(respErr)
		if jsonerr == nil && isResponseError {
			http.Error(rw, string(errbytes), 500)
			return
		} else {
			http.Error(rw, "{ \"error\": \"Encoding error\" }", 500)
			return
		}
	}

	// Encode Response
	bytes, err := json.Marshal(obj)
	if err != nil {
		http.Error(rw, "{ \"error\": \"Encoding error\" }", 500)
		return
	}
	rw.Write(bytes)
}

func mountTextApi(r *chi.Mux, mgr *server.TimerecServer) {
	txtapi := chi.NewRouter()
	txtapi.Use(middleware.Logger)
	txtapi.Use(middleware.AllowContentType("text/plain"))
	txtapi.Use(middleware.SetHeader("Content-Type", "text/plain"))

	txtapi.Get("/text/userStatus", func(rw http.ResponseWriter, r *http.Request) {
		state, err1 := mgr.StateProvider.Refresh(r.URL.Query().Get("user"))
		if err1 != nil {
			rw.WriteHeader(500)
			return
		}
		user, err2 := providers.GetUser(&state, api.User{Name: r.URL.Query().Get("user")})
		jobs, err3 := providers.ListJobs(&state)
		if err2 == providers.ProviderNotFound {
			rw.WriteHeader(404)
			return
		}
		if err3 != providers.ProviderOk {
			rw.WriteHeader(500)
			return
		}

		text := client.FormatUserStatus(user, jobs)
		rw.Write([]byte(text))
		rw.WriteHeader(200)
	})

	txtapi.Get("/text/day", func(rw http.ResponseWriter, r *http.Request) {
		state, err1 := mgr.StateProvider.Refresh(r.URL.Query().Get("user"))
		if err1 != nil {
			rw.WriteHeader(500)
			return
		}
		user, err2 := providers.GetUser(&state, api.User{Name: r.URL.Query().Get("user")})
		jobs, err3 := providers.ListJobs(&state)
		if err2 != providers.ProviderOk {
			rw.WriteHeader(500)
			return
		}
		if err3 != providers.ProviderOk {
			rw.WriteHeader(500)
			return
		}

		text := client.FormatUserStatus(user, jobs)
		rw.Write([]byte(text))
		rw.WriteHeader(200)
	})

	txtapi.Get("/text/week", func(rw http.ResponseWriter, r *http.Request) {

		// | Day       | Time    | Tasks (max 200 chars)                                              |
		// |-----------|---------|--------------------------------------------------------------------|
		// | Sunday    |  0h  0m | Fix Jira 123, Migrate databases to Operator, Deploy Argo Workflows |
		// | Monday    |  6h  0m | Fix Jira 123, Migrate databases to Operator, Deploy Argo Workflows |
		// | Tuesday   |  5h 43m | Fix Jira 123, Migrate databases to Operator, Deploy Argo Workflows |

		text := "TODO"
		rw.Write([]byte(text))
		rw.WriteHeader(200)
	})
	r.Mount("/", txtapi)
}

func mountUtils(r *chi.Mux) {
	// Health & Readiness Probe
	r.Get("/healthz", healthCheck)
	r.Get("/readyz", readinessCheck)
	r.Method(http.MethodGet, "/metrics", promhttp.Handler())

	// OpenAPI docs
	r.Get("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write(openapiContent)
		w.Header().Add("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
	})
	r.Mount("/docs", v3cdn.NewHandler("Timerec Service",
		"/docs/openapi.yaml", "/docs"))

	r.Mount("/debug", middleware.Profiler())
}
