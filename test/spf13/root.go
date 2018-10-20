// Copyright © 2018 Cybozu
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"fmt"
	"os"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "spf13",
	Short: "test combination with spf13 tools",
	Long: `This is a test program that combines cybozu-go/well with
spf13/{cobra,pflag,viper} tools.`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := well.LogConfig{}.Apply()
		if err != nil {
			log.ErrorExit(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("CI") != "true" {
			log.Info("log: info", nil)
			log.Error("log: error", nil)
			log.Critical("log: critical", nil)
			return
		}

		logger := log.DefaultLogger()
		if logger.Enabled(log.LvError) {
			fmt.Println("error logs should be disabled")
			os.Exit(2)
		}
		if !logger.Enabled(log.LvCritical) {
			fmt.Println("critical logs should be eabled")
			os.Exit(2)
		}
		if _, ok := logger.Formatter().(log.JSONFormat); !ok {
			fmt.Println("log format should be JSON")
			os.Exit(2)
		}

		fmt.Println("test passed")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.spf13.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".spf13" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".spf13")
	}

	// If a config file is found, read it in.
	viper.ReadInConfig()
}
