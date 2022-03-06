package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	mylog "github.com/Breeze0806/go/log"
)

//go:generate go run main.go
var log mylog.Logger = mylog.NewDefaultLogger(os.Stdout, mylog.ErrorLevel, "[datax]")

var readerCode = `package %v

import (
	"github.com/Breeze0806/go-etl/config"
	"github.com/Breeze0806/go-etl/datax/plugin/reader"
)

var _pluginConfig string

func init() {
	var err error
	maker := &maker{}
	if _pluginConfig, err = reader.RegisterReader(maker); err != nil {
		panic(err)
	}
}

var pluginConfig = %v

//NewReaderFromFile 创建读取器
func NewReaderFromFile(filename string) (rd reader.Reader, err error) {
	r := &Reader{}
	r.pluginConf, err = config.NewJSONFromFile(filename)
	if err != nil {
		return nil, err
	}
	rd = r
	return
}

//NewReaderFromString 创建读取器
func NewReaderFromString(filename string) (rd reader.Reader, err error) {
	r := &Reader{}
	r.pluginConf, err = config.NewJSONFromString(filename)
	if err != nil {
		return nil, err
	}
	rd = r
	return
}

type maker struct{}

func (m *maker) FromFile(filename string) (reader.Reader, error) {
	return NewReaderFromFile(filename)
}

func (m *maker) Default() (reader.Reader, error) {
	return NewReaderFromString(pluginConfig)
}
`

var writerCode = `package %v

import (
	"github.com/Breeze0806/go-etl/config"
	"github.com/Breeze0806/go-etl/datax/plugin/writer"
)

var _pluginConfig string

func init() {
	var err error
	maker := &maker{}
	if _pluginConfig, err = writer.RegisterWriter(maker); err != nil {
		panic(err)
	}
}

var pluginConfig = %v

//NewWriterFromFile 创建写入器
func NewWriterFromFile(filename string) (wr writer.Writer, err error) {
	w := &Writer{}
	w.pluginConf, err = config.NewJSONFromFile(filename)
	if err != nil {
		return nil, err
	}
	wr = w
	return
}

//NewWriterFromString 创建写入器
func NewWriterFromString(filename string) (wr writer.Writer, err error) {
	w := &Writer{}
	w.pluginConf, err = config.NewJSONFromString(filename)
	if err != nil {
		return nil, err
	}
	wr = w
	return
}

type maker struct{}

func (m *maker) FromFile(filename string) (writer.Writer, error) {
	return NewWriterFromFile(filename)
}

func (m *maker) Default() (writer.Writer, error) {
	return NewWriterFromString(pluginConfig)
}`

func main() {
	var imports []string
	parser := pluginParser{}
	if err := parser.readPackages("../reader"); err != nil {
		log.Errorf("%v", err)
		return
	}
	for _, info := range parser.infos {
		if err := info.genFile("../reader", readerCode); err != nil {
			log.Errorf("%v", err)
			return
		}
		imports = append(imports, info.genImport("reader"))
	}

	imports = append(imports, "")
	parser.infos = nil
	if err := parser.readPackages("../writer"); err != nil {
		log.Errorf("%v", err)
		return
	}
	for _, info := range parser.infos {
		if err := info.genFile("../writer", writerCode); err != nil {
			log.Errorf("%v", err)
			return
		}
		imports = append(imports, info.genImport("writer"))
	}

	f, err := os.Create("../plugin.go")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()
	f.WriteString(`package plugin

import (
`)
	for _, v := range imports {
		f.WriteString(v)
		f.WriteString("\n")
	}
	f.WriteString(")\n")
	return
}

type pluginParser struct {
	infos []pluginInfo
}

func (p *pluginParser) readPackages(path string) (err error) {
	var list []os.FileInfo
	list, err = ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	var data []byte
	for _, v := range list {
		if v.IsDir() {
			data, err = ioutil.ReadFile(filepath.Join(path, v.Name(), "resources", "plugin.json"))
			if err != nil {
				err = nil
				continue
			}
			info := pluginInfo{
				shotPackage:  v.Name(),
				pluginConfig: "`" + string(data) + "`",
			}
			p.infos = append(p.infos, info)
		}
	}
	return
}

type pluginInfo struct {
	shotPackage  string
	pluginConfig string
}

func (p *pluginInfo) genImport(typ string) string {
	return fmt.Sprintf(`	_ "github.com/Breeze0806/go-etl/datax/plugin/%s/%s"`, typ, p.shotPackage)
}

func (p *pluginInfo) genFile(path string, code string) (err error) {
	var f *os.File
	f, err = os.Create(filepath.Join(path, p.shotPackage, "plugin.go"))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, code, p.shotPackage, p.pluginConfig)
	return
}