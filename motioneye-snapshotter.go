package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	listenIp   string
	listenPort string
	meServer   string
	meSig      string
	meUser     string
)

func init() {
	configFile := flag.String("config", "", "Configuration file")
	configFilePath := flag.String("configpath", "", "Path to configuration file")
	flag.Bool("displayconfig", false, "Display configuration")
	flag.String("indexfile", "./index.html", "Default index file")
	flag.String("listenip", "127.0.0.1", "IP address to bind to")
	flag.String("listenport", "5757", "Port to bind to")
	flag.String("meuser", "", "MotionEye Username")
	flag.String("mesig", "", "MotionEye Snapshot Signiture")
	flag.String("meserver", "", "MotionEye Server URL")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if *configFilePath == "" {
		viper.AddConfigPath(".")
	} else {
		viper.AddConfigPath(*configFilePath)
	}

	if *configFile == "" {
		viper.SetConfigName("config")
	} else {
		viper.SetConfigName(*configFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Log file not found")
		} else {
			log.Fatalf("Config file was found but another error was produced")
		}
	}

	listenIp = viper.GetString("listenip")
	listenPort = viper.GetString("listenport")
	meServer = viper.GetString("meserver")
	meSig = viper.GetString("mesig")
	meUser = viper.GetString("meuser")

	log.Println("listenport=", listenPort)
	log.Println("listenip=", listenIp)
	log.Println("meserver=", meServer)
	log.Println("mesig=", meSig)
	log.Println("meuser=", meUser)
}

func main() {
	if viper.GetBool("displayconfig") {
		displayConfig()
		os.Exit(0)
	}

	startWeb(listenIp, listenPort)
}

func takeSnapshot(camera int) {
	currentTime := time.Now()
	fileUrl := meServer + "/picture/" + strconv.Itoa(camera) + "/current/?_username=" + meUser + "&_signature=" + meSig
	log.Println("fileUrl= ", fileUrl)
	fileName := (currentTime.Format("20060102_150405") + ".jpg")
	fileDir := "output/camera" + strconv.Itoa(camera) + "/"
	err := DownloadFile(fileDir+fileName, fileUrl)
	if err != nil {
		panic(err)
		fmt.Println("Could not download file: " + fileName)
	}
	log.Println("Downloaded: " + fileDir + fileName)
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func startWeb(listenip string, listenport string) {
	r := mux.NewRouter()

	r.HandleFunc("/", handlerIndex)

	hostsRouter := r.PathPrefix("/snap").Subrouter()
	hostsRouter.HandleFunc("", handlerSnap)

	log.Println("Starting HTTP Webserver: http://" + listenIp + ":" + listenPort)
	err := http.ListenAndServe(listenIp+":"+listenPort, r)

	if err != nil {
		log.Println("error starting webserver:")
		log.Println(err)
	}
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting handlerIndex")
	printFile(viper.GetString("IndexFile"), w)
}

func printFile(filename string, webprint http.ResponseWriter) {
	log.Println("Starting printFile")
	texttoprint, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("cannot open file: ", filename)
		http.Error(webprint, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	fmt.Fprintf(webprint, "%s", string(texttoprint))
}

func handlerSnap(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting handlerSnap")
	vars := mux.Vars(r)
	queries := r.URL.Query()
	log.Printf("vars = %q\n", vars)
	log.Printf("queries = %q\n", queries)

	var camera int = 0

	switch {
	case strings.ToLower(queries.Get("camera")) == "main-door":
		camera = 1
	case strings.ToLower(queries.Get("camera")) == "lobby":
		camera = 2
	default:
		fmt.Fprintf(w, "%s", "unknown camera")
		return
	}

	log.Println("handler snap camera " + string(camera))
	fmt.Fprintf(w, "%s", "camera "+strconv.Itoa(camera))
	takeSnapshot(camera)

}

func displayConfig() {
	allmysettings := viper.AllSettings()
	var keys []string
	for k := range allmysettings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println("CONFIG:", k, ":", allmysettings[k])
	}
}
