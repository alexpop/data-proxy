package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"../types"
)

func RightSplit(stringToSplit string, delimiter string) (string, string) {
	n := strings.LastIndex(stringToSplit, delimiter)
	if n >= 0 {
		return stringToSplit[0:n], stringToSplit[n+1:]
	}
	return stringToSplit, ""
}

// Generates a pretty printed json string from the object passed in
func PrettyPrintJson(v interface{}) []byte {
	outBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		// Should not be common, ok to eat it and keep the code clean in the api
		log.Printf("ERROR: PrettyPrintJson error '%s' for v='%+v'", err.Error(), v)
	}
	outBytes = append(outBytes, "\n"...)
	return outBytes
}

func JsonErrorIt(msg string) string {
	jsonErr := types.JsonResponse{
		Error: msg,
	}
	log.Printf(" < JsonErrorIt: %s\n", jsonErr.Error)
	outBytes, err := json.MarshalIndent(jsonErr, "", "  ")
	if err != nil {
		// Should not be common, ok to eat it and keep the code clean in the api
		log.Printf("ERROR: JsonErrorIt error '%s' for jsonErr='%+v'", err.Error(), jsonErr)
	}
	return string(outBytes)
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
