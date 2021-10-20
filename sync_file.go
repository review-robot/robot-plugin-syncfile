package main

import (
	"context"
	"errors"
	"path"
	"strings"

	"gitee.com/openeuler/go-gitee/gitee"
	"github.com/opensourceways/robot-gitee-plugin-lib/config"
	"github.com/opensourceways/robot-gitee-plugin-lib/plugin"
	"github.com/opensourceways/sync-file-server/grpc/client"
	"github.com/opensourceways/sync-file-server/protocol"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

const pluginName = "syncfile"

type pushInfo struct {
	owner string
	repo  string
	ref   string
}

type syncFile struct {
	rpcCli client.Client
	clear  func()
}

func newSyncFile(rpcCli client.Client, clear func()) plugin.Plugin {
	return &syncFile{rpcCli: rpcCli, clear: clear}
}

func (sf *syncFile) Exit() {
	_ = sf.rpcCli.Disconnect()
	if sf.clear != nil {
		sf.clear()
	}
}

func (sf *syncFile) NewPluginConfig() config.PluginConfig {
	return &configuration{}
}

func (sf *syncFile) RegisterEventHandler(p plugin.HandlerRegitster) {
	p.RegisterPushEventHandler(sf.handlePushEvent)
}

func (sf *syncFile) handlePushEvent(e *gitee.PushEvent, cfg config.PluginConfig, log *logrus.Entry) error {
	pInfo, err := getRepoInfo(e)
	if err != nil {
		return err
	}

	pCfg, err := sf.pluginConfig(cfg)
	if err != nil {
		return err
	}

	syncFiles := filterSyncFiles(e, pCfg.syncFileFor(pInfo.owner, pInfo.repo))
	if len(syncFiles) == 0 {
		return nil
	}

	bInfo, err := sf.getBranchInfo(pInfo)
	if err != nil {
		return err
	}

	param := &protocol.SyncFileRequest{
		Branch: &protocol.Branch{
			Org:       pInfo.owner,
			Repo:      pInfo.repo,
			Branch:    bInfo.Name,
			BranchSha: bInfo.Sha,
		},
		Files: syncFiles,
	}
	_, err = sf.rpcCli.SyncFile(context.Background(), param)
	return err
}

func (sf *syncFile) getBranchInfo(info pushInfo) (*protocol.BranchInfo, error) {
	param := &protocol.ListBranchesOfRepoRequest{
		Org:  info.owner,
		Repo: info.repo,
	}
	branches, err := sf.rpcCli.ListBranchesOfRepo(context.Background(), param)
	if err != nil {
		return nil, err
	}

	if branches != nil {
		for i := range branches.Branches {
			bif := branches.Branches[i]
			if bif.Name == info.ref {
				return bif, nil
			}
		}
	}

	return nil, errors.New("can't find branch info by call rpc method ")
}

func (sf *syncFile) pluginConfig(cfg config.PluginConfig) (*configuration, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, errors.New("can't convert to configuration")
	}
	return c, nil
}

func getRepoInfo(e *gitee.PushEvent) (pInfo pushInfo, err error) {
	if e == nil || e.Repository == nil {
		err = errors.New("webhook payload data is abnormal")
		return
	}

	getBranch := func(ref *string) string {
		if ref == nil || *ref == "" {
			return ""
		}
		splits := strings.Split(*ref, "/")
		return splits[len(splits)-1]
	}

	pInfo.owner = e.Repository.Namespace
	pInfo.repo = e.Repository.Path

	pInfo.ref = getBranch(e.Ref)
	return
}

func getChangeFiles(e *gitee.PushEvent) []string {
	cfs := sets.NewString()
	for _, v := range e.Commits {
		if len(v.Added) > 0 {
			cfs.Insert(v.Added...)
		}
		if len(v.Modified) > 0 {
			cfs.Insert(v.Modified...)
		}
	}
	return cfs.List()
}

func filterSyncFiles(e *gitee.PushEvent, cfg *pluginConfig) []string {
	var filters []string
	if e == nil || cfg == nil {
		return filters
	}

	changeFiles := getChangeFiles(e)
	fileNames := sets.NewString(cfg.FileNames...)

	for _, v := range changeFiles {
		if fileNames.Has(path.Base(v)) {
			filters = append(filters, v)
		}
	}
	return filters
}
