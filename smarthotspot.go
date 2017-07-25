package smarthotspot

import (
	"fmt"
	"time"

	"github.com/homesound/wifimanager"
	log "github.com/sirupsen/logrus"
)

type Command string

const (
	FORCE_HOSTAPD        Command = "force_hostapd"
	FORCE_WPA_SUPPLICANT Command = "force_wpa_supplicant"
)

type SmartHotspot struct {
	wifiManager            *wifimanager.WifiManager
	iface                  string
	scanResultListeners    []chan interface{}
	hostapdListeners       []chan interface{}
	wpaSupplicantListeners []chan interface{}
	CommandChannel         chan Command
}

func New(wifiManager *wifimanager.WifiManager, iface string) *SmartHotspot {
	s := &SmartHotspot{}
	s.wifiManager = wifiManager
	s.iface = iface
	s.scanResultListeners = make([]chan interface{}, 0)
	s.hostapdListeners = make([]chan interface{}, 0)
	s.wpaSupplicantListeners = make([]chan interface{}, 0)
	s.CommandChannel = make(chan Command, 0)
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
		select {
		case cmd := <-s.CommandChannel:
			switch cmd {
			case FORCE_HOSTAPD:
				s.EnableHostapd(false)
			case FORCE_WPA_SUPPLICANT:
				s.EnableWPASupplicant(false)
			default:
				log.Errorf("Unknown command received in command channel: %v", cmd)
			}
		default:
			log.Debugf("Scanning for known SSIDs...")
			if ssids, err = wm.ScanForKnownSSID(); err != nil {
				log.Errorf("Failed to scan for known SSIDs: %v", err)
			} else {
				// Inform all scan-result listeners
				informListeners(s.scanResultListeners, ssids)

				log.Debugf("Got known SSIDS: %v", ssids)
				now := time.Now()
				if len(ssids) > 0 {
					s.EnableHostapd(true)
					noKnownSSIDTimestamp = time.Now()
				}
				if len(ssids) == 0 && now.Sub(noKnownSSIDTimestamp) > 10*time.Second {
					s.EnableWPASupplicant(true)
				}
			}
			wm.Unlock()
			time.Sleep(3 * time.Second)
		}
	}
}

func (s *SmartHotspot) EnableHostapd(shouldInformListeners bool) (err error) {
	wm := s.wifiManager
	if wm.IsHostapdRunning() {
		// We found a known SSID and we're in hotspot mode.
		// Get out of hotspot and start wpa_supplicant
		log.Infoln("Found known SSIDs when hotspot is running. Disable hotspot and try to connect to SSID")
		if err = wm.StopHotspot(s.iface); err != nil {
			err = fmt.Errorf("Failed to stop hotspot: %v", err)
		} else {
			if shouldInformListeners {
				// Inform that hotspot has stopped
				informListeners(s.hostapdListeners, "stopped")
			}
			log.Infof("Hotspot stopped")
		}
	}
	if !wm.IsWPASupplicantRunning() {
		// We need to start wpa_supplicant
		if err = wm.StartWPASupplicant(s.iface, wm.WPAConfPath); err != nil {
			err = fmt.Errorf("Failed to start WPA supplicant: %v", err)
		} else {
			if shouldInformListeners {
				// Inform that wpa supplicant has started
				informListeners(s.wpaSupplicantListeners, "started")
			}
			log.Infof("WPA supplicant started")
		}
	}
	return
}

func (s *SmartHotspot) EnableWPASupplicant(shouldInformListeners bool) (err error) {
	wm := s.wifiManager
	if wm.IsWPASupplicantRunning() {
		if err = wm.StopWPASupplicant(s.iface); err != nil {
			err = fmt.Errorf("Failed to stop WPA supplicant: %v", err)
		} else {
			if shouldInformListeners {
				// Inform that wpa supplicant has stopped
				informListeners(s.wpaSupplicantListeners, "stopped")
			}
			log.Infof("WPA supplicant stopped")
		}
	}
	if !wm.IsHostapdRunning() {
		log.Infoln("Scanning timed out. Starting hotspot")
		if err = wm.StartHotspot(s.iface); err != nil {
			err = fmt.Errorf("Failed to start hotspot: %v", err)
		} else {
			if shouldInformListeners {
				// Inform that hostapd has started
				informListeners(s.hostapdListeners, "started")
			}
			log.Infof("Hostapd started")
		}
	}
	return
}
