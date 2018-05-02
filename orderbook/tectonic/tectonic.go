package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"

	"github.com/pquerna/ffjson/ffjson"
)

// TectonicTick : Contains data that was loaded in from the datastore
type TectonicTick struct {
	Timestamp float64 `json:"ts"`
	Seq       uint64  `json:"seq"`
	IsTrade   bool    `json:"is_trade"`
	IsBid     bool    `json:"is_bid"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
}

// Tectonic : Main type for single-instance connection to the Tectonic database
type Tectonic struct {
	// Connection settings
	Host       string
	Port       uint16 // type ensures port selected is valid
	Connection net.Conn

	// TODO: Create authentication mechanisms in TectonicDB project, then these will be functional
	Username string
	Password string

	CurrentDB       string
	CurrentSymbol   string
	CurrentExchange string
}

// TectonicDB function prototypes
// ****************************
// Help()										( string, error )		done
// Ping()										( string, error )		done
// Info()										( string, error )		done
// Perf()										( string, error )		done
// BulkAdd(ticks TTick)							error					done
// BulkAddInto(dbName string, ticks TTick)		error					done
// Use(dbName string) 							error					done
// Create(dbName string) 						error					done
// Get(amount int) 								( []TTick, error )		done
// GetFrom(amount int, dbName string)			( []TTick, error )		done
// Insert(t TTick)								error					incomplete
// InsertInto(dbName string, t TTick)			error					incomplete
// Count()										uint64					done
// CountAll()									uint64					done
// Clear()										error					done
// ClearAll()									error					done
// Flush()										error					done
// FlushAll()									error					done
// Subscribe(dbName, message chan string)		error					incomplete
// Unsubscribe()								error					incomplete
// Exists(dbName string)						bool					done
//
// Locally defined methods:
// ****************************
// Connect()									error
// SendMessage()								( string, error )
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
		connectAddress = fmt.Sprintf("%s:%d", t.Host, t.Port)
		connectErr     error
	)

	t.Connection, connectErr = net.Dial("tcp", connectAddress)

	return connectErr
}

// SendMessage : Sends message to TectonicDB
func (t *Tectonic) SendMessage(message string) (string, error) {
	var readBuf = make([]byte, (1 << 15))

	_, _ = t.Connection.Write([]byte(message + "\n"))
	_, readErr := t.Connection.Read(readBuf)

	return string(readBuf), readErr
}

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
func (t *Tectonic) BulkAdd(ticks []TectonicTick) error {
	_, _ = t.SendMessage("BULKADD")

	for _, tick := range ticks {
		var (
			isTrade = "f"
			isBid   = "f"
		)
		if tick.IsTrade {
			isTrade = "t"
		}
		if tick.IsBid {
			isBid = "t"
		}

		_, _ = t.SendMessage(fmt.Sprintf("%.3f, %d, %s, %s, %f, %f;", tick.Timestamp, tick.Seq, isTrade, isBid, tick.Price, tick.Size))
	}

	_, recvErr := t.SendMessage("DDAKLUB")

	return recvErr
}

// BulkAddInto : TODO
func (t *Tectonic) BulkAddInto(dbName string, ticks []TectonicTick) error {
	_, _ = t.SendMessage("BULKADD INTO " + dbName)

	for _, tick := range ticks {
		var (
			isTrade = "f"
			isBid   = "f"
		)
		if tick.IsTrade {
			isTrade = "t"
		}
		if tick.IsBid {
			isBid = "t"
		}

		_, _ = t.SendMessage(fmt.Sprintf("%.3f, %d, %s, %s, %f, %f;", tick.Timestamp, tick.Seq, isTrade, isBid, tick.Price, tick.Size))
	}

	_, recvErr := t.SendMessage("DDAKLUB")

	return recvErr
}

// Use : "Switch the current store"
func (t *Tectonic) Use(dbName string) error {
	_, readErr := t.SendMessage("USE " + dbName)

	if readErr == nil {
		t.CurrentDB = dbName
	}

	return readErr
}

// Create : "Create store"
func (t *Tectonic) Create(dbName string) error {
	_, readErr := t.SendMessage("CREATE " + dbName)
	return readErr
}

// Get : "Returns `amount` items from current store"
func (t *Tectonic) Get(amount uint64) ([]TectonicTick, error) {
	// We use a buffer here to make it easier to maintain
	var (
		msgBuf  = bytes.Buffer{}
		msgJSON = []TectonicTick{}
	)
	msgBuf.WriteString("GET ")
	msgBuf.WriteString(strconv.Itoa(int(amount)))
	msgBuf.WriteString(" AS JSON")

	msgRecv, recvErr := t.SendMessage(msgBuf.String())
	ffjson.Unmarshal(bytes.Trim([]byte(msgRecv[9:]), "\x00"), &msgJSON) // We get back a message starting with `\uFFFE` - Trim that and all null chars in array

	return msgJSON, recvErr
}

// GetFrom : Returns items from specified store
func (t *Tectonic) GetFrom(dbName string, amount uint64, asTick bool) ([]TectonicTick, error) {
	// We use a buffer here to make it easier to maintain
	var (
		msgBuf  = bytes.Buffer{}
		msgJSON = []TectonicTick{}
	)
	msgBuf.WriteString("GET ")
	msgBuf.WriteString(strconv.Itoa(int(amount)))
	msgBuf.WriteString(" FROM ")
	msgBuf.WriteString(dbName)
	msgBuf.WriteString(" AS JSON")

	msgRecv, recvErr := t.SendMessage(msgBuf.String())
	ffjson.Unmarshal(bytes.Trim([]byte(msgRecv[9:]), "\x00"), &msgJSON) // We get back a message starting with `\uFFFE` - Trim that and all null chars in array

	return msgJSON, recvErr
}

// Insert : Inserts a tick into the database
//func (t *Tectonic) Insert(tick TectonicTick) error {
//
//}
//
//func (t *Tectonic) InsertInto(dbName, tick TectonicTick) error {
//
//}

// Count : "Count of items in current store"
func (t *Tectonic) Count() uint64 {
	msg, _ := t.SendMessage("COUNT")
	count, _ := strconv.Atoi(msg)

	return uint64(count)
}

// CountAll : "Returns total count from all stores"
func (t *Tectonic) CountAll() uint64 {
	msg, _ := t.SendMessage("COUNT ALL")
	count, _ := strconv.Atoi(msg)

	return uint64(count)
}

// Clear : Deletes everything in current store (BE CAREFUL WITH THIS METHOD)
func (t *Tectonic) Clear() (string, error) {
	return t.SendMessage("CLEAR")
}

// ClearAll : "Drops everything in memory"
func (t *Tectonic) ClearAll() (string, error) {
	return t.SendMessage("CLEAR ALL")
}

// Flush : "Flush current store to disk"
func (t *Tectonic) Flush() (string, error) {
	return t.SendMessage("FLUSH")
}

// FlushAll : "Flush everything form memory to disk"
func (t *Tectonic) FlushAll() (string, error) {
	return t.SendMessage("FLUSH ALL")
}

//func (t *Tectonic) Subscribe(dbName, message chan string) (string, error) {
//
//}
//
//func (t *Tectonic) Unsubscribe() (string, error)    {
//
//}

// Exists : Checks if datastore exists
func (t *Tectonic) Exists(dbName string) bool {
	msg, _ := t.SendMessage("EXISTS " + dbName)

	// EXISTS command returns `1` for an existing datastore, and `ERR:...` otherwise.
	return msg[0] == 1
}

// DEBUG: remove later
func main() {
	connection := DefaultTectonic

	connErr := connection.Connect()

	if connErr != nil {
		fmt.Println(connErr)
		return
	}

	connection.ClearAll()

	if true {
		err := connection.BulkAddInto("testing", []TectonicTick{
			TectonicTick{Timestamp: 1505177059.684, Seq: 139010, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177069.685, Seq: 139011, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177079.685, Seq: 139012, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177089.685, Seq: 139013, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177099.685, Seq: 139014, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177019.685, Seq: 139015, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177029.685, Seq: 139016, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177039.685, Seq: 139017, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177049.685, Seq: 139018, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177159.685, Seq: 139019, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177259.685, Seq: 139020, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177359.685, Seq: 139021, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177459.685, Seq: 139022, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177559.685, Seq: 139023, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
			TectonicTick{Timestamp: 1505177659.685, Seq: 139024, IsTrade: true, IsBid: false, Price: 0.0703620, Size: 7.65064240},
		})
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	_ = connection.Use("testing")

	msg, _ := connection.Get(10)

	fmt.Println(fmt.Sprintf("%f", msg[0].Timestamp))
}
