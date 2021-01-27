package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alexpop/data-proxy/jlog"
	"github.com/alexpop/data-proxy/types"
)

func RightSplit(stringToSplit string, delimiter string) (string, string) {
	n := strings.LastIndex(stringToSplit, delimiter)
	if n >= 0 {
		return stringToSplit[0:n], stringToSplit[n+1:]
	}
	return stringToSplit, ""
}

// Generates a pretty printed json string from the object passed in
func PrettyPrintJson(v interface{}) string {
	outBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		// Should not be a common error, ok to eat it and keep the code clean outside
		jlog.Error(fmt.Sprintf("ERROR: PrettyPrintJson error '%s' for v='%+v'", err.Error(), v))
	}
	outBytes = append(outBytes, "\n"...)
	return fmt.Sprintf("%-512s\n", outBytes) // Pad right to avoid "Friendly HTTP error messages" issue in IE & Chrome
}

func JsonErrorIt(logData *types.JsonProxyLog, msg string, responseCode int) string {
	jsonErr := types.JsonResponse{
		Error: msg,
	}
	logData.Body = msg
	logData.Status = uint16(responseCode)
	jlog.Proxy(logData)
	outBytes, err := json.MarshalIndent(jsonErr, "", "  ")
	if err != nil {
		// Should not be a common error, ok to eat it and keep the code clean outside
		jlog.Error(fmt.Sprintf("ERROR: JsonErrorIt marshal error '%s'", err.Error()))
	}
	return fmt.Sprintf("%-512s\n", outBytes) // Pad right to avoid "Friendly HTTP error messages" issue in IE & Chrome
}

func GetFileSha256(file string) (sha256Val string, err error) {
	openFile, err := os.Open(file)
	if err != nil {
		return sha256Val, fmt.Errorf("open error: %s", err.Error())
	}
	defer openFile.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, openFile); err != nil {
		return sha256Val, fmt.Errorf("calculating sha256 error: %s", err.Error())
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
