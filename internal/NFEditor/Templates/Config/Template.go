package Config

//---***DO NOT TOUCH THIS FILE UNLESS YOU KNOW FOR SURE WHAT YOU ARE DOING***---//
//---***THIS FILE IS MANAGED BY THE EDITOR AND ALTHOUGH IT SHOULD BE OVERWRITTEN IF THINGS WENT WRONG***---//
//---***IT IS NOT RECOMMENDED TO MANUALLY EDIT THIS FILE***---//

import (
	"go.novellaforge.dev/novellaforge/pkg/NFConfig"
	"os"
)

type ProjectConfig struct {
	data   interface{}
	assets interface{}
	config *NFConfig.Config
}

type Location int

const (
	Data Location = iota
	Assets
)

func (p *ProjectConfig) SetAssets(assets interface{}) {
	p.assets = assets
}

func (p *ProjectConfig) SetData(data interface{}) {
	p.data = data
}

func (p *ProjectConfig) GetFile(loc Location, path string) os.File {
	switch loc {
	case Data:

	}
}

func (p *ProjectConfig) SetConfig(config *NFConfig.Config) {
	p.config = config
}

func (p *ProjectConfig) GetConfig() *NFConfig.Config {
	return p.config
}
