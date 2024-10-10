/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jeessy2/ddns-go/v6/config"
	"github.com/jeessy2/ddns-go/v6/dns"
	"github.com/jeessy2/ddns-go/v6/util"
	"github.com/jeessy2/ddns-go/v6/util/update"
	"github.com/jeessy2/ddns-go/v6/web"
	"github.com/spf13/cobra"
)

var (
	versionFlag    bool
	updateFlag     bool
	listen         string
	every          int
	ipCacheTimes   int
	configFilePath string
	noWebService   bool
	skipVerify     bool
	customDNS      string
	newPassword    string
)

//go:embed static
var staticEmbeddedFiles embed.FS

//go:embed favicon.ico
var faviconEmbeddedFile embed.FS

var (
	version = "DEV"
)

var rootCmd = &cobra.Command{
	Use:    "ddns-go",
	Short:  "Simple and easy to use DDNS.",
	PreRun: preRun,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "ddns-go version")
	rootCmd.Flags().BoolVarP(&updateFlag, "update", "u", false, "Upgrade ddns-go to the latest version")
	rootCmd.Flags().StringVarP(&listen, "listen", "l", ":9876", "Listen address")
	rootCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "Config file path")
	rootCmd.Flags().BoolVar(&noWebService, "noweb", false, "Disable web service")
	rootCmd.Flags().IntVarP(&every, "frequency", "f", 300, "Update frequency(seconds)")
	rootCmd.Flags().IntVar(&ipCacheTimes, "cacheTimes", 5, "Cache times")
	rootCmd.Flags().BoolVar(&skipVerify, "skipVerify", false, "Skip certificate verification")
	rootCmd.Flags().StringVar(&customDNS, "dns", "", "Custom DNS server address, example: 8.8.8.8")
	rootCmd.Flags().StringVar(&newPassword, "resetPassword", "", "Reset password to the one entered")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		println(err)
		os.Exit(1)
	}
}

func preRun(cmd *cobra.Command, args []string) {
	if versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}
	if updateFlag {
		update.Self(version)
		os.Exit(0)
	}
}

func run() {
	// check listen address
	if _, err := net.ResolveTCPAddr("tcp", listen); err != nil {
		log.Fatalf("Parse listen address failed! Exception: %s", err)
	}
	// set version
	os.Setenv(web.VersionEnv, version)
	if configFilePath != "" {
		absPath, _ := filepath.Abs(configFilePath)
		os.Setenv(util.ConfigFilePathENV, absPath)
	}
	if newPassword != "" {
		conf, _ := config.GetConfigCached()
		conf.ResetPassword(newPassword)
		return
	}
	if skipVerify {
		util.SetInsecureSkipVerify()
	}
	if customDNS != "" {
		util.SetDNS(customDNS)
	}
	os.Setenv(util.IPCacheTimesENV, strconv.Itoa(ipCacheTimes))

	start()
}

func start() {
	conf, _ := config.GetConfigCached()
	conf.CompatibleConfig()
	// initialize language
	util.InitLogLang(conf.Lang)

	if !noWebService {
		go func() {
			// start web service
			err := runWebServer()
			if err != nil {
				log.Println(err)
				time.Sleep(time.Minute)
				os.Exit(1)
			}
		}()
	}
	util.InitBackupDNS(customDNS, conf.Lang)

	util.WaitInternet(dns.Addresses)

	dns.RunTimer(time.Duration(every) * time.Second)
}

func runWebServer() error {
	// start static file server
	http.HandleFunc("/static/", web.AuthAssert(staticFsFunc))
	http.HandleFunc("/favicon.ico", web.AuthAssert(faviconFsFunc))
	http.HandleFunc("/login", web.AuthAssert(web.Login))
	http.HandleFunc("/loginFunc", web.AuthAssert(web.LoginFunc))

	http.HandleFunc("/", web.Auth(web.Writing))
	http.HandleFunc("/save", web.Auth(web.Save))
	http.HandleFunc("/logs", web.Auth(web.Logs))
	http.HandleFunc("/clearLog", web.Auth(web.ClearLog))
	http.HandleFunc("/webhookTest", web.Auth(web.WebhookTest))

	util.Log("Listening on %s", listen)

	l, err := net.Listen("tcp", listen)
	if err != nil {
		return errors.New(util.LogStr("Failed to listen on the port, please check if the port is occupied! %s", err))
	}

	return http.Serve(l, nil)
}

func staticFsFunc(writer http.ResponseWriter, request *http.Request) {
	http.FileServer(http.FS(staticEmbeddedFiles)).ServeHTTP(writer, request)
}

func faviconFsFunc(writer http.ResponseWriter, request *http.Request) {
	http.FileServer(http.FS(faviconEmbeddedFile)).ServeHTTP(writer, request)
}