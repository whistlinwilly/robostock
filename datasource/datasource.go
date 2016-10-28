package datasource

import (
	"io/ioutil"
	"os"
	"syscall"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/uber-go/zap"
)

const SYMBOL_FILE string = "/tmp/symbols.txt"

type Datasource struct {
	logger zap.Logger
}

func New(logger zap.Logger) *Datasource {
	if !HaveSymbols() {
		if err := FetchSymbols(); err != nil {
			logger.Panic(err.Error())
		}
	}
	logger.Info("GOT SYMBOLS")
	return &Datasource{}
}

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
