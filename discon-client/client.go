package main

import "C"

import (
	dw "discon-wrapper"
	"fmt"
	"log"
	"net/url"
	"os"
	"unsafe"

	"github.com/gorilla/websocket"
)

const program = "discon-client"
const version = "v0.1.0"

var debug = false

var ws *websocket.Conn
var payload dw.Payload

func init() {

	// Print info
	fmt.Println("Loaded", program, version)

	// Get debug flag from environment variable
	if len(os.Getenv("DISCON_CLIENT_DEBUG")) > 0 {
		debug = true
	}

	// Get discon-server address from environment variable
	serverAddr, found := os.LookupEnv("DISCON_SERVER_ADDR")
	if !found {
		log.Fatal("discon-client: environment variable DISCON_SERVER_ADDR not set (e.g. 'localhost:8080')")
	}

	// Get shared library path from environment variable
	disconPath, found := os.LookupEnv("DISCON_LIB_PATH")
	if !found {
		log.Fatal("discon-client: environment variable DISCON_LIB_PATH not set (e.g. 'discon.dll')")
	}

	// Get shared library function name from environment variable
	disconFunc, found := os.LookupEnv("DISCON_LIB_PROC")
	if !found {
		log.Fatal("discon-client: environment variable DISCON_LIB_PROC not set (e.g. 'discon')")
	}

	if debug {
		log.Println("discon-client: DISCON_SERVER_ADDR=", serverAddr)
		log.Println("discon-client: DISCON_LIB_PATH=", disconPath)
		log.Println("discon-client: DISCON_LIB_PROC=", disconFunc)
		log.Println("discon-client: DISCON_CLIENT_DEBUG=", debug)
	}

	// Create a URL object
	u, err := url.Parse(fmt.Sprintf("ws://%s/ws", serverAddr))
	if err != nil {
		log.Fatal(err)
	}

	// Add query parameters for shared library path and proc
	u.RawQuery = url.Values{"path": {disconPath}, "proc": {disconFunc}}.Encode()

	if debug {
		log.Printf("discon-client: connecting to discon-server at '%s'\n", u.String())
	}

	// Connect to websocket server
	ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("discon-client: error connecting to discon-server at %s: %s", serverAddr, err)
	}

	if debug {
		log.Printf("discon-client: Connecting to discon-server at '%s'\n", u.String())
	}
}

//export DISCON
func DISCON(avrSwap *C.float, aviFail *C.int, accInFile, avcOutName, avcMsg *C.char) {

	// Get first 130 entries of swap array
	swap := (*[1 << 24]float32)(unsafe.Pointer(avrSwap))

	// Get array sizes
	swapSize := int(swap[128])   // Maximum size of swap array
	inFileSize := int(swap[49])  // Maximum size of inFile string
	outNameSize := int(swap[63]) // Maximum size of outName string
	msgSize := int(swap[48])     // Maximum size of msg string

	// Resize payload arrays to match sizes
	if len(payload.Swap) != swapSize {
		payload.Swap = make([]float32, swapSize)
	}

	if debug {
		log.Printf("discon-client: size of avrSWAP:    % 5d\n", swapSize)
		log.Printf("discon-client: size of accINFILE:  % 5d\n", inFileSize)
		log.Printf("discon-client: size of avcOUTNAME: % 5d\n", outNameSize)
		log.Printf("discon-client: size of avcMSG:     % 5d\n", msgSize)
	}

	payload.Swap = swap[:swapSize:swapSize]
	payload.Fail = int32(*aviFail)
	payload.InFile = (*[1 << 24]byte)(unsafe.Pointer(accInFile))[:inFileSize:inFileSize]
	payload.OutName = (*[1 << 24]byte)(unsafe.Pointer(avcOutName))[:outNameSize:outNameSize]
	payload.Msg = (*[1 << 24]byte)(unsafe.Pointer(avcMsg))[:msgSize:msgSize]

	// Convert payload to binary and send over websocket
	b, err := payload.MarshalBinary()
	if err != nil {
		log.Fatalf("discon-client: %s", err)
	}
	ws.WriteMessage(websocket.BinaryMessage, b)

	if debug {
		log.Println("discon-client: sent payload:\n", payload)
	}

	// Read response from server
	_, b, err = ws.ReadMessage()
	if err != nil {
		log.Fatalf("discon-client: %s", err)
	}

	// Unmarshal binary data into payload
	err = payload.UnmarshalBinary(b)
	if err != nil {
		log.Fatalf("discon-client: %s", err)
	}

	if debug {
		log.Println("discon-client: received payload:\n", payload)
	}

	// Set fail flag
	*aviFail = C.int(payload.Fail)
}

func main() {}
