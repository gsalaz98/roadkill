package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
)

// TODO: Decide whether to load from JSON config file, or from environment variables

type TectonicTick struct {
}

// Tectonic : Main type for single-instance connection to the Tectonic database
type Tectonic struct {
	// Connection settings
	Host       string
	Port       uint16 // type ensures port selected is valid
	Connection net.Conn

	// TODO: Create authentication in TectonicDB project, then these will be functional
	Username string
	Password string

	CurrentDB       string
	CurrentSymbol   string
	CurrentExchange string
}

// TectonicDB function prototypes
// ****************************
// Help()										( string, error )
// Ping()										( string, error )
// Info()										( []string, error )
// Perf()										( []string, error )
// BulkAdd(ticks string)						error
// BulkAddInto(dbName string) 					error
// Use(dbName string) 							error
// Create(dbName string) 						error
// Get(amount int) 								( []string, error )
// GetFrom(amount int, dbName string)			( []string, error )
// Insert(tick string)							error
// Count()										( uint64, error )
// CountAll()									( uint64, error )
// Clear()										error
// ClearAll()									error
// Flush()										error
// FlushDoAll()									error
// Subscribe(dbName, message chan string)		error
// Unsubscribe()								error
// Exists(dbName string)						( bool, error )
//
// Locally defined methods:
// ****************************
// Connect()									error
//
// GetAsDelta(amount int)						( orderbook.DeltaBatch, error )
// GetFromAsDelta(amount int, dbName string)	( orderbook.DeltaBatch, error )
// BulkAddDelta(deltas orderbook.DeltaBatch)	error
// BulkAddDeltaInto(dbName string, deltas)		error
// InsertDelta(delta orderbook.Delta)			error
// ****************************

// TectonicPool : TODO
type TectonicPool struct{}

// DefaultTectonic : Default settings for Tectonic structure
var DefaultTectonic = Tectonic{
	Host: "127.0.0.1",
	Port: 9001,
}

// Connect : Connects Tectonic instance to the database. Run to initialize
func (t *Tectonic) Connect() error {
	var (
		connectAddressBuf = bytes.Buffer{}
		connectErr        error
	)

	// Create connection string via concatenation using buffer
	connectAddressBuf.WriteString(t.Host)
	connectAddressBuf.WriteByte(':')
	connectAddressBuf.WriteString(strconv.Itoa(int(t.Port)))

	t.Connection, connectErr = net.Dial("tcp", connectAddressBuf.String())

	return connectErr
}

// SendMessage : Sends message to TectonicDB
func (t *Tectonic) SendMessage(message string) (string, error) {
	var (
		messageBuf = bytes.Buffer{}
		readBuf    = make([]byte, (1<<15)-1)
	)

	messageBuf.WriteString(message)
	messageBuf.WriteString("\n")

	_, _ = t.Connection.Write(messageBuf.Bytes())
	_, readErr := t.Connection.Read(readBuf)

	return string(readBuf), readErr
}

// SendBulkMessage : TODO
//func (t *Tectonic) SendBulkMessage(messageHeader string, messages []string, messageTail string) error {
//
//}

// Help : Return help string from Tectonic server
func (t *Tectonic) Help() (string, error) {
	return t.SendMessage("HELP")
}

// Ping : Sends a ping message to the TectonicDB server
func (t *Tectonic) Ping() (string, error) {
	return t.SendMessage("PING")
}

// Info : From official documentation: "Returns info about table schemas"
func (t *Tectonic) Info() (string, error) {
	return t.SendMessage("INFO")
}

// Perf : From official documentation: "Returns the answercount of items over time"
func (t *Tectonic) Perf() (string, error) {
	return t.SendMessage("PERF")
}

// BulkAdd : TODO
func (t *Tectonic) BulkAdd(ticks []string) error {
	_, _ = t.SendMessage("BULKADD")

	for _, tick := range ticks {
		_, _ = t.SendMessage(tick)
	}

	_, recvErr := t.SendMessage("DDAKLUB")

	return recvErr
}

// BulkAddInto : TODO
func (t *Tectonic) BulkAddInto(dbName string, ticks []string) error {
	msgBuf := bytes.Buffer{}
	msgBuf.WriteString("BULKADD INTO ")
	msgBuf.WriteString(dbName)

	_, _ = t.SendMessage(msgBuf.String())

	for _, tick := range ticks {
		_, _ = t.SendMessage(tick)
	}

	_, recvErr := t.SendMessage("DDAKLUB")

	return recvErr
}

// Use : "Switch the current store"
func (t *Tectonic) Use(dbName string) error {
	t.CurrentDB = dbName

	msgBuf := bytes.Buffer{}

	msgBuf.WriteString("USE ")
	msgBuf.WriteString(dbName)

	_, readErr := t.SendMessage(msgBuf.String())
	return readErr
}

// Create : "Create store"
func (t *Tectonic) Create(dbName string) error {
	msgBuf := bytes.Buffer{}

	msgBuf.WriteString("CREATE ")
	msgBuf.WriteString(dbName)

	_, readErr := t.SendMessage(msgBuf.String())
	return readErr
}

// Get : "Returns `amount` items from current store"
//func (t *Tectonic) Get(amount uint64) ([]string, error) {
//
//}

//func (t *Tectonic) GetAsDelta() (string, error)     {}
//func (t *Tectonic) GetFrom() (string, error)        {}
//func (t *Tectonic) GetFromAsDelta() (string, error) {}
//func (t *Tectonic) Count() (string, error)          {}
//func (t *Tectonic) CountAll() (string, error)       {}
//func (t *Tectonic) Clear() (string, error)          {}
//func (t *Tectonic) ClearAll() (string, error)       {}
//func (t *Tectonic) Flush() (string, error)          {}
//func (t *Tectonic) FlushDoAll() (string, error)     {}
//func (t *Tectonic) Subscribe() (string, error)      {}
//func (t *Tectonic) Unsubscribe() (string, error)    {}
//func (t *Tectonic) Exists() (string, error)         {}

// DEBUG: remove later
func main() {
	connection := DefaultTectonic

	connErr := connection.Connect()

	if connErr != nil {
		fmt.Println(connErr)
		return
	}
	msg, err := connection.Info()

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(msg)
}
