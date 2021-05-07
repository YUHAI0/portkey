package session

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/DataDog/zstd"
	log "github.com/sirupsen/logrus"

	"github.com/pion/quic"
	"github.com/sdslabs/portkey/pkg/utils"
)

func ReadLoopN(streams []*quic.BidirectionalStream, blockNumber int,receivePath string, wg *sync.WaitGroup) {
	defer wg.Done()

	if len(streams) != blockNumber {
		log.Errorf("streams size %d is not blockNumber: %d\n", len(streams), blockNumber)
		return
	}

	var _wg sync.WaitGroup
	_wg.Add(len(streams))
	for i, stream := range streams {
		stream := stream
		i := i
		go func() {
			quicgoStream := stream.Detach()
			log.Info("##############")
			log.Infof("Stream %d reading...\n", i)

			//tempfile, err := ioutil.TempFile(os.TempDir(), "portkey*")
			//if err != nil {
			//	log.WithError(err).Errorf("Error in tempfile creation in receiver stream %d\n", quicgoStream.StreamID())
			//	return
			//}
			//defer os.Remove(tempfile.Name())
			zstdReader := zstd.NewReader(quicgoStream)
			var intBytes []byte = make([]byte, 8)
			_, _ = zstdReader.Read(intBytes)
			var index = utils.BytesToInt(intBytes)
			log.Infof("Read index %d\n", utils.BytesToInt(intBytes))

			pName := fmt.Sprintf("partial.%d", index)
			partialFile, _ := os.OpenFile(pName, os.O_CREATE|os.O_WRONLY, 0600)
			bytesWritten, err := io.Copy(partialFile, zstdReader)

			log.Infof("%s index %s, file has %d bytes\n", partialFile.Name(), index, bytesWritten)
			if err != nil {
				log.WithError(err).Errorf("Error in copying from zstdReader to tempfile in receiver stream %d\n", quicgoStream.StreamID())
			}
			log.Infof("Copied %d bytes from zstdReader to tempfile in receiver stream %d", bytesWritten, quicgoStream.StreamID())

			if err = zstdReader.Close(); err != nil {
				log.WithError(err).Errorf("Error in closing zstdWriter in receiver stream %d\n", quicgoStream.StreamID())
			}
			if err = quicgoStream.Close(); err != nil {
				log.WithError(err).Errorf("Error in closing stream %d\n", quicgoStream.StreamID())
			}
			log.Infof("Finished reading from stream %d\n", quicgoStream.StreamID())
			_wg.Done()
		}()
	}

	_wg.Wait()
	log.Infoln("Reading DONE")
	// merge partial file and untar
	log.Infof("Merge partial files...")

	tempfile, err := ioutil.TempFile(os.TempDir(), "portkey*")
	if err != nil {
		log.WithError(err).Errorf("Error in tempfile creation in receiver stream\n")
		return
	}
	defer os.Remove(tempfile.Name())

	var merges []byte
	for i := 0; i < blockNumber; i++ {
		pname := fmt.Sprintf("partial.%d", i)
		defer os.Remove(pname)
		pfile, _ := os.Open(pname)
		//var buffer []byte
		buf := new(bytes.Buffer)
		buf.ReadFrom(pfile)
		//pfile.Read(buffer)
		log.Infof("%s has %d bytes", pname, len(buf.Bytes()))
		merges = append(merges, buf.Bytes()...)
	}
	log.Infof("Merged %d bytes\n", len(merges))
	cnt, err := tempfile.Write(merges)
	if err != nil {
		log.WithError(err).Error("Write merge file failed. ")
		return
	}
	log.Infof("Write merge file success with %d bytes\n", cnt)

	log.Infof("Untaring received file in receiver stream ..\n")
	if receivePath == "" {
		receivePath, err = os.Getwd()
		if err != nil {
			log.WithError(err).Errorf("Error in finding working directory in receiver stream \n")
			return
		}
	}

	err = utils.Untar(tempfile, receivePath)
	if err != nil {
		log.WithError(err).Errorf("Error in untaring file in receiver stream \n")
	}
}



func ReadLoop(stream *quic.BidirectionalStream, receivePath string, wg *sync.WaitGroup) {
	defer wg.Done()

	quicgoStream := stream.Detach()

	tempfile, err := ioutil.TempFile(os.TempDir(), "portkey*")
	if err != nil {
		log.WithError(err).Errorf("Error in tempfile creation in receiver stream %d\n", quicgoStream.StreamID())
		return
	}
	defer os.Remove(tempfile.Name())

	zstdReader := zstd.NewReader(quicgoStream)
	bytesWritten, err := io.Copy(tempfile, zstdReader)
	if err != nil {
		log.WithError(err).Errorf("Error in copying from zstdReader to tempfile in receiver stream %d\n", quicgoStream.StreamID())
	}
	log.Infof("Copied %d bytes from zstdReader to tempfile in receiver stream %d", bytesWritten, quicgoStream.StreamID())

	if err = zstdReader.Close(); err != nil {
		log.WithError(err).Errorf("Error in closing zstdWriter in receiver stream %d\n", quicgoStream.StreamID())
	}
	if err = quicgoStream.Close(); err != nil {
		log.WithError(err).Errorf("Error in closing stream %d\n", quicgoStream.StreamID())
	}
	log.Infof("Finished reading from stream %d\n", quicgoStream.StreamID())

	log.Infof("Untaring received file in receiver stream %d ...", quicgoStream.StreamID())
	if receivePath == "" {
		receivePath, err = os.Getwd()
		if err != nil {
			log.WithError(err).Errorf("Error in finding working directory in receiver stream %d\n", quicgoStream.StreamID())
			return
		}
	}

	err = utils.Untar(tempfile, receivePath)
	if err != nil {
		log.WithError(err).Errorf("Error in untaring file in receiver stream %d\n", quicgoStream.StreamID())
	}
}
