package entity

type MsgClass string

const (
	ClsAlarm   MsgClass = "alarm"
	ClsWarning MsgClass = "warning"
	ClsInfo    MsgClass = "info"
	ClsEvent   MsgClass = "event"
	ClsComand  MsgClass = "comand"
)

type Device struct {
	Mqtt      string   `json:"mqtt"`
	Invid     string   `json:"invid"`
	UnitGuid  string   `json:"unit_guid"`
	MsgId     string   `json:"msg_id"`
	Text      string   `json:"text"`
	Context   string   `json:"context"`
	Class     MsgClass `json:"class"`
	Level     int      `json:"level"`
	Area      string   `json:"area"`
	Addr      string   `json:"addr"`
	Block     bool     `json:"block"` // подразумеваю, что в файле 0 - не использовать, 1 - использовать
	Type      string   `json:"type"`
	Bit       int      `json:"bit"`
	InvertBit bool     `json:"invert_bit"` // подразумеваю, что в файле 0 - не инвертировать, 1 - инвертировать
}
