package config

import (
	"encoding/xml"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/taskworker"
	"io/ioutil"
	"os"
)

type Config struct {
	XMLName	xml.Name `xml:"server"`
	NetWorkers []taskworker.WorkerConfig `xml:"net_workers>worker"`
	Host string `xml:"host"`
	Port int `xml:"port"`
	Delay int `xml:"delay"`
	InitConn int `xml:"init_conn_num"`
	MaxConn int `xml:"max_conn_num"`
	GoodByeTimeoutS int `xml:"good_bye_timeout_s"`
	SecretKey string `xml:"secret_key"`
	HasConfig bool `xml:"-"`
}

var (
	ServerConf Config
)

func init() {
	dirs := []string{"./", "./conf/"} // , "../conf/", "../../conf/"
	for _, dir := range dirs {
		xmlFile, err := os.Open(dir + "server.xml")
		if err == nil {
			defer xmlFile.Close()

			b, _ := ioutil.ReadAll(xmlFile)
			err = xml.Unmarshal(b, &ServerConf)
			if err != nil {
				log.Panicf("load config file: %sserver.xml %s", dir, err)
			}
			ServerConf.HasConfig = true
			log.Infof("load config file %sserver.xml\n %#v", dir, ServerConf)
		}
	}
}
