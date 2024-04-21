/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package util

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	util "github.com/richinsley/comfycli/pkg"
	"github.com/spf13/cobra"
)

// fileserverCmd represents the info command
var fileserverCmd = &cobra.Command{
	Use:   "fileserve <path>",
	Short: "Serve files in a folder over HTTP/HTTPS",
	Long:  `Serve files in a folder over HTTP/HTTPS`,
	Run: func(cmd *cobra.Command, args []string) {
		// the default path is the current directory
		path, err := os.Getwd()
		if err != nil {
			slog.Warn("error getting current directory", "error", err)
		}

		if len(args) > 0 {
			path = args[0]
		}

		if path == "" {
			slog.Warn("no path specified")
			os.Exit(1)
		}

		// get the path to serve files from
		port, _ := cmd.Flags().GetInt("port")
		storage, _ := cmd.Flags().GetString("storage")
		auth, _ := cmd.Flags().GetString("auth")
		cert, _ := cmd.Flags().GetString("cert")
		key, _ := cmd.Flags().GetString("key")
		selfsigned, _ := cmd.Flags().GetBool("selfsigned")

		options := util.FileServerOptions{
			Port:        port,
			RootPath:    path,
			StoragePath: storage,
			AuthToken:   auth,
			CertFile:    cert,
			KeyFile:     key,
			SelfSigned:  selfsigned,
		}

		// create the file server
		server, err := util.StartFileServer(options)
		if err != nil {
			slog.Error("error starting file server", "error", err)
			os.Exit(1)
		}

		// Create a channel to receive signals
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		// Block until a signal is received
		sig := <-c
		log.Printf("Received signal %s, stopping server...", sig)

		// Stop the server
		if err := util.StopFileServer(server); err != nil {
			log.Printf("Error stopping server: %s", err)
		}

		log.Println("Server stopped")
	},
}

func InitFileServe(systemCmd *cobra.Command) {
	systemCmd.AddCommand(fileserverCmd)

	// port to serve files on
	fileserverCmd.Flags().IntP("port", "", 8080, "Port to serve files on")

	// storage path for uploaded files
	fileserverCmd.Flags().StringP("storage", "", "", "Path to store uploaded files")

	// optional auth token
	fileserverCmd.Flags().StringP("auth", "", "", "Authorization token")

	// optional cert and key files for HTTPS
	fileserverCmd.Flags().StringP("cert", "", "", "Path to cert file")
	fileserverCmd.Flags().StringP("key", "", "", "Path to key file")

	// self-signed cert
	fileserverCmd.Flags().BoolP("selfsigned", "", false, "Generate a self-signed certificate")
}
