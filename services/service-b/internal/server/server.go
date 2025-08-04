package server

import (
	"context"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/rs/zerolog"

	apb "github.com/mrdan4es/bazel-argocd-grpc-example/services/service-a/api/v1"
)

//go:embed index.html.tmpl
var indexTemplateBytes []byte

var indexTmpl = template.Must(template.New("index").Parse(string(indexTemplateBytes)))

func New(ctx context.Context, aClient apb.ServiceAClient) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", infoHandler(ctx, aClient))

	return mux
}

func infoHandler(ctx context.Context, aClient apb.ServiceAClient) func(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(ctx)

	return func(w http.ResponseWriter, r *http.Request) {
		info, err := aClient.GetSystemInfo(ctx, &apb.GetSystemInfoRequest{})
		if err != nil {
			log.Err(err).Msg("get system info from service-a")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := indexTmpl.Execute(w, info); err != nil {
			log.Err(err).Msg("execute template")
		}
	}
}
