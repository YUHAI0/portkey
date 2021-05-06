package signal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/pion/webrtc/v3"
	"github.com/sdslabs/portkey/pkg/utils"
)

type Signal struct {
	ICECandidates  []webrtc.ICECandidate `json:"iceCandidates"`
	ICEParameters  webrtc.ICEParameters  `json:"iceParameters"`
	QuicParameters webrtc.QUICParameters `json:"quicParameters"`
}


type SignalN struct {
	ICECandidates  []webrtc.ICECandidate `json:"iceCandidates"`
	ICEParameters  webrtc.ICEParameters  `json:"iceParameters"`
	QuicParameters webrtc.QUICParameters `json:"quicParameters"`
	BlockNumber int `json:"blockNumber"`
}

var serverURL string = "https://portkey-server.herokuapp.com/"

func SignalExchangeN(localSignal, remoteSignal *SignalN) error {
	connParams, err := utils.Encode(localSignal)
	if err != nil {
		return err
	}

	log.Infoln("Requesting a key...")

	resp, err := http.PostForm(serverURL, url.Values{
		"connParams": {connParams},
	})
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	key := string(body)
	fmt.Printf("Your Portkey: %s\n", key)

	log.Infoln("Waiting for peer...")
	resp, err = http.PostForm((serverURL + "wait"), url.Values{
		"key": {key},
	})
	if err != nil {
		return err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	fmt.Printf("remote in exchange:1 %s\n\n", remoteSignal)
	err = utils.Decode(string(body), remoteSignal)
	fmt.Printf("remote in exchange:2 %s\n\n", remoteSignal)
	return err
}


func SignalExchange(localSignal, remoteSignal *Signal) error {
	connParams, err := utils.Encode(localSignal)
	if err != nil {
		return err
	}

	log.Infoln("Requesting a key...")

	resp, err := http.PostForm(serverURL, url.Values{
		"connParams": {connParams},
	})
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	key := string(body)
	fmt.Printf("Your Portkey: %s\n", key)

	log.Infoln("Waiting for peer...")
	resp, err = http.PostForm((serverURL + "wait"), url.Values{
		"key": {key},
	})
	if err != nil {
		return err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	fmt.Printf("remote in exchange:1 %s\n\n", remoteSignal)
	err = utils.Decode(string(body), remoteSignal)
	fmt.Printf("remote in exchange:2 %s\n\n", remoteSignal)
	return err
}

func SignalExchangeWithKeyN(localSignal, remoteSignal *SignalN, key string) error {
	connParams, err := utils.Encode(localSignal)
	if err != nil {
		return err
	}

	log.Infoln("Sending key to signalling server...")

	fmt.Printf("local: %s\n\n",localSignal)

	fmt.Printf("remote:1 %s\n\n", remoteSignal)
	resp, err := http.PostForm((serverURL + "key"), url.Values{
		"key":        {key},
		"connParams": {connParams},
	})
	//fmt.Print(resp)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	err = utils.Decode(string(body), remoteSignal)
	fmt.Printf("remote:2 %s\n\n", remoteSignal)
	return err
}

func SignalExchangeWithKey(localSignal, remoteSignal *Signal, key string) error {
	connParams, err := utils.Encode(localSignal)
	if err != nil {
		return err
	}

	log.Infoln("Sending key to signalling server...")

	fmt.Printf("local: %s\n\n",localSignal)

	fmt.Printf("remote:1 %s\n\n", remoteSignal)
	resp, err := http.PostForm((serverURL + "key"), url.Values{
		"key":        {key},
		"connParams": {connParams},
	})
	//fmt.Print(resp)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	err = utils.Decode(string(body), remoteSignal)
	fmt.Printf("remote:2 %s\n\n", remoteSignal)
	return err
}
