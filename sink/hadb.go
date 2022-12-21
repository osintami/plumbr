// © 2022 Sloan Childers
package sink

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

type HashDBReader struct {
	fileName string
	fh       *os.File
	fps      map[string]*HashDBEntry
}

type HashDBEntry struct {
	filePtr int64
	rowLen  int64
}

var ErrEntryNotFound = errors.New("row not found")

// create an HADB reader object
func NewHADBReader(fileName string) (*HashDBReader, error) {
	x := &HashDBReader{
		fileName: fileName,
		fh:       nil,
		fps:      make(map[string]*HashDBEntry)}

	var err error
	x.fh, err = os.Open(x.fileName)
	if err != nil {
		return x, err
	}

	return x.IndexFile()
}

func (x HashDBReader) IndexFile() (*HashDBReader, error) {
	// iterate over each line in the file to build the indexes
	scanner := bufio.NewScanner(x.fh)
	scanner.Split(bufio.ScanLines)
	var filePtr int64 = 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		result := gjson.Get(line, "Key")
		if !result.Exists() || result.Str == "" {
			continue
		}
		key := strings.Trim(result.Str, "\"")

		ce := &HashDBEntry{
			filePtr: filePtr,
			rowLen:  int64(len(line) + 1)}

		x.fps[key] = ce
		filePtr += ce.rowLen
	}

	if scanner.Err() != nil {
		return &x, scanner.Err()
	}
	return &x, nil
}

func (x HashDBReader) SaveIndex(fileName string) error {
	fh, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer fh.Close()
	for k, v := range x.fps {
		fh.WriteString(fmt.Sprintf("%s,%d,%d", k, v.filePtr, v.rowLen))
	}
	return nil
}

func (x HashDBReader) Find(key string, column string) (*gjson.Result, error) {
	ce := x.fps[key]
	if ce == nil {
		return &gjson.Result{}, ErrEntryNotFound
	}
	row := make([]byte, ce.rowLen-1)

	_, err := x.fh.ReadAt(row, ce.filePtr)
	if err != nil {
		return nil, err
	}
	result := gjson.GetBytes(row, column)
	return &result, nil
}

func (x HashDBReader) Lookup(key string, result interface{}) (bool, error) {
	ce := x.fps[key]
	if ce == nil {
		return false, nil
	}
	row := make([]byte, ce.rowLen-1)

	_, err := x.fh.ReadAt(row, ce.filePtr)
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(row, result)
}

type HashDBWriter struct {
	file string
	fh   *os.File
	fps  map[string]*HashDBEntry
}

func NewHADBWriter(file string) (*HashDBWriter, error) {
	fh, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	return &HashDBWriter{
		file: file,
		fh:   fh,
		fps:  make(map[string]*HashDBEntry)}, nil
}

func (x HashDBWriter) InsertFunc(key string, row json.RawMessage) error {
	out := string(row) + "\n"
	_, err := x.fh.Write([]byte(out))
	if err != nil {
		log.Error().Err(err).Str("component", "hadb").Msg("write")
	}
	return nil
}

func (x HashDBWriter) Close() error {
	return x.fh.Close()
}
