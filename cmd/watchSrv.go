package cmd


import (
    tm "github.com/buger/goterm"
    "strconv"
    "time"
    "github.com/spf13/cobra"

    "github.com/docker/engine-api/client"
    "github.com/docker/engine-api/types"
    "github.com/docker/engine-api/types/filters"
    "golang.org/x/net/context"
)

var loop int
var loopDelay int


// watchSrv loops over nodes, services and tasks
var watchSrv = &cobra.Command{
	Use:   "watchSrv",
	Short: "Loops over services and their tasks",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
    //defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
    cli, err := client.NewEnvClient()
    //NewClient("tcp://wrex2.r4.05.laxa.gaikai.net:2376", "v1.24", nil, defaultHeaders)
    if err != nil {
        panic(err)
    }

    // Nodes
    nodeMap := make(map[string]string)
    nodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
    if err != nil {
        panic(err)
    }
    /*fmt.Println(">>> Nodes")
    nodeForm := "%-35s %-10s %-10s %-7s %-20s\n"
    leader := ""
    fmt.Printf(nodeForm, "Name", "Status", "Roles", "Leader", "Reachability")
    */
    for _,n := range nodes {
      nodeMap[n.ID] = n.Description.Hostname
      /*
      if n.ManagerStatus.Leader {
        leader = "*"
      }
      fmt.Printf(nodeForm, n.Description.Hostname, n.Status.State, n.Spec.Role, leader, n.ManagerStatus.Reachability)
      */
    }

    cnt := 0
    for {
      cnt += 1
      // Services
      services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{})
      if err != nil {
          panic(err)
      }
      if (loop != 1) {
        tm.Clear() // Clear current screen
        // By moving cursor to top-left position we ensure that console output
        // will be overwritten each time, instead of adding new.
        tm.MoveCursor(1, 1)
      }
      tm.Printf(">>> Services \t\t(%s)\n", time.Now().Format(time.RFC3339))
      srvForm := "%-15s %-10s %-30s\n"
      tm.Printf(srvForm, "Name", "Replicas", "Image")
      for _,s := range services {
        replicas := strconv.FormatUint(*s.Spec.Mode.Replicated.Replicas, 10)
        tm.Printf(srvForm, s.Spec.Annotations.Name, replicas, s.Spec.TaskTemplate.ContainerSpec.Image)

        tfilter := filters.NewArgs()
        tfilter.Add("desired-state", "Running")
        tfilter.Add("service", s.Spec.Annotations.Name)
        tasks, err := cli.TaskList(context.Background(), types.TaskListOptions{Filter: tfilter})
        if err != nil {
            panic(err)
        }
        taskForm := "%15s %-10s %-35s %-20s %-20s\n"
        tm.Printf(taskForm, "", "Slot", "Node", "TaskStatus", "Image")
        for _, t := range tasks {
          if (t.Status.State == "rejected") && (t.Status.State == "shutdown" ) {
            continue
            }
          tm.Printf(taskForm, "", strconv.Itoa(t.Slot), nodeMap[t.NodeID], t.Status.State, t.Spec.ContainerSpec.Image)
        }
      }
      if (loop == 0) {
        cnt = -1
      }
      tm.Flush() // Call it every time at the end of rendering
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
  watchSrv.PersistentFlags().IntVar(&loopDelay, "loopDelay",5, "Loop delay in seconds")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rLatestUrlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
