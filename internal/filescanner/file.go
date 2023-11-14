package filescanner

import (
	"fmt"
	"reflect"
	"strconv"
)

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
	Context   string
	Class     MsgClass
	Level     int
	Area      string
	Addr      string
	Block     bool // подразумеваю, что в файле 0 - не использовать, 1 - использовать
	Type      string
	Bit       int
	InvertBit bool // подразумеваю, что в файле 0 - не инвертировать, 1 - инвертировать
}

type File struct {
	Rows []FileRow
}

func FileRowFromRecord(record []string) (FileRow, error) {
	// написал эту функцию на reflect-ах, сейчас полей в файле 14, а завтра будет 20, послезавтра 30, что делать?
	// еще и поля могут иметь разные типы данных!
	// лучше написать одну такую "нечитаемую" функцию, зато потом будет чуть легче поддерживать код...
	row := FileRow{}
	ptr := reflect.ValueOf(&row)
	s := ptr.Elem()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		switch f.Type().String() {
		case "string":
			f.SetString(record[i])
		case "int":
			intValue, err := strconv.ParseInt(record[i], 10, 0)
			if err != nil {
				fmt.Println(i)
				return row, err
			}
			f.SetInt(intValue)
		case "bool":
			boolValue, err := strconv.ParseBool(record[i])
			if err != nil {
				return row, err
			}
			f.SetBool(boolValue)
		case "filescanner.MsgClass":
			f.SetString(record[i])
		default:
			// just chill
		}
	}
	return row, nil
}
