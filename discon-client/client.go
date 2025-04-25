package main

import "C"

import (
	dw "discon-wrapper"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"unsafe"

	"github.com/gorilla/websocket"
)

const program = "discon-client"
const version = "v0.1.0"

var debugLevel int = 0

var ws *websocket.Conn
var payload dw.Payload
var sentSwapFile *os.File
var recvSwapFile *os.File

func init() {

	// Print info
	fmt.Println("Loaded", program, version)

	// Get debug flag from environment variable
	csvFileName := "discon_swap"
	debugStr, debug := os.LookupEnv("DISCON_CLIENT_DEBUG")
	if debug {
		var err error
		debugLevel, err = strconv.Atoi(debugStr)
		if err != nil {
			// If not a number, treat as filename and set debug level to 1
			csvFileName = debugStr
			debugLevel = 1
		}

		// Only create log files if debugLevel > 0
		if debugLevel > 0 {
			sentSwapFile, err = os.Create(csvFileName + "_sent.csv")
			if err != nil {
				log.Fatal("discon-client: error creating sent swap file:", err)
			}
			recvSwapFile, err = os.Create(csvFileName + "_recv.csv")
			if err != nil {
				log.Fatal("discon-client: error creating recv swap file:", err)
			}
		}
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

	if debugLevel >= 1 {
		log.Println("discon-client: DISCON_SERVER_ADDR=", serverAddr)
		log.Println("discon-client: DISCON_LIB_PATH=", disconPath)
		log.Println("discon-client: DISCON_LIB_PROC=", disconFunc)
		log.Println("discon-client: DISCON_CLIENT_DEBUG=", debugLevel)
	}

	// Create a URL object
	u, err := url.Parse(fmt.Sprintf("ws://%s/ws", serverAddr))
	if err != nil {
		log.Fatal(err)
	}

	// Add query parameters for shared library path and proc
	u.RawQuery = url.Values{"path": {disconPath}, "proc": {disconFunc}}.Encode()

	if debugLevel >= 1 {
		log.Printf("discon-client: connecting to discon-server at '%s'\n", u.String())
	}

	// Connect to websocket server
	ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("discon-client: error connecting to discon-server at %s: %s", serverAddr, err)
	}

	if debugLevel >= 1 {
		log.Printf("discon-client: Connected to discon-server at '%s'\n", u.String())
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

	if debugLevel >= 2 {
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

	if debugLevel >= 2 {
		log.Println("discon-client: sent payload:\n", payload)
	}

	if debugLevel >= 1 && sentSwapFile != nil {
		outSwapSize := min(swapSize, 163)
		for _, v := range payload.Swap[:outSwapSize-1] {
			fmt.Fprintf(sentSwapFile, "%g,", v)
		}
		fmt.Fprintf(sentSwapFile, "%g\n", payload.Swap[outSwapSize-1])
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

	if debugLevel >= 2 {
		log.Println("discon-client: received payload:\n", payload)
	}

	if debugLevel >= 1 && recvSwapFile != nil {
		outSwapSize := min(swapSize, 163)
		for _, v := range payload.Swap[:outSwapSize-1] {
			fmt.Fprintf(recvSwapFile, "%g,", v)
		}
		fmt.Fprintf(recvSwapFile, "%g\n", payload.Swap[outSwapSize-1])
	}

	// Set fail flag
	*aviFail = C.int(payload.Fail)
}

func main() {}
