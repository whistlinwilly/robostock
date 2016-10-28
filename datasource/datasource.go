package datasource

import (
	"bufio"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"syscall"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/uber-go/zap"
)

const SYMBOL_FILE string = "/tmp/symbols.txt"

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
	logger zap.Logger
	r      *rand.Rand
}

func New(logger zap.Logger) *Datasource {
	if !HaveSymbols() {
		if err := FetchSymbols(); err != nil {
			logger.Panic(err.Error())
		}
	}
	return &Datasource{
		logger: logger,
		r:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (ds *Datasource) Next() ([]byte, error) {
	var line, l []byte
	f, err := os.Open(SYMBOL_FILE)
	if err != nil {
		ds.logger.Info("Couldn't open symbol file")
		return []byte{}, err
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
		return []byte{}, err
	}
	return line, nil
}
