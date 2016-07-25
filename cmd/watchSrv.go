package cmd


import (
    "time"
    "container/list"

    tm "github.com/buger/goterm"
    "github.com/spf13/cobra"
    "github.com/qnib/go-dockercli/lib"
)

var loop int
var loopDelay int
var noClear bool

// watchSrv loops over nodes, services and tasks
var watchSrv = &cobra.Command{
	Use:   "watchSrv",
	Short: "Loops over services and their tasks",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
    qd := dockerlib.NewQnibDocker()
    cnt := 0
    for {
      cnt += 1
      if loop != 1 || noClear {
        tm.Clear() // Clear current screen
        tm.MoveCursor(1, 1)
      }
      tm.Printf(">> Loop %ds\t\t(%s)\n", loopDelay, time.Now())
      qd.UpdateNodeList()
      qd.Services, _ = qd.UpdateServiceList()
      qd.SrvTasks = qd.UpdateTaskList()

      qd.PrintServices()
      qd.PrintLogs()
      qd.PrintEvents()
      qd.Logs = list.New()
      tm.Flush()
      if cnt == loop {
        break
      }
      time.Sleep(time.Duration(loopDelay) * time.Second)
    }
  },
}

func init() {
	RootCmd.AddCommand(watchSrv)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	watchSrv.PersistentFlags().IntVar(&loop, "loop", 1, "Loop command [0: infinite]")
  watchSrv.PersistentFlags().IntVar(&loopDelay, "loopDelay", 2, "Loop delay in seconds")
  watchSrv.PersistentFlags().BoolVar(&noClear, "no-clear", false, "Do not clear the screen for each loop (implicit when loop==1)")


	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rLatestUrlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
