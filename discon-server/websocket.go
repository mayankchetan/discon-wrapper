package main

// #include <stdlib.h>
// void discon(int connID, float* avrSWAP, int* aviFAIL, char* accINFILE, char* avcOUTNAME, char* avcMSG);
// int load_shared_library(int connID, const char* library_path, const char* function_name);
// void unload_shared_library(int connID);
import "C"
import (
	"crypto/sha256"
	dw "discon-wrapper"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// Unique identifier for the websocket connection
var connectionID int32

// Mutex to protect the connectionID variable
var connectionIDMutex sync.Mutex

// WaitGroup to wait for all goroutines to finish
var wg sync.WaitGroup

// GH-Cp gen: Map of temporary files created for each connection
var tempFiles = make(map[int32][]string)
var tempFilesMutex sync.Mutex

// GH-Cp gen: Function to check if data is a file transfer
func isFileTransfer(payload *dw.Payload) bool {
	return len(payload.FileContent) > 0 && len(payload.ServerFilePath) > 0
}

// GH-Cp gen: Function to validate filename for security
func validateFileName(filename string) error {
	// Remove null terminators
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Check if filename contains suspicious patterns
	suspiciousPatterns := []string{"../", "/..", "~", "$", "|", ";", "&", "\\"}
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(filename, pattern) {
			return fmt.Errorf("filename contains invalid pattern: %s", pattern)
		}
	}

	// Ensure filename is just a basename, not a path
	if filepath.Base(filename) != filename {
		return fmt.Errorf("filename must not contain path separators")
	}

	return nil
}

// GH-Cp gen: Handle file transfer from client to server
func handleFileTransfer(connID int32, payload *dw.Payload, debugLevel int) (*dw.Payload, error) {
	// Initialize response payload
	response := dw.Payload{
		Swap:    make([]float32, 1),
		Fail:    0,
		InFile:  []byte{0},
		OutName: []byte{0},
		Msg:     make([]byte, 256), // Reserve space for error message
	}

	// Get server file path from payload
	serverFilePath := string(payload.ServerFilePath)
	nullIndex := strings.IndexByte(serverFilePath, 0)
	if nullIndex >= 0 {
		serverFilePath = serverFilePath[:nullIndex]
	}

	// Validate filename for security
	err := validateFileName(serverFilePath)
	if err != nil {
		response.Fail = 1
		errMsg := fmt.Sprintf("Security error: %v", err)
		copy(response.Msg, []byte(errMsg+"\x00"))
		return &response, fmt.Errorf("file validation error: %w", err)
	}

	// Verify file contents with hash
	hash := sha256.New()
	hash.Write(payload.FileContent)
	contentHash := hex.EncodeToString(hash.Sum(nil))

	if debugLevel >= 1 {
		log.Printf("Received file transfer request for %s (size: %d bytes, hash: %s)",
			serverFilePath, len(payload.FileContent), contentHash[:8])
	}

	// Create the file with the server file path
	file, err := os.Create(serverFilePath)
	if err != nil {
		response.Fail = 1
		errMsg := fmt.Sprintf("Failed to create file: %v", err)
		copy(response.Msg, []byte(errMsg+"\x00"))
		return &response, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write the file contents
	_, err = file.Write(payload.FileContent)
	if err != nil {
		response.Fail = 1
		errMsg := fmt.Sprintf("Failed to write file: %v", err)
		copy(response.Msg, []byte(errMsg+"\x00"))
		os.Remove(serverFilePath) // Clean up the partial file
		return &response, fmt.Errorf("failed to write file: %w", err)
	}

	// Register the file to be cleaned up when the connection closes
	tempFilesMutex.Lock()
	tempFiles[connID] = append(tempFiles[connID], serverFilePath)
	tempFilesMutex.Unlock()

	if debugLevel >= 1 {
		log.Printf("File %s created successfully", serverFilePath)
	}

	// Set success message
	successMsg := fmt.Sprintf("File transferred successfully: %s", serverFilePath)
	copy(response.Msg, []byte(successMsg+"\x00"))

	return &response, nil
}

func ServeWs(w http.ResponseWriter, r *http.Request, debugLevel int) {
	wg.Add(1)
	defer wg.Done()

	// Get unique identifier for this connection
	connectionIDMutex.Lock()
	connID := connectionID
	connectionID++
	if connectionID > 8191 {
		connectionID = 0
	}
	connectionIDMutex.Unlock()

	// Read controller path and function name from post parameters
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Error parsing url parameters: "+err.Error(), http.StatusInternalServerError)
		return
	}
	path := params.Get("path")
	proc := params.Get("proc")

	if debugLevel >= 1 {
		log.Printf("Received request to load function '%s' from shared controller '%s'\n", proc, path)
	}

	// Check if controller exists at path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		http.Error(w, "Controller not found at '"+path+"'", http.StatusInternalServerError)
		return
	}

	// Create a copy of the shared library with a number suffix so multiple instances
	// of the same library can be shared at the same time
	tmpPath, err := duplicateLibrary(path, connID)
	if err != nil {
		http.Error(w, "Error duplicating controller: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpPath)

	if debugLevel >= 1 {
		log.Printf("Duplicated controller to '%s'\n", tmpPath)
	}

	// GH-Cp gen: Initialize tempFiles entry for this connection
	tempFilesMutex.Lock()
	tempFiles[connID] = make([]string, 0)
	tempFilesMutex.Unlock()

	// GH-Cp gen: Clean up any temporary files when connection closes
	defer func() {
		tempFilesMutex.Lock()
		fileList := tempFiles[connID]
		delete(tempFiles, connID)
		tempFilesMutex.Unlock()

		// Remove all temporary files for this connection
		for _, filePath := range fileList {
			if debugLevel >= 1 {
				log.Printf("Cleaning up temporary file: %s", filePath)
			}
			os.Remove(filePath)
		}
	}()

	// Load the shared library
	libraryPath := C.CString(tmpPath)
	defer C.free(unsafe.Pointer(libraryPath))
	functionName := C.CString(proc)
	defer C.free(unsafe.Pointer(functionName))
	status := C.load_shared_library(C.int(connID), libraryPath, functionName)
	if status == 1 {
		http.Error(w, "Error loading shared library", http.StatusInternalServerError)
		return
	} else if status == 2 {
		http.Error(w, "Error loading function from shared library", http.StatusInternalServerError)
		return
	}

	if debugLevel >= 1 {
		log.Printf("Library and function loaded successfully\n")
	}

	// Convert connection to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer ws.Close()

	// Create payload structure
	payload := dw.Payload{}

	// Loop while receiving messages over socket
	for {

		// If not in debug mode, set read deadline to 5 seconds
		// This will disconnect the client if no message is received in 5 seconds
		// which allows the controller to be unloaded and deleted
		if debugLevel == 0 {
			ws.SetReadDeadline(time.Now().Add(time.Second * 5))
		}

		// Read message from websocket
		messageType, b, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		if messageType != websocket.BinaryMessage {
			log.Println("message type:", messageType)
			continue
		}

		err = payload.UnmarshalBinary(b)
		if err != nil {
			log.Println("payload.UnmarshalBinary:", err)
			break
		}

		// GH-Cp gen: Only log full payload at debug level 2
		if debugLevel >= 2 {
			log.Println("discon-server: received payload:", payload)
			// } else if debugLevel == 1 {
			// 	// At level 1, just log basic info without full payload details
			// 	inFilePath := string(payload.InFile)
			// 	nullIndex := strings.IndexByte(inFilePath, 0)
			// 	if nullIndex >= 0 {
			// 		inFilePath = inFilePath[:nullIndex]
			// 	}
			// 	log.Printf("discon-server: received request with InFile: %s", inFilePath)
		}

		// GH-Cp gen: Check if the payload is a file transfer
		if isFileTransfer(&payload) {
			response, err := handleFileTransfer(connID, &payload, debugLevel)
			if err != nil {
				log.Println("handleFileTransfer:", err)
			}
			b, err = response.MarshalBinary()
			if err != nil {
				log.Println("response.MarshalBinary:", err)
				break
			}
			err = ws.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				log.Println("write:", err)
				break
			}
			continue
		}

		// Call the function from the shared library with data in payload
		C.discon(C.int(connID),
			(*C.float)(unsafe.Pointer(&payload.Swap[0])),
			(*C.int)(unsafe.Pointer(&payload.Fail)),
			(*C.char)(unsafe.Pointer(&payload.InFile[0])),
			(*C.char)(unsafe.Pointer(&payload.OutName[0])),
			(*C.char)(unsafe.Pointer(&payload.Msg[0])))

		// Convert payload to binary and send over websocket
		b, err = payload.MarshalBinary()
		if err != nil {
			log.Println("payload.MarshalBinary:", err)
			break
		}
		err = ws.WriteMessage(websocket.BinaryMessage, b)
		if err != nil {
			log.Println("write:", err)
			break
		}

		// GH-Cp gen: Only log full response at debug level 2
		if debugLevel >= 2 {
			log.Println("discon-server: sent payload:", payload)
			// } else if debugLevel == 1 {
			// 	// At level 1, just log that response was sent
			// 	log.Println("discon-server: sent response")
		}
	}

	// Unload the shared library
	C.unload_shared_library(C.int(connID))
}

func duplicateLibrary(path string, connID int32) (string, error) {
	// Create a copy of the shared library with a number suffix so multiple instances
	// of the same library can be shared at the same time
	outFile, err := os.CreateTemp(".", fmt.Sprintf("%s-%03d-", filepath.Base(path), connID))
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	inFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer inFile.Close()

	// Copy the file contents
	_, err = io.Copy(outFile, inFile)
	if err != nil {
		return "", err
	}

	return outFile.Name(), nil
}
