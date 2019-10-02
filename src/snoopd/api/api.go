package api

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"snoop/src/shared/cfg"
	"snoop/src/shared/log"
	"snoop/src/shared/protobuf"
	"strconv"
	"strings"
)

const historyLen = 20

type GrpcApiService struct {
	accessLogPath string
}

var storingPath = cfg.GetString("snoopd.storing_path")

func NewGrpcApiService(accessLogPath string) GrpcApiService {
	return GrpcApiService{
		accessLogPath: accessLogPath,
	}
}

func (apiService *GrpcApiService)GetHistory(ctx context.Context, in *protobuf.Nothing)(*protobuf.History, error) {
	cmd := exec.Command("tail", "-n", strconv.Itoa(historyLen), cfg.GetString("snoopd.log.access_logger_file"))
	output, err := cmd.Output()
	if err != nil {
		log.Error("Unable to grep access log, err:", err)
		return nil, err
	}

	var history protobuf.History
	roundTripRecords := strings.Split(string(output), "\n")
	for _, roundTripRecord := range roundTripRecords {
		history.RoundTrips = append(history.RoundTrips, roundTripRecord)
	}
	return &history, nil
}

var ErrNoStoredReqFound = errors.New("no stored request with this id found")

func (apiService *GrpcApiService)Resend(ctx context.Context, in *protobuf.ReqID)(*protobuf.Response, error) {
	storedReqFiles, err := ioutil.ReadDir(storingPath)
	if err != nil {
		log.Error("Unable to read storing path directory, err:", err)
		return nil, err
	}

	var reqFileInfo *os.FileInfo
	for i, storedReqFile := range storedReqFiles {
		if strings.HasPrefix(storedReqFile.Name(), in.ID) {
			reqFileInfo = &storedReqFiles[i]
			break
		}
	}

	if reqFileInfo == nil {
		return nil, ErrNoStoredReqFound
	}

	reqBytes, err := ioutil.ReadFile(filepath.Join(storingPath, (*reqFileInfo).Name()))
	if err != nil {
		log.Error("Unable to read request file, err:", err)
		return nil, err
	}

	reqBuf := bufio.NewReader(bytes.NewBuffer(reqBytes))
	req, err := http.ReadRequest(reqBuf)
	if err != nil {
		log.Error("Unable to read request from buffer, err:", err)
		return nil, err
	}
	req.Body.Close()

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Error("Unable to get response for stored request, err:", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBuf := bytes.NewBuffer(make([]byte, 0))
	_, err = respBuf.WriteTo(respBuf)
	if err != nil {
		log.Error("Unable to write response into buffer, err:", err)
		return nil, err
	}

	var protoResponse protobuf.Response
	protoResponse.Response = respBuf.String()
	return &protoResponse, nil
}
