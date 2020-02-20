package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func ServeProfiler(logger *logrus.Logger) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	profilerAddr := fmt.Sprintf("%s:%s", viper.GetString("profiler.address"), viper.GetString("profiler.port"))
	logger.Infof("Server profiler on %s", profilerAddr)

	return http.ListenAndServe(profilerAddr, mux)
}
