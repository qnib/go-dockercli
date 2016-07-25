package dockerlib

import (
  "fmt"
  "strconv"
  "container/list"
  "time"
  //"regexp"

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
  Services      []swarm.Service
  NodeMap       map[string]string
  RuStart       bool
  CurrentConf   map[string]StackConf
  NextConf      map[string]StackConf
  Logs          *list.List // flushed at each loop iteration
  Events        *list.List // not flushed, therefore kept while looping through
}

func NewQnibDocker() (QnibDocker) {
  cli, err := client.NewEnvClient()
  if err != nil {
    panic(err)
  }
  qd := QnibDocker{
    DockerCli: cli,
    RuStart: false,
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

func (qd QnibDocker) UpdateTaskList() (bool) {
  for n,s := range qd.NextConf {
    tfilter := filters.NewArgs()
    tfilter.Add("desired-state", "Running")
    tfilter.Add("service", n)
    tasks, err := qd.DockerCli.TaskList(context.Background(), types.TaskListOptions{Filter: tfilter})
    if err != nil {
        panic(err)
    }
    taskForm := "%15s %-35s %-20s %-20s %s\n"
    tm.Printf(taskForm, n, "Slot", "Node", "TaskStatus", "Image", "DesiredImage")
    for _, t := range tasks {
      if (s.Image.IsEqual(qd.CurrentConf[n].Image) && t.Status.State == "running") {
        tm.Printf(taskForm, strconv.Itoa(t.Slot), qd.NodeMap[t.NodeID], t.Status.State, t.Spec.ContainerSpec.Image, s.Image.Name)
      }
    }
  }
  return true
}
