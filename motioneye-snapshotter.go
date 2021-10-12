package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	listenIp   string
	listenPort string
	meServer   string
	meSig      string
	meUser     string
	outputDir  string
)

const applicationVersion string = "v0.4"

// header file for webpages
const webpageheader string = `<!DOCTYPE HTML>
<html>
<body>
`

// footer file for webpages
const webpagefooter string = `</body>
</html>
`

func init() {
	flag.String("config", "config.yaml", "Configuration file: /path/to/file.yaml, default = ./config.yaml")
	flag.Bool("displayconfig", false, "Display configuration")
	flag.Bool("help", false, "Display help information")
	flag.String("indexfile", "./index.html", "Default index file")
	flag.String("listenip", "0.0.0.0", "IP address to bind to (0.0.0.0 = all IPs)")
	flag.String("listenport", "5757", "Port to bind to")
	flag.String("meuser", "", "MotionEye Username")
	flag.String("mesig", "", "MotionEye Snapshot Signiture")
	flag.String("meserver", "", "MotionEye Server URL")
	flag.String("outputdir", "./output", "Output Directory")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if viper.GetBool("help") {
		displayHelp()
		os.Exit(0)
	}

	configdir, configfile := filepath.Split(viper.GetString("config"))

	// set default configuration directory to current directory
	if configdir == "" {
		configdir = "."
	}

	viper.SetConfigType("yaml")
	viper.AddConfigPath(configdir)

	config := strings.TrimSuffix(configfile, ".yaml")
	config = strings.TrimSuffix(config, ".yml")

	viper.SetConfigName(config)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Config file not found")
		} else {
			log.Fatal("Config file was found but another error was discovered")
		}
	}

	listenIp = viper.GetString("listenip")
	listenPort = viper.GetString("listenport")
	meServer = viper.GetString("meserver")
	meSig = viper.GetString("mesig")
	meUser = viper.GetString("meuser")
	outputDir = viper.GetString("outputdir")

	log.Println("listenport=", listenPort)
	log.Println("listenip=", listenIp)
	log.Println("meserver=", meServer)
	log.Println("mesig=", meSig)
	log.Println("meuser=", meUser)
	log.Println("outputdir=", outputDir)
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
	fileDir := outputDir + "/camera" + strconv.Itoa(camera) + "/"
	err := downloadFile(fileDir+fileName, fileUrl)
	if err != nil {
		log.Println("ERROR: Could not download file: " + fileName)
		log.Println(err)
	}
	log.Println("Downloaded: " + fileDir + fileName)
}

func downloadFile(filepath string, url string) error {
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

	camerasRouter := r.PathPrefix("/cameras").Subrouter()
	camerasRouter.HandleFunc("", handlerCameras)

	filesRouter := r.PathPrefix("/files").Subrouter()
	filesRouter.HandleFunc("", handlerCameraFiles)
	filesRouter.HandleFunc("/{camera}", handlerShowImage).Queries("file", "")
	filesRouter.HandleFunc("/{camera}", handlerFiles)

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

func handlerCameras(w http.ResponseWriter, r *http.Request) {
	groups := viper.GetStringMap("cameras")
	for k, v := range groups {
		fmt.Fprintf(w, "%s: %s\n", k, v)
	}
}

// https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func byteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func handlerCameraFiles(w http.ResponseWriter, r *http.Request) {
	groups := viper.GetStringMap("cameras")

	fmt.Fprintf(w, webpageheader)
	fmt.Fprintf(w, "<table>\n")

	keys := make([]string, 0, len(groups))

	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(w, "  <tr><td>%s</td><td><a href=\"./files/%s\">%s</a></td></tr>\n", k, groups[k], groups[k])
	}

	fmt.Fprintf(w, "</table>\n")
	fmt.Fprintf(w, webpagefooter)
}

func handlerFiles(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting handlerFiles")

	fullcameras := viper.GetStringMap("cameras")

	namecameras := make([]string, 0, len(fullcameras))

	for _, v := range fullcameras {
		namecameras = append(namecameras, strings.ToLower(v.(string)))
	}
	sort.Strings(namecameras)

	vars := mux.Vars(r)

	if len(vars["camera"]) > 0 {
		// check if valid camera
		if !contains(namecameras, strings.ToLower(vars["camera"])) {
			fmt.Fprintf(w, "camera not found")
			return
		}
	}

	camerakey, ok := mapkey(fullcameras, strings.ToLower(vars["camera"]))

	if !ok {
		fmt.Fprintf(w, "camera key for %s not found\n", strings.ToLower(vars["camera"]))
		return
	}

	files, err := ioutil.ReadDir(outputDir + "/camera" + camerakey)
	if err != nil {
		fmt.Fprintf(w, "Camera output directory doesnt exist")
		log.Println(err)
		return
	}

	fmt.Fprintf(w, webpageheader)
	fmt.Fprintf(w, "<table>\n")

	for _, file := range files {
		if file.IsDir() {
			fmt.Fprintf(w, "  <tr><td>Directory: %s</td><td></td></tr>\n", file.Name())
		} else {
			fmt.Fprintf(w, "  <tr><td><a href=\"./%s?file=%s\">%s</a></td><td>%s</td></tr>\n", strings.ToLower(vars["camera"]), file.Name(), file.Name(), byteCountSI(file.Size()))
		}
	}

	fmt.Fprintf(w, "</table>\n")
	fmt.Fprintf(w, webpagefooter)
}

// does a string slice contain a value
// https://freshman.tech/snippets/go/check-if-slice-contains-element/
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// looks up a key in a map when given a value
// https://stackoverflow.com/questions/33701828/simple-way-of-getting-key-depending-on-value-from-hashmap-in-golang
func mapkey(m map[string]interface{}, value string) (key string, ok bool) {
	for k, v := range m {
		if v.(string) == value {
			key = k
			ok = true
			return
		}
	}
	return
}

func handlerShowImage(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting handlerShowImage")

	fullcameras := viper.GetStringMap("cameras")

	namecameras := make([]string, 0, len(fullcameras))

	for _, v := range fullcameras {
		namecameras = append(namecameras, strings.ToLower(v.(string)))
	}
	sort.Strings(namecameras)

	vars := mux.Vars(r)
	queries := r.URL.Query()

	if queries.Get("file") == "" {
		fmt.Fprintf(w, "no image specified")
		return
	}

	if len(vars["camera"]) > 0 {
		// check if valid camera
		if !contains(namecameras, strings.ToLower(vars["camera"])) {
			fmt.Fprintf(w, "camera not found")
			return
		}
	}

	camerakey, ok := mapkey(fullcameras, strings.ToLower(vars["camera"]))

	if !ok {
		fmt.Fprintf(w, "camera key for %s not found\n", strings.ToLower(vars["camera"]))
		return
	}

	// clean file name to prevent path traversal
	cleanFileName := path.Join("/", queries.Get("file"))

	w.Header().Set("Content-Type", "image/jpeg")
	printFile(outputDir+"/camera"+camerakey+cleanFileName, w)
}

func displayHelp() {
	message := `      --config string       Configuration file
      --config string       Configuration file: /path/to/file.yaml (default "./config.yaml")
      --displayconfig       Display configuration
      --help                Display help information
      --indexfile string    Default index file (default "./index.html")
      --listenip string     IP address to bind to (0.0.0.0 = all IPs) (default "0.0.0.0")
      --listenport string   Port to bind to (default "5757")
      --meserver string     MotionEye Server URL
      --mesig string        MotionEye Snapshot Signiture
      --meuser string       MotionEye Username
      --outputdir string    Output Directory (default "./output")
`
	fmt.Println("motioneye-snapshotter " + applicationVersion)
	fmt.Println(message)
}
