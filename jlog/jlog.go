package jlog

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/alexpop/data-proxy/types"
)

const (
	LevelNone  = 0
	LevelInfo  = 1
	LevelDebug = 2
	LevelError = 3
)

var Level = LevelInfo // Default log level is Info

func Info(message string) {
	if Level >= LevelInfo {
		jsonPrint(types.JsonRootLog{Info: message})
	}
}

func Debug(message string) {
	if Level >= LevelDebug {
		jsonPrint(types.JsonRootLog{Debug: message})
	}
}

func Error(message string) {
	if Level >= LevelError {
		jsonPrint(types.JsonRootLog{Debug: message})
	}
}

func Fatal(message string) {
	log.Fatal(message)
}

func Proxy(logData *types.JsonProxyLog) {
	if Level >= LevelInfo {
		if logData.Status == 0 {
			logData.Status = http.StatusOK
		}
		jsonPrint(types.JsonRootLog{ProxyLog: logData})
	}
}

func jsonPrint(jsonData types.JsonRootLog) {
	jsonData.Time = time.Now().UTC().Format(time.RFC3339)
	b, _ := json.Marshal(jsonData)
	println(string(b))
}
