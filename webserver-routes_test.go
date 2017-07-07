package smarthotspot

import (
	"net/http"
	"testing"
	"time"

	"github.com/gurupras/go-stoppable-net-listener"
	"github.com/homesound/wifimanager"
	"github.com/parnurzeal/gorequest"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func success(r http.ResponseWriter, req *http.Request) {

}

func TestWebserver(t *testing.T) {
	require := require.New(t)

	wm, err := wifimanager.New("/etc/wpa_supplicant/wpa_supplicant.conf")
	require.Nil(err, "Failed to get handle to wifi manager")

	snl, err := stoppablenetlistener.New(31233)
	require.Nil(err, "Failed to bind to port: 31233")
	snl.Timeout = 100 * time.Millisecond

	handler := SetupRoutes("", wm, nil)
	http.Handle("/", handler)
	http.HandleFunc("/success", func(r http.ResponseWriter, req *http.Request) {
		snl.Stop()
	})
	server := http.Server{}
	go func() {
		for {
			resp, body, errs := gorequest.New().Get("http://127.0.0.1:31233/success").End()
			log.Debugf("Got: %v\n%v\n%v\n\n", resp, body, errs)

			time.Sleep(100 * time.Millisecond)
		}
	}()
	server.Serve(snl)
}
