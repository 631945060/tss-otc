package command

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var tssNodeID string
var tssNodeAddr string

var tssNodeCmd = &cobra.Command{
	Use:   "tssNode",
	Short: "start a local test TSS participant endpoint",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if tssNodeID == "" || tssNodeAddr == "" {
			return fmt.Errorf("node-id and addr are required")
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"status":"ok","node_id":%q,"role":"test-tss-participant"}`, tssNodeID)
		})
		mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		server := &http.Server{Addr: tssNodeAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
		go func() {
			<-cmd.Context().Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(shutdownCtx)
		}()
		log.Printf("test TSS participant %s listening on %s", tssNodeID, tssNodeAddr)
		return server.ListenAndServe()
	},
}

func init() {
	tssNodeCmd.Flags().StringVar(&tssNodeID, "node-id", "", "participant id")
	tssNodeCmd.Flags().StringVar(&tssNodeAddr, "addr", "", "listen address")
	rootCmd.AddCommand(tssNodeCmd)
}
