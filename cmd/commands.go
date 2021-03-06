package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/sdslabs/portkey/pkg/connection"
)

var (
	key            string
	sendPath       string
	streamNumber   int32
	receive        bool
	receivePath    string
	doBenchmarking bool
)

// rootCmd represents the run command
var rootCmd = &cobra.Command{
	Use:   "portkey",
	Short: "Portkey is a p2p file transfer tool.",
	Long:  `Portkey is a p2p file transfer tool that uses ORTC p2p API over QUIC protocol to achieve very fast file transfer speeds`,
	Run: func(cmd *cobra.Command, args []string) {
		connection.Connect(key, streamNumber, sendPath, receive, receivePath, doBenchmarking)
	},
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	rootCmd.Flags().StringVarP(&key, "key", "k", "", "Key to connect to peer")
	rootCmd.Flags().StringVarP(&sendPath, "send", "s", "", "Absolute path of directory/file to send")
	rootCmd.Flags().Int32VarP(&streamNumber, "number", "n", 1, "Set stream number")
	rootCmd.Flags().BoolVarP(&receive, "receive", "r", false, "Set to receive files")
	rootCmd.Flags().StringVarP(&receivePath, "rpath", "p", "", "Absolute path of where to receive files, pwd by default")
	rootCmd.Flags().BoolVarP(&doBenchmarking, "benchmark", "b", false, "Set to benchmark locally(for local testing)")
}
