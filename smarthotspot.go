package smarthotspot

import (
	"time"

	"github.com/homesound/wifimanager"
	log "github.com/sirupsen/logrus"
)

type SmartHotspot struct {
	wifiManager            *wifimanager.WifiManager
	iface                  string
	scanResultListeners    []chan interface{}
	hostapdListeners       []chan interface{}
	wpaSupplicantListeners []chan interface{}
}

func New(wifiManager *wifimanager.WifiManager, iface string) *SmartHotspot {
	s := &SmartHotspot{}
	s.wifiManager = wifiManager
	s.iface = iface
	s.scanResultListeners = make([]chan interface{}, 0)
	s.hostapdListeners = make([]chan interface{}, 0)
	s.wpaSupplicantListeners = make([]chan interface{}, 0)
	return s
}

func (s *SmartHotspot) RegisterScanResultListener(c chan interface{}) {
	s.scanResultListeners = append(s.scanResultListeners, c)
}
func (s *SmartHotspot) RegisterHostapdListener(c chan interface{}) {
	s.hostapdListeners = append(s.hostapdListeners, c)
}
func (s *SmartHotspot) RegisterWPASupplicantListener(c chan interface{}) {
	s.wpaSupplicantListeners = append(s.wpaSupplicantListeners, c)
}

func informListeners(listeners []chan interface{}, data interface{}) {
	for _, listener := range listeners {
		listener <- data
	}
}

func (s *SmartHotspot) Start() error {
	// Test for wifi connection. If no wifi connection is available for
	// more than 10 seconds, then turn on the hotspot and wait until a
	// connection is available.
	var err error
	var ssids []string

	wm := s.wifiManager
	iface := s.iface

	noKnownSSIDTimestamp := time.Now()

	// Make sure the interface is up
	err = wm.IfUp(iface)
	if err != nil {
		log.Fatalf("Failed to bring wifi interface '%v' up: %v", iface, err)
	}

	log.Infof("Known SSIDs=%v", wm.KnownSSIDs.List())

	for {
		wm.Lock()
		log.Debugf("Scanning for known SSIDs...")
		if ssids, err = wm.ScanForKnownSSID(); err != nil {
			log.Errorf("Failed to scan for known SSIDs: %v", err)
		} else {
			// Inform all scan-result listeners
			informListeners(s.scanResultListeners, ssids)

			log.Debugf("Got known SSIDS: %v", ssids)
			now := time.Now()
			if len(ssids) > 0 {
				if wm.IsHostapdRunning() {
					// We found a known SSID and we're in hotspot mode.
					// Get out of hotspot and start wpa_supplicant
					log.Infoln("Found known SSIDs when hotspot is running. Disable hotspot and try to connect to SSID")
					if err = wm.StopHotspot(iface); err != nil {
						log.Errorf("Failed to stop hotspot: %v", err)
					} else {
						// Inform that hotspot has stopped
						informListeners(s.hostapdListeners, "stopped")
						log.Infof("Hotspot stopped")
					}
				}
				if !wm.IsWPASupplicantRunning() {
					// We need to start wpa_supplicant
					if err = wm.StartWPASupplicant(iface, wm.WPAConfPath); err != nil {
						log.Errorf("Failed to start WPA supplicant: %v", err)
					} else {
						// Inform that wpa supplicant has started
						informListeners(s.wpaSupplicantListeners, "started")
						log.Infof("WPA supplicant started")
					}
				}
				noKnownSSIDTimestamp = time.Now()
			}
			if len(ssids) == 0 && now.Sub(noKnownSSIDTimestamp) > 10*time.Second {
				if wm.IsWPASupplicantRunning() {
					if err = wm.StopWPASupplicant(iface); err != nil {
						log.Errorf("Failed to stop WPA supplicant: %v", err)
					} else {
						// Inform that wpa supplicant has stopped
						informListeners(s.wpaSupplicantListeners, "stopped")
					}
				}
				if !wm.IsHostapdRunning() {
					log.Infoln("Scanning timed out. Starting hotspot")
					if err = wm.StartHotspot(iface); err != nil {
						log.Errorf("Failed to start hotspot: %v", err)
					} else {
						// Inform that hostapd has started
						informListeners(s.hostapdListeners, "started")
					}
				}
			}
		}
		wm.Unlock()
		time.Sleep(3 * time.Second)
	}
}
