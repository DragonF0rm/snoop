package log

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"log/syslog"
	"net/http"
	"runtime"
	"snoop/src/shared/cfg"
	"strconv"
	"strings"
)

var (
	logger       *syslog.Writer
	accessLogger *syslog.Writer
)

func init() {
	var err error
	logger, err = syslog.New(syslog.LOG_INFO, cfg.GetString("snoopd.log.logger_tag"))
	if err != nil {
		log.Fatal("Unable to create new logger:", err)
	}

	accessLogger, err = syslog.New(syslog.LOG_INFO, cfg.GetString("snoopd.log.access_logger_tag"))
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
	logLine := fmt.Sprintln(fatalLogPrefix,logCaller, msgStr)
	logger.Err(logLine)
	log.Fatal(logLine)
}

func Error(msg ...interface{}) {
	logLine := fmt.Sprint(errorLogPrefix, getLogCaller(), fmt.Sprintln(msg...))
	if cfg.GetBool("snoopd.debug_mode") {
		fmt.Print(logLine)
	} else {
		logger.Err(logLine)
	}
}

func Warning(msg ...interface{}) {
	logLine := fmt.Sprint(warningLogPrefix, getLogCaller(), fmt.Sprintln(msg...))
	if cfg.GetBool("snoopd.debug_mode") {
		fmt.Print(logLine)
	} else {
		logger.Warning(logLine)
	}
}

func Info(msg ...interface{}) {
	logLine := fmt.Sprint(infoLogPrefix, getLogCaller(), fmt.Sprintln(msg...))
	if cfg.GetBool("snoopd.debug_mode") {
		fmt.Print(logLine)
	} else {
		logger.Info(logLine)
	}
}

func Debug(msg ...interface{}) {
	if !cfg.GetBool("snoopd.debug_mode") {
		return
	}
	fmt.Print(debugLogPrefix, getLogCaller(), fmt.Sprintln(msg...))
}

func Access(code int, reqStr, reqId string)  {
	logLine := fmt.Sprintln("[" + strconv.Itoa(code) +"]", reqStr, "< " + reqId + " >")
	if cfg.GetBool("snoopd.debug_mode") {
		fmt.Print(logLine)
	}
	accessLogger.Info(logLine)
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