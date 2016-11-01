package datasource

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/uber-go/zap"
)

const SYMBOL_FILE string = "/tmp/symbols.txt"
const DATASOURCE_URL string = "https://www.google.com/finance/historical?output=csv&q="

var VALID_SYMBOL = regexp.MustCompile(`^[A-Z]+|`)

func FetchSymbols() error {
	c, err := ftp.Connect("ftp.nasdaqtrader.com:21")
	if err != nil {
		return err
	}
	err = c.Login("anonymous", "anonymous")
	if err != nil {
		return err
	}
	remoteFile, err := c.Retr("./SymbolDirectory/nasdaqlisted.txt")
	if err != nil {
		return err
	}
	defer remoteFile.Close()
	b, err := ioutil.ReadAll(remoteFile)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(SYMBOL_FILE, b, os.FileMode(int(0664)))
	if err != nil {
		return err
	}
	return nil
}

func HaveSymbols() bool {
	f, err := os.Stat(SYMBOL_FILE)
	if err != nil {
		// Symbols file is missing
		return false
	}
	stat := f.Sys().(*syscall.Stat_t)
	ctime := time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	if time.Since(ctime).Hours() > 12 {
		// Symbols file is old
		return false
	}
	return true
}

type Datasource struct {
	logger  zap.Logger
	r       *rand.Rand
	sampler Sampler
}

func New(logger zap.Logger, sampleSize int) *Datasource {
	if !HaveSymbols() {
		if err := FetchSymbols(); err != nil {
			logger.Panic(err.Error())
		}
	}
	return &Datasource{
		logger:  logger,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		sampler: NewFibonacciSampler(logger, sampleSize),
	}
}

type Sampler interface {
	Read(r io.ReadCloser) ([]float64, error)
}

func NewFibonacciSampler(logger zap.Logger, sampleSize int) Sampler {
	return &FibonacciSampler{logger: logger, sampleSize: sampleSize}
}

type FibonacciSampler struct {
	logger     zap.Logger
	sampleSize int
}

func (fs *FibonacciSampler) Read(r io.ReadCloser) ([]float64, error) {
	data := make([]float64, fs.sampleSize)
	fibA := 0
	fibB := 1
	reader := bufio.NewReader(r)
	reader.ReadLine() // Column Headers
	line, _, err := reader.ReadLine()
	if err != nil {
		return []float64{}, err
	}
	if f, err := strconv.ParseFloat(strings.Split(string(line), ",")[4], 64); err == nil {
		data[0] = f
	} else {
		return []float64{}, err
	}
	for i := 1; i < fs.sampleSize; i++ {
		for j := 0; j < fibA; j++ {
			reader.ReadLine() // Skip fibA lines
		}
		line, _, err := reader.ReadLine()
		if err != nil {
			return []float64{}, err
		}
		if f, err := strconv.ParseFloat(strings.Split(string(line), ",")[4], 64); err == nil {
			data[i] = f
		} else {
			return []float64{}, err
		}
		tmp := fibB
		fibB = fibA + fibB
		fibA = tmp
	}
	return data, nil
}

func (ds *Datasource) Next() ([]float64, error) {
	var line, l []byte
	f, err := os.Open(SYMBOL_FILE)
	if err != nil {
		ds.logger.Info("Couldn't open symbol file")
		return []float64{}, err
	}
	reader := bufio.NewReader(f)
	linesRead := 1
	for err = nil; err == nil; l, _, err = reader.ReadLine() {
		if ds.r.Intn(linesRead) == 0 {
			if VALID_SYMBOL.Match(l) {
				l = VALID_SYMBOL.Find(l)
				if l != nil {
					line = make([]byte, len(l))
					if copy(line, l) != len(l) {
						ds.logger.Info("Bad copy")
					}
				}
			}
		}
		linesRead += 1
	}
	if err != io.EOF {
		ds.logger.Info("what", zap.Error(err))
		return []float64{}, err
	}
	ds.logger.Info("Fetching data", zap.String("symbol", string(line)))
	return ds.DataFromSymbol(string(line))
}

func (ds *Datasource) DataFromSymbol(symbol string) ([]float64, error) {
	var buffer bytes.Buffer
	buffer.WriteString(DATASOURCE_URL)
	buffer.WriteString(symbol)
	resp, err := http.Get(buffer.String())
	if err != nil {
		ds.logger.Panic("Unable to fetch data from url", zap.Error(err))
	}
	defer resp.Body.Close()
	return ds.sampler.Read(resp.Body)
}
