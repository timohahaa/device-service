package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/timohahaa/device-service/internal/entity"
	"github.com/timohahaa/postgres"
)

// так как хендлер у приложения довольно простой, а сам эндпоинт всего один
// я решил не делать отдельный dto/repository, а просто передал объект базы данных в хендлер

type v1Handler struct {
	db *postgres.Postgres
}

func NewV1Handler(pg *postgres.Postgres) *v1Handler {
	return &v1Handler{
		db: pg,
	}
}

func (h *v1Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// игнорируем не гет-запросы
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input getInput
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// не указанно одно из необходимых полей для пагинации
	if input.Limit == nil || input.Offset == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	responce, err := h.getResponce(input.UnitGuid, *input.Limit, *input.Offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(responce)
}

type getInput struct {
	Limit    *int   `json:"limit"`
	Offset   *int   `json:"offset"`
	UnitGuid string `json:"unit_guid"`
}

type getOutput struct {
	Info []entity.Device `json:"info"`
}

func (h *v1Handler) getResponce(unitGuid string, limit int, offset int) ([]byte, error) {
	sql, args, _ := h.db.Builder.
		Select("*").
		From("devices").
		Where("unit_guid = ?", unitGuid).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	rows, err := h.db.ConnPool.Query(context.Background(), sql, args...)
	if err != nil {
		return []byte(nil), err
	}

	defer rows.Close()
	devices := []entity.Device{}
	for rows.Next() {
		var device entity.Device
		err := rows.Scan(
			&device.Mqtt,
			&device.Invid,
			&device.UnitGuid,
			&device.MsgId,
			&device.Text,
			&device.Context,
			&device.Class,
			&device.Level,
			&device.Area,
			&device.Addr,
			&device.Block,
			&device.Type,
			&device.Bit,
			&device.InvertBit,
		)
		if err != nil {
			return []byte(nil), err
		}

		devices = append(devices, device)
	}
	output := getOutput{Info: devices}
	data, err := json.Marshal(&output)
	if err != nil {
		return []byte(nil), err
	}

	return data, nil
}
