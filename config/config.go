package config

import (
	"encoding/xml"
	"io/ioutil"
	"os"

	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/taskworker"
)

type Neighbour struct {
	Name      string `xml:"name,attr"`
	IP        string `xml:"ip,attr"`
	Port      uint32 `xml:"port,attr"`
	AgentPort uint32 `xml:"agent_port,attr"`
}

type NodeConfig struct {
	Name                   string      `xml:"name"`
	Port                   uint32      `xml:"port"`
	Neighbours             []Neighbour `xml:"neighbours>info"`
	NeighbourMaxConnection int         `xml:"neighbour_max_connection"`
}

type AgentConfig struct {
	Name                string `xml:"name"`
	Port                uint32 `xml:"port"`
	ClientMaxConnection int    `xml:"client_max_connection"`
}

type Config struct {
	XMLName   xml.Name `xml:"server"`
	LogLevel  string   `xml:"log_level"`
	LogMode   string   `xml:"log_mode"`
	LogPath   string   `xml:"log_path"`
	SecretKey string   `xml:"secret_key"`
	taskworker.WorkerConfigs

	Host                 string      `xml:"host"`
	HostEnv              string      `xml:"host_env"`
	NodeConf             NodeConfig  `xml:"node"`
	AgentConf            AgentConfig `xml:"agent"`
	TimeoutCheckInterval int64       `xml:"timeout_check_interval"`
	ConnectTimeout       int64       `xml:"connect_timeout"`
	SignTimeout          int64       `xml:"sign_timeout"`
	ClientTimeout        int64       `xml:"client_timeout"`

	StopTimeout int64 `xml:"stop_timeout"`
}

var ServerConf Config

func init() {
	dirs := []string{"./", "./conf/", "../conf/", "../../conf/"}
	for _, dir := range dirs {
		xmlFile, err := os.Open(dir + "server.xml")
		if err == nil {
			defer xmlFile.Close()

			b, _ := ioutil.ReadAll(xmlFile)
			err = xml.Unmarshal(b, &ServerConf)
			if err != nil {
				log.LogF("panic", "load config file: %sserver.xml %s", dir, err)
			}
			log.LogF("info", "load config file %sserver.xml\n %#v", dir, ServerConf)
		}
	}

	if len(ServerConf.Host) == 0 {
		ServerConf.Host = os.Getenv(ServerConf.HostEnv)
	}
	log.LogF("info", "load config: %#v", ServerConf)
}
