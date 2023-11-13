package filescanner

type MsgClass string

const (
	ClsAlarm   MsgClass = "alarm"
	ClsWarning MsgClass = "warning"
	ClsInfo    MsgClass = "info"
	ClsEvent   MsgClass = "event"
	ClsComand  MsgClass = "comand"
)

type FileRow struct {
	Mqtt      string
	Invid     string
	UnitGuid  string
	MsgId     string
	Text      string
	Class     MsgClass
	Level     int
	Area      string
	Addr      string
	Block     bool
	Type      string
	Bit       int
	InvertBit bool
}

type File struct {
	Rows []FileRow
}
