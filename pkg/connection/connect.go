package connection

import (
	"fmt"
	"sync"

	"github.com/pion/quic"
	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"

	"github.com/sdslabs/portkey/pkg/signal"
	"github.com/sdslabs/portkey/pkg/transfer"
)

var wg sync.WaitGroup
var stunServers []string = []string{"stun:stun.l.google.com:19302"}

func Connect(key string, sendPath string, receive bool, receivePath string) {
	isOffer := (key == "")
	api := webrtc.NewAPI()
	iceOptions := webrtc.ICEGatherOptions{
		ICEServers: []webrtc.ICEServer{
			{URLs: stunServers},
		},
	}
	gatherer, err := api.NewICEGatherer(iceOptions)
	if err != nil {
		log.Fatal(err)
	}

	ice := api.NewICETransport(gatherer)

	qt, err := api.NewQUICTransport(ice, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = gatherer.Gather()
	if err != nil {
		log.Fatal(err)
	}

	iceCandidates, err := gatherer.GetLocalCandidates()
	if err != nil {
		log.Fatal(err)
	}

	iceParams, err := gatherer.GetLocalParameters()
	if err != nil {
		log.Fatal(err)
	}

	quicParams, err := qt.GetLocalParameters()
	if err != nil {
		log.Fatal(err)
	}

	receiveErr := make(chan error)
	if receive {
		wg.Add(1)
		qt.OnBidirectionalStream(func(stream *quic.BidirectionalStream) {
			fmt.Printf("New stream received: streamid = %d\n", stream.StreamID())
			go transfer.ReadLoop(stream, receivePath, receiveErr, &wg)
		})
	}

	s := signal.Signal{
		ICECandidates:  iceCandidates,
		ICEParameters:  iceParams,
		QuicParameters: quicParams,
	}

	remoteSignal := signal.Signal{}

	if isOffer {
		err = signal.SignalExchange(&s, &remoteSignal)
	} else {
		err = signal.SignalExchangeWithKey(&s, &remoteSignal, key)
	}
	if err != nil {
		log.WithError(err).Fatalln("Unable to exchange signal")
	}
	iceRole := webrtc.ICERoleControlled
	if isOffer {
		iceRole = webrtc.ICERoleControlling
	}

	err = ice.SetRemoteCandidates(remoteSignal.ICECandidates)
	if err != nil {
		log.Fatal(err)
	}

	err = ice.Start(nil, remoteSignal.ICEParameters, &iceRole)
	if err != nil {
		log.Fatal(err)
	}

	err = qt.Start(remoteSignal.QuicParameters)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n\n------------Connection established------------")

	sendErr := make(chan error)
	if sendPath != "" {
		stream, err := qt.CreateBidirectionalStream()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("New stream created: streamid = %d\n", stream.StreamID())
		wg.Add(1)
		go transfer.WriteLoop(stream, sendPath, sendErr, &wg)
	}

	wg.Wait()
}
