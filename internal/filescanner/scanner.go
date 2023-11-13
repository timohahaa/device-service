package filescanner

import (
	"fmt"
	"github.com/timohahaa/postgres"
	"os"
	"strings"
	"time"
)

type Scanner struct {
	db           *postgres.Postgres
	scanInterval time.Duration
	inputDir     string
	outputDir    string
}

func NewScanner(pg *postgres.Postgres, interval time.Duration, inputDir, outputDir string) *Scanner {
	s := &Scanner{
		db:           pg,
		scanInterval: interval,
		inputDir:     inputDir,
		outputDir:    outputDir,
	}
	return s
}

func (s *Scanner) Start(stopChan chan struct{}) {
	ticker := time.NewTicker(s.scanInterval)
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
}

func (s *Scanner) rescanFileSystem() {
	/*
						есть два пути:

					    1) достать из БД ВСЕ уже обработанные файлы и сохранить их в хэш-мапу
					    считать директорию - если файла нет в мапе - обработать его
						очевидная слабая сторона этого метода - а если файлов в БД ОЧЕНЬ МНОГО - приложение будет тратить много памяти
						вариант решения - хранить обработанные файлы в кэше по типу редиса, который переодически подчищается

						2) проверять каждый файл ПОШТУЧНО
				        то есть считали дирректорию - выбрали файл - проверили, обрабатывали ли мы его до этого (сделали запрос в базу данных)
				        если нет - то обрабатываем
				        слабые стороны: много файлов => много запросов к базе => снижается производительность приложения
				        вариант решения - если база поддерживает достаточно много параллельных подключений,
		                можно сильно распараллелить обработку файлов

				        КАКОЙ ВАРИАНТ ВЫБРАЛ Я ЗДЕСЬ
				        наверное все таки тестовое задание не предполагает оверинжиниринг и решение подобных (не очень простых) архитектурных вопросов
				        поэтому я решил не усложнять и предположил следующие вещи:
				            - "входная" директория переодически отчищается от старых файлов, и особо много файлов там лежать одновременно не будет
				            - для этого сервиса важна не скорость работы, а ее безошибочность,
				              так что можно немножко пожертвовать производительностью
				            - желательно использовать реляционную БД (ну, мало ли почему нет доступа к редису)
				        => с каждым файлом идем и делаем запрос к БД, ПОКА ЧТО не занимаемся супер-распараллеливанием
				        (это всегда можно сделать в дальнейшем, если нас не устроят бенчмарки)
	*/
	// считываем директорию
	// можно игнорировать ошибку, так как позже произойдет рескан
	entries, _ := os.ReadDir(s.inputDir)

	// проходимся по результатам, если видим файл - к тому же .tsv файл, обрабатываем его
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// проверим на .tsv файл
		filename := entry.Name()
		if strings.HasSuffix(filename, ".tsv") {
			s.parseFile(filename)
		}
	}
}

func (s *Scanner) parseFile(filename string) {
	// как написал выше - сначала лезем в базу, проверяем, был ли там такой файл
	// ошибку игнорируем, потому что запрос составляется не динамически
	sql, args, _ := s.db.Builder.Select("*").From("files").Where("filename = ?", filename).ToSql()
}

func (s *Scanner) Test() {
	s.rescanFileSystem()
}
