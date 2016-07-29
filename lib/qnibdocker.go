package dockerlib

import (
  "fmt"
  "strconv"
  "container/list"
  "time"
  "os"
  "regexp"

  tm "github.com/buger/goterm"
  "github.com/docker/engine-api/client"
  "github.com/docker/engine-api/types"
  "github.com/docker/engine-api/types/filters"
  "github.com/docker/engine-api/types/swarm"
  "golang.org/x/net/context"
)


//var imageExp = regexp.MustCompile(`(?P<registry>[a-zA-Z0-9\.\-]+)/(?P<registry>[a-zA-Z0-9\.\-]+)`)

type QnibDocker struct {
  DockerCli     *client.Client
  ServiceList   string
  ServiceTimeout int
  Services      []swarm.Service
  SrvTasks      map[string][]TaskConf
  NodeMap       map[string]string
  RuStart       bool
  CurrentConf   map[string]StackConf
  NextConf      map[string]StackConf
  Logs          *list.List // flushed at each loop iteration
  Events        *list.List // not flushed, therefore kept while looping through
}

func NewQnibDocker(serviceList string, timeout int) (QnibDocker) {
  cli, err := client.NewEnvClient()
  if err != nil {
    panic(err)
  }
  qd := QnibDocker{
    DockerCli: cli,
    ServiceTimeout: timeout,
    RuStart: false,
    ServiceList: serviceList,
    NodeMap: make(map[string]string),
    CurrentConf: make(map[string]StackConf),
    NextConf: make(map[string]StackConf),
    Logs:   list.New(),
    Events:   list.New(),
   }
  return qd
}

func (qd QnibDocker) AddEvent(event string) (error) {
  now := time.Now().Format(time.RFC3339)
  qd.Events.PushBack(fmt.Sprintf("%s > %s", now, event))
  return nil
}

func (qd QnibDocker) PrintEvents() (error) {
  if qd.Events.Len() > 0 {
    tm.Printf("\n\n>> Events\n")
  }
  for e := qd.Events.Front(); e != nil; e = e.Next() {
		tm.Println(e.Value)
	}
  return nil
}

func (qd QnibDocker) AddLog(log string) (error) {
  now := time.Now().Format(time.RFC3339)
  qd.Logs.PushBack(fmt.Sprintf("%s > %s", now, log))
  return nil
}

func (qd QnibDocker) PrintLogs() (error) {
  if qd.Logs.Len() > 0 {
    tm.Printf("\n\n>> Logs within Loop (flushed afterwards)\n")
  }
  for e := qd.Logs.Front(); e != nil; e = e.Next() {
		tm.Println(e.Value)
	}
  return nil
}

func (qd QnibDocker) UpdateNodeList() (error) {
  nodes, err := qd.DockerCli.NodeList(context.Background(), types.NodeListOptions{})
  if err != nil {
      panic(err)
  }
  for _,n := range nodes {
    qd.NodeMap[n.ID] = n.Description.Hostname
  }
  return nil
}

func (qd QnibDocker) UpdateServiceList() ([]swarm.Service, error) {
  services, err := qd.DockerCli.ServiceList(context.Background(), types.ServiceListOptions{})
  if err != nil {
      return nil, err
  }
  for _,s := range services {
    replicas := int(*s.Spec.Mode.Replicated.Replicas)
    srvName := s.Spec.Annotations.Name
    srvImage := s.Spec.TaskTemplate.ContainerSpec.Image
    sc := NewStackConf(srvImage, replicas)
    qd.UpdateServiceConf(srvName, sc)
  }
  return services, nil
}

func (qd QnibDocker) PrintServices() {
  srvForm := " %-15s %-10s %-40s %-40s\n"
  tm.Printf(srvForm, "Name", "Replicas", "Image", "Tag")
  for _,s := range qd.Services {
    replicas := int(*s.Spec.Mode.Replicated.Replicas)
    srvName := s.Spec.Annotations.Name
    srvImage := s.Spec.TaskTemplate.ContainerSpec.Image
    ic := NewImageConf(srvImage)
    tm.Printf(srvForm, srvName, strconv.Itoa(replicas), ic.PrintImage(), ic.PrintTag())
    qd.PrintTasks(srvName)
  }
}

func (qd QnibDocker) UpdateServiceConf(srvName string, sc StackConf) {
  _, srv := qd.CurrentConf[srvName]
  if ! srv {
    qd.CurrentConf[srvName] = sc
  } else {
    qd.AddLog(fmt.Sprintf("Service '%s' already in CurrentConf", srvName))
  }
}

func (qd QnibDocker) IsRuFinished() (bool) {
  tm.Println("Don't think so...")
  return false
}

func (qd QnibDocker) UpdateTaskList() (map[string][]TaskConf) {
  qt := make(map[string][]TaskConf)
  for _,s := range qd.Services {
    //replicas := int(*s.Spec.Mode.Replicated.Replicas)
    srvName := s.Spec.Annotations.Name
    tm.Printf(">>>>> %s %s\n", srvName, qd.ServiceList)
    //srvImage := s.Spec.TaskTemplate.ContainerSpec.Image
    //ic := NewImageConf(srvImage)
    tfilter := filters.NewArgs()
    tfilter.Add("desired-state", "Running")
    tfilter.Add("service", srvName)
    tasks, err := qd.DockerCli.TaskList(context.Background(), types.TaskListOptions{Filter: tfilter})
    if err != nil {
        panic(err)
    }
    _, srv := qt[srvName]
    if ! srv {
      qt[srvName] = []TaskConf{}
    }
    for _, t := range tasks {
      tic := NewImageConf(t.Spec.ContainerSpec.Image)
      nTask := NewTaskConf(t, tic)
      qt[srvName] = append(qt[srvName], nTask)
    }

  }
  //tm.Println(qt)
  return qt
}

func (qd QnibDocker) PrintTasks(srv string) (error) {
  taskForm := "%-27s %-7s %-25s %-10s %-10s %-15s %-35s %-35s\n"
  tm.Printf(taskForm, "ID", "Slot", "Node", "TaskState", "SecSince", "CntStatus", "Image", "DesiredImage")
  for _, t := range qd.SrvTasks[srv] {
      cStatus, cElapse, faulty := qd.CheckTaskHealth(t)
      if faulty {
        cStatus = "FAULTY"
        qd.AddLog(fmt.Sprintf("Slot %d became faulty", t.Slot))
        tm.Flush()
        os.Exit(1)
      }
      tm.Printf(taskForm, t.ID, strconv.Itoa(t.Slot), qd.NodeMap[t.NodeID], t.State, fmt.Sprintf("%.1f", cElapse), cStatus, t.Image.PrintImage(), "<dunno>")
  }

  return nil
}

func (qd QnibDocker) CheckTaskHealth(task TaskConf) (string, float64, bool) {
  hReg := regexp.MustCompile("(healthy|healthy|starting)")
  tempCli, err := client.NewEnvClient()
  if err != nil {
    panic(err)
  }
  cfilter := filters.NewArgs()
  cfilter.Add("id", task.ContainerID)
  containers, _ := tempCli.ContainerList(context.Background(), types.ContainerListOptions{Filter: cfilter})
  var cElapse float64
  var cStatus string
  faulty := false
  if len(containers) == 1 {
    c := containers[0]
    cTime := time.Unix(c.Created,0)
    cElapse = time.Since(cTime).Seconds()
    if float64(qd.ServiceTimeout) < cElapse {
      faulty = true
    }
    cStatus = hReg.FindString(c.Status)
  }
  return cStatus, cElapse, faulty
}
