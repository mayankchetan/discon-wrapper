package main

// #include <stdlib.h>
// void discon(int connID, float* avrSWAP, int* aviFAIL, char* accINFILE, char* avcOUTNAME, char* avcMSG);
// int load_shared_library(int connID, const char* library_path, const char* function_name);
// void unload_shared_library(int connID);
import "C"
import (
	dw "discon-wrapper"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

func ServeWs(w http.ResponseWriter, r *http.Request, debug bool) {
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

	if debug {
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

	if debug {
		log.Printf("Duplicated controller to '%s'\n", tmpPath)
	}

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

	if debug {
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

		// If not debug, set read deadline to 5 seconds
		// This will disconnect the client if no message is received in 5 seconds
		// which allows the controller to be unloaded and deleted
		if !debug {
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

		if debug {
			log.Println("discon-server: received payload:", payload)
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

		if debug {
			fmt.Println("discon-server: sent payload:", payload)
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
