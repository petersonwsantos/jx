package cmd

import (
	"fmt"
	"io"

	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/spf13/cobra"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
)

type MetricsOptions struct {
	CommonOptions
	Namespace string
}

var (
	MetricsLong = templates.LongDesc(`
		Gets the metrics of the newest pod for a Deployment.

`)

	MetricsExample = templates.Examples(`
		# displays metrics of the latest pod in deployment myapp
		jx metrics myapp

		# Tails the metrics of the container foo in the latest pod in deployment myapp
		jx metrics myapp -c foo
`)
)

func NewCmdMetrics(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := &MetricsOptions{
		CommonOptions: CommonOptions{
			Factory: f,
			Out:     out,
			Err:     errOut,
		},
	}
	cmd := &cobra.Command{
		Use:     "metrics [deployment]",
		Short:   "Gets the metrics of the latest pod for a deployment",
		Long:    MetricsLong,
		Example: MetricsExample,
		Aliases: []string{"metrics"},
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdutil.CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&options.Namespace, "namespace", "n", "", "the namespace to look for the Deployment. Defaults to the current namespace")
	return cmd
}

func (o *MetricsOptions) Run() error {
	args := o.Args

	client, curNs, err := o.Factory.CreateClient()
	if err != nil {
		return err
	}

	ns := o.Namespace
	if ns == "" {
		ns = curNs
	}
	name := ""
	if len(args) == 0 {
		names, err := GetDeploymentNames(client, ns)
		if err != nil {
			return err
		}
		if len(names) == 0 {
			return fmt.Errorf("There are no Deployments running")
		}
		n, err := util.PickName(names, "Pick Deployment:")
		if err != nil {
			return err
		}
		name = n
	} else {
		name = args[0]
	}

	pod, err := waitForReadyPodForDeployment(client, ns, name)

	if err != nil {
		return err
	}

	if pod == "" {
		return fmt.Errorf("No pod found for namespace %s with name %s", ns, name)
	}

	namespaceVar := "--namespace="+ns

	args = []string{"top", "pod", pod,namespaceVar}

	err = o.runCommand("kubectl", args...)
	if err != nil {
		return err
	}
	return nil

}