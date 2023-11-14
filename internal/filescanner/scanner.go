package filescanner

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/timohahaa/postgres"
)

type Scanner struct {
	db           *postgres.Postgres
	scanInterval time.Duration
	inputDir     string
	outputDir    string
	filesChan    chan string
}

func NewScanner(pg *postgres.Postgres, interval time.Duration, inputDir, outputDir string) *Scanner {
	s := &Scanner{
		db:           pg,
		scanInterval: interval,
		inputDir:     inputDir,
		outputDir:    outputDir,
		filesChan:    nil,
	}
	return s
}

func (s *Scanner) Start(stopChan chan struct{}) {
	ticker := time.NewTicker(s.scanInterval)
	// создаем очередь под обработку файлов
	// канал закроется при шатдауне сканера
	// буфер в 100 файлов - чтобы не ждать обработки файлов при записи в канал-очередь
	files := make(chan string, 100)
	s.filesChan = files
	// бесконечный цикл пока не дадим сигнал остановки
	go func() {
		for {
			select {
			case <-ticker.C:
				s.rescanFileSystem()
			case <-stopChan:
				return
			}
		}
	}()
	// функция читает из очереди файлы (названия) и обрабатывает их
	go s.parseFiles()

}

func (s *Scanner) rescanFileSystem() {
	/*
	   есть два пути:

	   1) достать из БД ВСЕ уже обработанные файлы и сохранить их в хэш-мапу
	   считать директорию - если файла нет в мапе - обработать его
	   очевидная слабая сторона этого метода - а если файлов в БД ОЧЕНЬ МНОГО - приложение будет тратить много памяти
	   вариант решения - хранить обработанные файлы в кэше по типу редиса, который переодически подчищается

	   2) проверять каждый файл ПОШТУЧНО
	   то есть считали дирректорию - выбрали файл - проверили, обрабатывали ли мы его до этого
	   (сделали запрос в базу данных)
	   если нет - то обрабатываем
	   слабые стороны: много файлов => много запросов к базе => снижается производительность приложения
	   вариант решения - если база поддерживает достаточно много параллельных подключений,
	   можно сильно распараллелить обработку файлов

	   КАКОЙ ВАРИАНТ ВЫБРАЛ Я ЗДЕСЬ
	   наверное все таки тестовое задание не предполагает оверинжиниринг и решение подобных (не очень простых)
	   архитектурных вопросов
	   поэтому я решил не усложнять и предположил следующие вещи:
	       - "входная" директория переодически отчищается от старых файлов,
	          и особо много файлов там лежать одновременно не будет
	       - для этого сервиса важна не скорость работы, а ее безошибочность,
	         так что можно немножко пожертвовать производительностью
	       - желательно использовать реляционную БД (ну, мало ли почему нет доступа к редису)
	   => с каждым файлом идем и делаем запрос к БД, ПОКА ЧТО не занимаемся супер-распараллеливанием
	   (это всегда можно сделать в дальнейшем, если нас не устроят бенчмарки)
	*/

	// считываем директорию
	// можно игнорировать ошибку, так как позже произойдет рескан
	entries, _ := os.ReadDir(s.inputDir)

	// проходимся по результатам, если видим файл - к тому же .tsv файл, добавляем его в очередь на обработку
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// проверим на .tsv файл
		filename := entry.Name()
		if strings.HasSuffix(filename, ".tsv") {
			// добавляем файл в очередь
			s.filesChan <- filename
		}
	}
}

func (s *Scanner) parseSingleFile(filename string) (*File, error) {
	f, err := os.Open(s.inputDir + filename)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(f)
	// у нас .tsv файл, поэтому меняем separator
	r.Comma = '\t'

	file := &File{}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			// возникла ошибка парсинга
			return nil, err
		}

		row, err := FileRowFromRecord(record)
		if err != nil {
			return nil, err
		}
		fmt.Println(row)
		file.Rows = append(file.Rows, row)
	}
	return file, nil
}

func (s *Scanner) parseFiles() {
	// канал закроется при шатдауне сканера
	// КАКОЙ ЕЩЕ ЕСТЬ ВАРИАНТ: читаем имя файла из канала и обрабатываем каждый файл в отдельной горутине,
	// НО это может сильно упереться в количество подключений к базе данных (размер пула),
	// возможно в таком случае стоит протестировать оба варианта по скорости и выбрать подходящий
	// пока что так - есть очередь, из очереди последовательно достаем и обрабатываем файлы один за другим
	for filename := range s.filesChan {
		// как написал выше - сначала лезем в базу, проверяем, был ли там такой файл
		// ошибку игнорируем, потому что запрос составляется не динамически
		sql, args, _ := s.db.Builder.Select("COUNT(*)").From("files").Where("filename = ?", filename).ToSql()

		var count int
		err := s.db.ConnPool.QueryRow(context.Background(), sql, args...).Scan(&count)
		if err != nil {
			// ошибку с базы данных логируем, так как это не ошибка парсинга файла
			slog.Error("error while searching file in db", filename, err)
		}
		// если такого файла нет, обрабатываем его и потенциальную ошибку записываем в базу данных
		if count == 0 {
			file, err := s.parseSingleFile(filename)
			if err != nil {
				// есть ошибка, пишем в базу
				sql, args, _ := s.db.Builder.Insert("files").Columns("filename", "error").Values(filename, err.Error()).ToSql()
				_, err = s.db.ConnPool.Exec(context.Background(), sql, args...)
				if err != nil {
					slog.Error("error while inserting file error in db", "filename", filename, "err", err)
				}
			}
			// ошибки нет - пишем в базу данные о файле
			for _, row := range file.Rows {
				sql, args, _ := s.db.Builder.Insert("devices").Columns(
					"mqtt",
					"invid",
					"unit_guid",
					"msg_id",
					"text",
					"context",
					"class",
					"level",
					"area",
					"addr",
					"block",
					"type",
					"bit",
					"invert_bit",
				).Values(
					row.Mqtt,
					row.Invid,
					row.UnitGuid,
					row.MsgId,
					row.Text,
					row.Context,
					row.Class,
					row.Level,
					row.Area,
					row.Addr,
					row.Block,
					row.Type,
					row.Bit,
					row.InvertBit,
				).ToSql()
				_, err = s.db.ConnPool.Exec(context.Background(), sql, args...)
				if err != nil {
					slog.Error("error while inserting file rows in db", "filename", filename, "err", err)
				}
			}
			// пишем инфу о файле
			sql, args, _ := s.db.Builder.Insert("files").Columns("filename").Values(filename).ToSql()
			_, err = s.db.ConnPool.Exec(context.Background(), sql, args...)
			if err != nil {
				slog.Error("error while inserting file info in db", "filename", filename, "err", err)
			}
		}
	}
}

func (s *Scanner) Test() {
	files := make(chan string, 100)
	s.filesChan = files
	go s.rescanFileSystem()
	s.parseFiles()
}
