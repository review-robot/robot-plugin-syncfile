package main

import (
	"context"
	"fmt"
	"testing"

	"gitee.com/openeuler/go-gitee/gitee"
	"github.com/opensourceways/robot-gitee-plugin-lib/config"
	"github.com/opensourceways/sync-file-server/protocol"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/sets"
)

type rpcTestClient struct {
}

func (rtc *rpcTestClient) SyncFile(ctx context.Context, in *protocol.SyncFileRequest, opts ...grpc.CallOption) (*protocol.Result, error) {
	excepts := map[string][]string{}
	excepts["hasFile"] = []string{"config/OWNERS", "OWNERS"}
	excepts["noFile"] = []string{}

	v, ok := excepts[in.Branch.Branch]
	if !ok {
		return nil, fmt.Errorf("no matching branch")
	}

	judge := func(a, b []string) bool {
		if (a == nil) != (b == nil) {
			return false
		}

		if len(a) != len(b) {
			return false
		}

		for i := range a {
			bs := sets.NewString(b...)
			if !bs.Has(a[i]) {
				return false
			}
		}

		return true
	}

	if !judge(in.Files, v) {
		return nil, fmt.Errorf("no matching files")
	}

	return nil, nil
}

func (rtc *rpcTestClient) SyncRepoFile(ctx context.Context, in *protocol.SyncRepoFileRequest, opts ...grpc.CallOption) (*protocol.Result, error) {
	return nil, nil
}

func (rtc *rpcTestClient) ListRepos(ctx context.Context, in *protocol.ListRepoRequest, opts ...grpc.CallOption) (*protocol.ListRepoResponse, error) {
	return nil, nil
}

func (rtc *rpcTestClient) ListBranchesOfRepo(ctx context.Context, in *protocol.ListBranchesOfRepoRequest, opts ...grpc.CallOption) (*protocol.ListBranchesOfRepoResponse, error) {
	if in == nil {
		return nil, fmt.Errorf("no input")
	}
	result := &protocol.ListBranchesOfRepoResponse{
		Branches: []*protocol.BranchInfo{
			{Name: "hasFile", Sha: "hjhjksdcbbkcjdskk"}, {Name: "noFile", Sha: "12edewe545345sdfsadf"},
		},
	}
	return result, nil
}

func (rtc *rpcTestClient) Disconnect() error {
	return nil
}

func genPlugin() *syncFile {
	return &syncFile{rpcCli: &rpcTestClient{}}
}

func genEventHasCfgFile() *gitee.PushEvent {
	ref := "ref/head/hasFile"
	return &gitee.PushEvent{
		Repository: &gitee.ProjectHook{
			Namespace: "cve-manage-test",
			Path:      "config",
		},
		Ref: &ref,
		Commits: []gitee.CommitHook{
			{
				Added: []string{
					"config/OWNERS",
					"OWNERS",
					"config/conf.ini",
					"pkg/main.go",
				},
				Modified: []string{
					"OWNERS",
					"pkg/publish.sh",
				},
			},
		},
	}
}

func genEventNoCfgFile() *gitee.PushEvent {
	ref := "ref/head/noFile"
	return &gitee.PushEvent{
		Repository: &gitee.ProjectHook{
			Namespace: "cve-manage-test",
			Path:      "config",
		},
		Ref: &ref,
		Commits: []gitee.CommitHook{
			{
				Added: []string{
					"123.txt",
					"333.txt",
				},
				Modified: []string{
					"666.txt",
				},
			},
		},
	}
}

func genCfg() *configuration {
	return &configuration{
		SyncFile: []pluginConfig{
			{
				PluginForRepo: config.PluginForRepo{
					Repos: []string{"cve-manage-test/config"},
				},
				FileNames: []string{"OWNERS"},
			},
		},
	}
}

func TestSyncFile_handlePushEventHasFile(t *testing.T) {
	plugin := genPlugin()
	err := plugin.handlePushEvent(genEventHasCfgFile(), genCfg(), logrus.WithContext(context.Background()))
	if err != nil {
		t.Error(err)
	}
}

func TestSyncFile_handlePushEventNoFile(t *testing.T) {
	plugin := genPlugin()
	err := plugin.handlePushEvent(genEventNoCfgFile(), genCfg(), logrus.WithContext(context.Background()))
	if err != nil {
		t.Error(err)
	}
}
