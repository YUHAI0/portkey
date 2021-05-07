package session

import (
	"bytes"
	"fmt"
	"github.com/DataDog/zstd"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/pion/quic"
	"github.com/sdslabs/portkey/pkg/utils"
)

func WriteLoopN(streams []*quic.BidirectionalStream, blockNumber int, sendPath string, wg *sync.WaitGroup) {
	defer wg.Done()

	tempfile, err := ioutil.TempFile(os.TempDir(), "portkey*")
	if err != nil {
		log.WithError(err).Errorf("Error in tempfile creation\n")
		return
	}
	defer os.Remove(tempfile.Name())

	err = utils.Tar(sendPath, tempfile)
	if err != nil {
		log.WithError(err).Errorf("Error in making tarball\n")
		return
	}

	if _, err = tempfile.Seek(0, 0); err != nil {
		log.WithError(err).Errorf("Error in going to tempfile start\n")
		return
	}

	if len(streams) != blockNumber {
		log.WithError(err).Errorf("streams's number is not the same as blockNumber")
	}

	//
	splitFiles, _ := utils.SplitFile(tempfile, blockNumber)

	fmt.Printf("splitfiles length: %d \n", len(splitFiles))

	//var stream *quic.BidirectionalStream
	var _wg sync.WaitGroup
	_wg.Add(len(streams))
	for i, stream := range streams {
		i := i
		stream := stream
		go func() {
			log.Info("================")
			log.Infof("Stream %d Writing %s...\n", i, splitFiles[i])
			quicgoStream := stream.Detach()

		 	//qw := bufio.NewWriter(quicgoStream)
			//_, _ = qw.WriteString(fmt.Sprintf("%d;", i))
			//_ = qw.Flush()
			//
			//quicgoStream = stream.Detach()

			partialFile, err := os.Open(splitFiles[i])

			zstdWriter := zstd.NewWriter(quicgoStream)

			zstdWriter.Write(utils.IntToBytes(i))
			bytesWritten, err := io.Copy(zstdWriter, partialFile)
			if err != nil {
				log.WithError(err).Errorf("Error in copying from tempfile to zstdWriter in sender stream %d\n", quicgoStream.StreamID())
			}
			log.Infof("Copied %d bytes from tempfile to zstdWriter in sender stream %d\n", bytesWritten, quicgoStream.StreamID())

			if err = zstdWriter.Close(); err != nil {
				log.WithError(err).Errorf("Error in closing zstdWriter in sender stream %d\n", quicgoStream.StreamID())
			}
			if err = quicgoStream.Close(); err != nil {
				log.WithError(err).Errorf("Error in closing sender stream %d", quicgoStream.StreamID())
			}
			log.Infof("Finished writing to sender stream %d", quicgoStream.StreamID())

			log.Infoln("Waiting for peer to close stream...")

			var dummyBuffer bytes.Buffer
			_, _ = io.Copy(&dummyBuffer, quicgoStream)
			_wg.Done()
		}()
	}
	log.Info("Waiting multi writing...")
	_wg.Wait()
	log.Infoln("DONE")
}

func WriteLoop(stream *quic.BidirectionalStream, sendPath string, wg *sync.WaitGroup) {
	defer wg.Done()

	quicgoStream := stream.Detach()

	tempfile, err := ioutil.TempFile(os.TempDir(), "portkey*")
	if err != nil {
		log.WithError(err).Errorf("Error in tempfile creation in sender stream %d\n", quicgoStream.StreamID())
		return
	}
	defer os.Remove(tempfile.Name())

	err = utils.Tar(sendPath, tempfile)
	if err != nil {
		log.WithError(err).Errorf("Error in making tarball in sender stream %d\n", quicgoStream.StreamID())
		return
	}

	if _, err = tempfile.Seek(0, 0); err != nil {
		log.WithError(err).Errorf("Error in going to tempfile start in sender stream %d\n", quicgoStream.StreamID())
		return
	}

	zstdWriter := zstd.NewWriter(quicgoStream)
	bytesWritten, err := io.Copy(zstdWriter, tempfile)
	if err != nil {
		log.WithError(err).Errorf("Error in copying from tempfile to zstdWriter in sender stream %d\n", quicgoStream.StreamID())
	}
	log.Infof("Copied %d bytes from tempfile to zstdWriter in sender stream %d\n", bytesWritten, quicgoStream.StreamID())

	if err = zstdWriter.Close(); err != nil {
		log.WithError(err).Errorf("Error in closing zstdWriter in sender stream %d\n", quicgoStream.StreamID())
	}
	if err = quicgoStream.Close(); err != nil {
		log.WithError(err).Errorf("Error in closing sender stream %d", quicgoStream.StreamID())
	}
	log.Infof("Finished writing to sender stream %d", quicgoStream.StreamID())

	log.Infoln("Waiting for peer to close stream...")

	var dummyBuffer bytes.Buffer
	_, _ = io.Copy(&dummyBuffer, quicgoStream)

}
