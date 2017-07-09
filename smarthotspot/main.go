package main

import (
	"net/http"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/gurupras/go-stoppable-net-listener"
	"github.com/homesound/smarthotspot"
	"github.com/homesound/wifimanager"
	log "github.com/sirupsen/logrus"
)

var (
	app         = kingpin.New("smarthotspot", "Auto-host hotspot if no network found")
	iface       = app.Arg("iface", "Interface to use").String()
	wpaConfPath = app.Flag("wpa-conf", "Path to wpa_supplicant configuration file").Short('w').Default("/etc/wpa_supplicant/wpa_supplicant.conf").String()
	serverPath  = app.Flag("server-path", "Path to webserver files").Short('s').Default("/www/smarthotspot").String()
	port        = app.Flag("port", "Port to start webserver on").Short('p').Default("80").Int()
	verbose     = app.Flag("verbose", "Enable verbose messages").Short('v').Default("false").Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	wifiManager, err := wifimanager.New(*wpaConfPath)
	if err != nil {
		log.Errorf("Failed with error: %v", err)
		os.Exit(-1)
	}
	smartHotspot := smarthotspot.New(wifiManager, *iface)
	// Register the listeners
	hostapdChan := make(chan interface{})
	wpaSupplicantChan := make(chan interface{})
	smartHotspot.RegisterHostapdListener(hostapdChan)
	smartHotspot.RegisterWPASupplicantListener(wpaSupplicantChan)

	// Run the listeners
	go func() {
		for data := range hostapdChan {
			message := data.(string)
			switch message {
			case "started":
				// hostapd started
				// Start the webserver
				log.Infof("Starting webserver")
				StartWebServer(wifiManager, *port)
			case "stopped":
				// hostapd stopped
				// Stop the webserver
				log.Infof("Stopping webserver")
				StopWebServer()
			default:
				log.Errorf("Unknown hostapd message: %v", message)
			}
		}
	}()

	go func() {
		for data := range wpaSupplicantChan {
			message := data.(string)
			switch message {
			case "started":
			case "stopped":
			default:
				log.Errorf("Unknown WPASupplicant message: %v", message)
			}
		}
	}()

	if err := smartHotspot.Start(); err != nil {
		log.Errorf("Error in smart-hotspot: %v", err)
		os.Exit(-1)
	}
}

var snl *stoppablenetlistener.StoppableNetListener

func StartWebServer(wifiManager *wifimanager.WifiManager, port int) {
	// Set up the server
	handler := smarthotspot.SetupRoutes(*serverPath, wifiManager, nil)
	http.Handle("/", handler)
	server := http.Server{}
	snl, err := stoppablenetlistener.New(port)
	if err != nil {
		log.Fatalf("Failed to listen to port 80: %v", err)
	}
	server.Serve(snl)
}

func StopWebServer() {
	if snl != nil {
		snl.Stop()
	}
}
