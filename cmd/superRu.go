package cmd


import (
    "time"
    "container/list"
    "os"

    tm "github.com/buger/goterm"
    "github.com/spf13/cobra"
    "github.com/qnib/go-dockercli/lib"
)

// watchSrv loops over nodes, services and tasks
var superRu = &cobra.Command{
	Use:   "superRu",
	Short: "Loops over services and their tasks",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
    qd := dockerlib.NewQnibDocker(serviceList, timeout, true)
    cnt := 0
    for {
      cnt += 1
      qd.UpdateNodeList()
      qd.Services, _ = qd.UpdateServiceList()
      qd.SrvTasks = qd.UpdateTaskList()

      qd.Logs = list.New()
      done, rc := qd.CheckRUFinish()
      if done {
        qd.PrintServices()
        qd.PrintLogs()
        qd.PrintEvents()
        if rc == 0 {
          tm.Printf(">>> All Services are updated and healthy -> OK")
        } else if rc == 1 {
          tm.Printf(">>> Some services are faulty (timeout reached and not healthy) -> FAIL")
        }
        tm.Flush()
        os.Exit(rc)
      }
      time.Sleep(time.Duration(loopDelay) * time.Second)
    }
  },
}

func init() {
	RootCmd.AddCommand(superRu)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	superRu.PersistentFlags().IntVar(&loopDelay, "loopDelay", 2, "Loop delay in seconds")
  superRu.PersistentFlags().StringVar(&serviceList, "services", "", "Comma separated list of services to watch")
  superRu.PersistentFlags().IntVar(&timeout, "timeout", 0, "Timeout for a service to become healthy [0: disabled]")


	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rLatestUrlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
