package log

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"log/syslog"
	"net/http"
	"runtime"
	"snoopd/cfg"
	"strconv"
	"strings"
)

var (
	logger 	       *syslog.Writer
	responseLogger *syslog.Writer
)

func init() {
	var err error
	logger, err = syslog.New(syslog.LOG_INFO, cfg.GetString("snoopd.log.logger_tag"))
	if err != nil {
		log.Fatal("Unable to create new logger:", err)
	}

	responseLogger, err = syslog.New(syslog.LOG_INFO, cfg.GetString("snoopd.log.access_logger_tag"))
	if err != nil {
		log.Fatal("Unable to create new access logger:", err)
	}
}

var (
	errorLogColor = color.New(color.FgRed).Add(color.Bold)
)

const (
	fatalLogPrefix   = "[FATA]"
	errorLogPrefix   = "[ERRO]"
	warningLogPrefix = "[WARN]"
	infoLogPrefix    = "[INFO]"
	debugLogPrefix   = "[DEBU]"
)

func getLogCaller() string {
	_, file, line, _ := runtime.Caller(2)
	return " <" + file + ":" + strconv.Itoa(line) + "> "
}

func Fatal(msg ...interface{}) {
	logCaller := getLogCaller()
	msgStr := fmt.Sprintln(msg...)
	logger.Err(fmt.Sprintln(fatalLogPrefix,logCaller, msgStr))
	log.Fatal(errorLogColor.Sprint(fatalLogPrefix, logCaller, msgStr))
}

func Error(msg ...interface{}) {
	if cfg.GetBool("snoopd.debug_mode") {
		fmt.Print(fmt.Sprint(errorLogPrefix, getLogCaller(), fmt.Sprintln(msg...)))
	} else {
		logger.Err(fmt.Sprint(errorLogPrefix, getLogCaller(), fmt.Sprintln(msg...)))
	}
}

func Warning(msg ...interface{}) {
	if cfg.GetBool("snoopd.debug_mode") {
		fmt.Print(fmt.Sprint(warningLogPrefix, getLogCaller(), fmt.Sprintln(msg...)))
	} else {
		logger.Warning(fmt.Sprint(warningLogPrefix, getLogCaller(), fmt.Sprintln(msg...)))
	}
}

func Info(msg ...interface{}) {
	if cfg.GetBool("snoopd.debug_mode") {
		fmt.Print(fmt.Sprint(infoLogPrefix, getLogCaller(), fmt.Sprintln(msg...)))
	} else {
		logger.Info(fmt.Sprint(infoLogPrefix, getLogCaller(), fmt.Sprintln(msg...)))
	}
}

func Debug(msg ...interface{}) {
	if !cfg.GetBool("snoopd.debug_mode") {
		return
	}
	fmt.Print(debugLogPrefix, getLogCaller(), fmt.Sprintln(msg...))
}

func Response(code int, reqStr, reqId string)  {
	//TODO remove
	responseLogger.Info(fmt.Sprintln("[" + strconv.Itoa(code) +"]", reqStr, "<" + reqId + ">"))
}

func Request(req *http.Request) {
	//TODO remove
	fmt.Printf("%v %v %v\r\n", req.Method, req.URL, req.Proto)
	for headerName, headerValues := range req.Header {
		fmt.Printf("%s:%s\r\n", headerName, fmt.Sprint(strings.Join(headerValues,"")))
	}
	//fmt.Print("\r\n")
	//var body bytes.Buffer
	//body.ReadFrom(req.Body)
	//if body.Len() != 0 {
	//	fmt.Printf("%s\r\n", body.String())
	//}
}