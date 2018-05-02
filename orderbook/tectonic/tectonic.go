package tectonic

import (
	"bytes"
	"fmt"
	"net"
	"strconv"

	"github.com/pquerna/ffjson/ffjson"
	"gitlab.com/CuteQ/roadkill/orderbook"
)

// Tick : Contains data that was loaded in from the datastore
type Tick struct {
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
// Insert(t TTick)								error					done
// InsertInto(dbName string, t TTick)			error					done
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
// Connect()									error					done
// SendMessage()								( string, error )		done
//
// DeltaToTick(delta orderbook.Delta)			Tick			done
// DeltaBatchToTick(deltas []orderbook.Delta)	[]Tick			done
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
func (t *Tectonic) BulkAdd(ticks []Tick) error {
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
func (t *Tectonic) BulkAddInto(dbName string, ticks []Tick) error {
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
func (t *Tectonic) Get(amount uint64) ([]Tick, error) {
	// We use a buffer here to make it easier to maintain
	var (
		msgBuf  = bytes.Buffer{}
		msgJSON = []Tick{}
	)
	msgBuf.WriteString("GET ")
	msgBuf.WriteString(strconv.Itoa(int(amount)))
	msgBuf.WriteString(" AS JSON")

	msgRecv, recvErr := t.SendMessage(msgBuf.String())
	ffjson.Unmarshal(bytes.Trim([]byte(msgRecv[9:]), "\x00"), &msgJSON) // We get back a message starting with `\uFFFE` - Trim that and all null chars in array

	return msgJSON, recvErr
}

// GetFrom : Returns items from specified store
func (t *Tectonic) GetFrom(dbName string, amount uint64, asTick bool) ([]Tick, error) {
	// We use a buffer here to make it easier to maintain
	var (
		msgBuf  = bytes.Buffer{}
		msgJSON = []Tick{}
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

// Insert : Inserts a single tick into the currently selected datastore
func (t *Tectonic) Insert(tick Tick) error {
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
	tickString := fmt.Sprintf("%.3f, %d, %s, %s, %f, %f;", tick.Timestamp, tick.Seq, isTrade, isBid, tick.Price, tick.Size)

	_, err := t.SendMessage("INSERT " + tickString)

	return err
}

// InsertInto : Inserts a single tick into the datastore specified by `dbName`
func (t *Tectonic) InsertInto(dbName string, tick Tick) error {
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
	tickString := fmt.Sprintf("%.3f, %d, %s, %s, %f, %f;", tick.Timestamp, tick.Seq, isTrade, isBid, tick.Price, tick.Size)

	_, err := t.SendMessage("INSERT " + tickString + " INTO " + dbName)

	return err
}

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

// TODO: Make these methods functional
//func (t *Tectonic) Subscribe(dbName, message chan string) (string, error) {
//
//}
//
//func (t *Tectonic) Unsubscribe() (string, error)    {
//
//}

// DeltaToTick : Converts single `orderbook.Delta` into `Tick`` format
func DeltaToTick(delta orderbook.Delta) Tick {
	return Tick{
		Timestamp: float64(delta.Timestamp) * 1e-6,
		Seq:       uint64(delta.Seq),
		IsTrade:   (orderbook.IsTrade &^ delta.Event) == 0,
		IsBid:     (orderbook.IsBid &^ delta.Event) == 0,
		Price:     delta.Price,
		Size:      delta.Size,
	}
}

// DeltaBatchToTick : converts `[]orderbook.Delta` into `[]Tick`
func DeltaBatchToTick(deltas []orderbook.Delta) []Tick {
	tickBatch := make([]Tick, len(deltas))
	for i, delta := range deltas {
		tickBatch[i] = Tick{
			Timestamp: float64(delta.Timestamp) * 1e-6,
			Seq:       uint64(delta.Seq),
			IsTrade:   (orderbook.IsTrade &^ delta.Event) == 0,
			IsBid:     (orderbook.IsBid &^ delta.Event) == 0,
			Price:     delta.Price,
			Size:      delta.Size,
		}
	}

	return tickBatch
}

// Exists : Checks if datastore exists
func (t *Tectonic) Exists(dbName string) bool {
	msg, _ := t.SendMessage("EXISTS " + dbName)

	// EXISTS command returns `1` for an existing datastore, and `ERR:...` otherwise
	return msg[0] == 1
}
