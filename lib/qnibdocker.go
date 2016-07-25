package dockerlib

import (
  "strconv"
  "time"

  tm "github.com/buger/goterm"
  "github.com/docker/engine-api/client"
  "github.com/docker/engine-api/types"
  "github.com/docker/engine-api/types/filters"
  "golang.org/x/net/context"
)



type QnibDocker struct {
  DockerCli     *client.Client
  NodeMap       map[string]string
  RuStart       bool
  CurrentConf   map[string]StackConf
  NextConf      map[string]StackConf
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
   }
  return qd
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

func (qd QnibDocker) UpdateServiceList() (bool) {
  ret := qd.RuStart
  services, err := qd.DockerCli.ServiceList(context.Background(), types.ServiceListOptions{})
  if err != nil {
      panic(err)
  }
  tm.Printf(">>> Services \t\t(%s)\n", time.Now().Format(time.RFC3339))
  srvForm := "%-15s %-10s %-30s\n"
  tm.Printf(srvForm, "Name", "Replicas", "Image")
  for _,s := range services {
    replicas := int(*s.Spec.Mode.Replicated.Replicas)
    srvName := s.Spec.Annotations.Name
    srvImage := s.Spec.TaskTemplate.ContainerSpec.Image
    tm.Printf(srvForm, srvName, strconv.Itoa(replicas), srvImage)
    sc := NewStackConf(srvImage, "", replicas)
    qd.UpdateConf(sc)
  }
  
  return ret
}

func (qd QnibDocker) UpdateServiceConf(sc StackConf) {
  _, srv := qd.CurrentConf[srvName]
  if ! srv {
    qd.CurrentConf[srvName] = sc
  } else if ! qd.RuStart {
    // If RU had not yet started, we check for changes in the CurrentConf
    if ( qd.CurrentConf[srvName].Replicas != replicas ) || ( qd.CurrentConf[srvName].ImageName != srvImage ) {
      tm.Printf("%s: has changed! %b\n", srvName, qd.RuStart)
      ret = true
      qd.NextConf[srvName] = sc
    } else {
      tm.Printf("%s: are still the same\n", srvName)
    }
  } else {
    // if RU had started, we wait for it to finish successfully
    tm.Printf("%s: on his way\n", srvName)
    tm.Printf("old: %s\n", qd.CurrentConf)
    tm.Printf("new: %s\n", qd.NextConf)
  }
}


func (qd QnibDocker) IsRuFinished() (bool) {
  tm.Println("Don't think so...")
  return false
}

func (qd QnibDocker) UpdateTaskList() (bool) {
  for n,s := range qd.NextConf {
    allRunning := true
    tfilter := filters.NewArgs()
    tfilter.Add("desired-state", "Running")
    tfilter.Add("service", n)
    tasks, err := qd.DockerCli.TaskList(context.Background(), types.TaskListOptions{Filter: tfilter})
    if err != nil {
        panic(err)
    }
    taskForm := "%15s %-10s %-35s %-20s %-20s %s\n"
    tm.Printf(taskForm, n, "Slot", "Node", "TaskStatus", "Image", "DesiredImage")
    for _, t := range tasks {
      if (t.Spec.ContainerSpec.Image == s.ImageName) && (t.Status.State == "running") {
        tm.Printf(taskForm, allRunning, strconv.Itoa(t.Slot), qd.NodeMap[t.NodeID], t.Status.State, t.Spec.ContainerSpec.Image, s.ImageName)
      }
    }
  }
  return true
}
