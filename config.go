package main

import (
	"errors"

	"github.com/opensourceways/robot-gitee-plugin-lib/config"
)

type configuration struct {
	SyncFile []pluginConfig `json:"syncfile" required:"true"`
}

func (c *configuration) Validate() error {
	if c != nil {
		cs := c.SyncFile
		for i := range cs {
			if err := cs[i].validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *configuration) SetDefault() {

}

func (c *configuration) syncFileFor(org, repo string) *pluginConfig {
	if c == nil {
		return nil
	}
	cs := c.SyncFile
	v := make([]config.IPluginForRepo, 0, len(cs))
	for i := range cs {
		v = append(v, &cs[i])
	}
	if i := config.FindConfig(org, repo, v); i >= 0 {
		return &cs[i]
	}
	return nil
}

type pluginConfig struct {
	config.PluginForRepo

	// FileNames is the list of files to be synchronized.
	FileNames []string `json:"file_names" required:"true"`
}

func (pc *pluginConfig) validate() error {
	if len(pc.FileNames) == 0 {
		return errors.New("the file name list cannot be empty")
	}
	return pc.Validate()
}
