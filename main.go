package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/opensourceways/robot-gitee-plugin-lib/logrusutil"
	liboptions "github.com/opensourceways/robot-gitee-plugin-lib/options"
	"github.com/opensourceways/robot-gitee-plugin-lib/plugin"
	"github.com/opensourceways/sync-file-server/grpc/client"
	"github.com/sirupsen/logrus"
)

type options struct {
	plugin   liboptions.PluginOptions
	endpoint string
}

func (o *options) Validate() error {
	if o.endpoint == "" {
		return fmt.Errorf("the endpoint parameter can not be empty")
	}

	return o.plugin.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.plugin.AddFlags(fs)
	fs.StringVar(&o.endpoint, "endpoint", "", "The one of synchronizing file server which is a grpc server")

	fs.Parse(args)
	return o
}

func main() {
	logrusutil.ComponentInit(pluginName)

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	rpcCli, err := client.NewClient(o.endpoint)
	if err != nil {
		logrus.WithError(err).Fatal()
	}

	p := newSyncFile(rpcCli, nil)

	plugin.Run(p, o.plugin)
}
