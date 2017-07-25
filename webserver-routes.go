package smarthotspot

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/homesound/go-networkmanager"
	"github.com/homesound/wifimanager"
	log "github.com/sirupsen/logrus"
)

var staticPath string

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join(staticPath, "html", "index.html")
	http.ServeFile(w, r, filePath)
}

func WifiHandler(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join(staticPath, "html", "wifi-configuration.html")
	http.ServeFile(w, r, filePath)
}

func SetupRoutes(path string, wifiManager *wifimanager.WifiManager, io *socketio.Server) http.Handler {
	if strings.Compare(path, "") == 0 {
		path = "."
	}
	staticPath = filepath.Join(path, "static")

	r := mux.NewRouter()
	r.HandleFunc("/", WifiHandler)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	var err error
	if io == nil {
		// Set up the socket.io server
		io, err = socketio.NewServer(nil)
		if err != nil {
			log.Fatalf("Failed to start socket.io server: %v\n", err)
		}
	}

	io.On("connection", func(so socketio.Socket) {
		if err != nil {
			log.Errorf("Failed to get wifi manager: %v", err)
			return
		}

		so.On("wifi-scan", func(msg string) {
			log.Infoln("/wifi-scan")
			ifaces, err := wifiManager.GetWifiInterfaces()
			if err != nil {
				log.Errorf("Failed to list wifi interfaces: %v\n", err)
				return
			}

			type ifaceResult struct {
				Interface   string                            `json:"interface"`
				ScanResults []*networkmanager.WifiScanResult `json:"scanResults"`
			}

			results := make([]*ifaceResult, 0)
			for _, iface := range ifaces {
				scanResults, err := wifiManager.WifiScan(iface)
				if err != nil {
					log.Errorf("Failed to perform wifi-scan on interface '%v': %v\n", iface, err)
					continue
				}
				res := &ifaceResult{iface, scanResults}
				results = append(results, res)
			}

			b, err := json.Marshal(results)
			if err != nil {
				log.Errorf("Failed to marshal scan results: %v\n", err)
			}
			so.Emit("wifi-scan-results", string(b))
			//fmt.Printf("Sending back results:\n%v\n", string(b))
		})

		so.On("wifi-connect", func(s string) {
			log.Infoln("/wifi-connect")
			type wifiCred struct {
				SSID     string `json:"SSID"`
				Password string `json:"password"`
			}

			var cred wifiCred
			err := json.Unmarshal([]byte(s), &cred)
			if err != nil {
				log.Errorf("Failed to unmarshal data into wifiCred: %v\n%v\n", err, s)
				return
			}
			wifiInterfaces, err := wifiManager.GetWifiInterfaces()
			if err != nil {
				log.Errorf("Failed to get wifi interfaces: %v", err)
				return
			}
			log.Infof("Testing wifi connection with SSID=%v psk=%v", cred.SSID, cred.Password)
			for _, iface := range wifiInterfaces {
				networkStr, err := wifimanager.WPAPassphrase(cred.SSID, cred.Password)
				if err != nil {
					log.Errorf("Failed to call wpa_passphrase: %v", err)
					return
				}
				network := wifimanager.ParseWPANetwork(networkStr)
				if network == nil {
					log.Errorf("Failed to parse WPA network from:\n%v\n", networkStr)
					return
				}
				err = wifiManager.TestConnect(iface, network)
				if err != nil {
					log.Errorf("Failed to connect: %v\n", err)
					return
				} else {
					log.Infoln("Successfully tested connection. Adding it to WPA supplicant conf file")
				}
				// Connection succeeded
				// Update wpa supplicant file
				wifiManager.Lock()
				if err = wifiManager.AddNetworkConf(cred.SSID, cred.Password); err != nil {
					log.Errorf("Failed to add network to WPA supplicant conf: %v\n", err)
				} else {
					log.Infoln("Added network info to WPA conf file")
				}

				wifiManager.UpdateKnownSSIDs()
				wifiManager.Unlock()
			}
		})
	})

	http.Handle("/socket.io/", io)
	return r
}
