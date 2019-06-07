package mxt

import (
	"io/ioutil"
	"strconv"

	"github.com/d5/tengo/stdlib"

	"github.com/d5/tengo/script"
	"github.com/gin-gonic/gin"
	"github.com/txn2/ack"
	"github.com/txn2/micro"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// EpConfig
type EpConfig struct {
	Endpoints  map[string]Endpoint  `yaml:"endpoints"`
	Transforms map[string]Transform `yaml:"transforms"`
}

// Transform
type Transform struct {
	Description string `yaml:"description"`
	Script      string `yaml:"script"`
}

// Endpoint
type Endpoint struct {
	Description string `yaml:"description"`
	Location    string `yaml:"location"`
	Transform   string `yaml:"transform"`
}

// CfgFromFile creates a configuration file from YAML
func CfgFromFile(file string) (*EpConfig, error) {
	ymlData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	cfg := &EpConfig{}

	err = yaml.Unmarshal([]byte(ymlData), &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

type ProxyCfg struct {
	EpConfig        *EpConfig
	Logger          *zap.Logger
	HttpClient      *micro.Client
	compiledScripts map[string]*script.Compiled
}

// NewProxy
func NewProxy(cfg *ProxyCfg) (*Proxy, error) {
	// make a default logger if not defined
	if cfg.Logger == nil {
		// logger configuration
		zapCfg := zap.NewProductionConfig()

		logger, err := zapCfg.Build()
		if err != nil {
			return nil, err
		}

		cfg.Logger = logger
	}

	// compile scripts (transforms)
	for name, transform := range cfg.EpConfig.Transforms {
		s := script.New([]byte(transform.Script))
		s.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
		s.EnableFileImport(true)

		// define input by adding zero string, usable
		// as a default string or convert to number eg. int(input)
		err := s.Add("input", "0")
		if err != nil {
			cfg.Logger.Fatal("Could not add input to script.")
		}

		c, err := s.Compile()
		if err != nil {
			return nil, err
		}

		if cfg.compiledScripts == nil {
			cfg.compiledScripts = make(map[string]*script.Compiled)
		}

		err = c.Run()
		if err != nil {
			cfg.Logger.Fatal("Script could not run", zap.String("script", name), zap.Error(err))
		}

		cfg.compiledScripts[name] = c

	}

	// @TODO VALIDATION: validate that transforms defined in each input have
	// corresponding scripts.

	return &Proxy{cfg}, nil
}

// Proxy
type Proxy struct {
	*ProxyCfg
}

// EpHandler is an Endpoint Handler accepting a gin context
// and returns plain text
func (p *Proxy) EpHandler(c *gin.Context) {

	// get endpoint
	ep := c.Param("ep")

	// lookup endpoint
	if epCfg, ok := p.EpConfig.Endpoints[ep]; ok {
		// endpoint was found now go get the data
		resp, err := p.HttpClient.Http.Get(epCfg.Location)
		if err != nil {
			p.Logger.Error("endpoint request error", zap.String("endpoint", ep), zap.Error(err))
			ak := ack.Gin(c)
			ak.GinErrorAbort(500, "E500", err.Error())
			return
		}

		if resp.StatusCode != 200 {
			p.Logger.Error("endpoint response error", zap.String("endpoint", ep), zap.Error(err))
			ak := ack.Gin(c)
			ak.GinErrorAbort(resp.StatusCode, "E"+strconv.Itoa(resp.StatusCode), err.Error())
			return
		}

		input, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			p.Logger.Error("endpoint body read error", zap.String("endpoint", ep), zap.Error(err))
			ak := ack.Gin(c)
			ak.GinErrorAbort(500, "E500", err.Error())
		}

		p.Logger.Info("setting input", zap.ByteString("input", input))
		err = p.compiledScripts[epCfg.Transform].Set("input", string(input))
		if err != nil {
			p.Logger.Error("script add input error",
				zap.String("script", epCfg.Transform),
				zap.ByteString("input", input),
				zap.String("endpoint", ep),
				zap.Error(err),
			)
			ak := ack.Gin(c)
			ak.GinErrorAbort(500, "E500", err.Error())
		}

		err = p.compiledScripts[epCfg.Transform].Run()
		if err != nil {
			p.Logger.Error("error running script",
				zap.String("script", epCfg.Transform),
				zap.ByteString("input", input),
				zap.Error(err),
			)
			ak := ack.Gin(c)
			ak.GinErrorAbort(500, "E500", err.Error())
		}

		// get the output
		c.String(200, p.compiledScripts[epCfg.Transform].Get("output").String())
		return
	}

	ak := ack.Gin(c)
	ak.GinErrorAbort(500, "E500", "Endpoint "+ep+" not found.")
}
