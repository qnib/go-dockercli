package dockerlib

import (
  "fmt"
  "strconv"
  "container/list"
  "time"
  "regexp"
  "strings"

  tm "github.com/buger/goterm"
  "github.com/docker/engine-api/client"
  "github.com/docker/engine-api/types"
  "github.com/docker/engine-api/types/filters"
  "github.com/docker/engine-api/types/swarm"
  "golang.org/x/net/context"
)


//var imageExp = regexp.MustCompile(`(?P<registry>[a-zA-Z0-9\.\-]+)/(?P<registry>[a-zA-Z0-9\.\-]+)`)

type QnibDocker struct {
  DockerCli       *client.Client
  ServiceList     string
  LabelReg        string
  ServiceTimeout  int
  PrintFaulty     bool
  NoPrint          bool
  Services        []swarm.Service
  SrvTasks        map[string][]TaskConf
  NodeMap         map[string]string
  SrvConf         map[string]StackConf
  Logs            *list.List // flushed at each loop iteration
  Events          *list.List // not flushed, therefore kept while looping through
}

func NewQnibDocker(serviceList string, timeout int, pFaulty bool, noPrint bool, labelReg string) (QnibDocker) {
  cli, err := client.NewEnvClient()
  if err != nil {
    panic(err)
  }
  qd := QnibDocker{
    DockerCli: cli,
    ServiceTimeout: timeout,
    ServiceList: serviceList,
    LabelReg: labelReg,
    PrintFaulty: pFaulty,
    NoPrint: noPrint,
    NodeMap: make(map[string]string),
    SrvConf: make(map[string]StackConf),
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
    if ! qd.NoPrint {
      tm.Printf("\n\n>> Events\n")
    }
  }
  for e := qd.Events.Front(); e != nil; e = e.Next() {
    if ! qd.NoPrint {
      tm.Println(e.Value)
    }
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
      if ! qd.NoPrint {
        tm.Printf("\n\n>> Logs within Loop (flushed afterwards)\n")
      }
  }
  for e := qd.Logs.Front(); e != nil; e = e.Next() {
    if ! qd.NoPrint {
      tm.Println(e.Value)
    }
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
    srvName := s.Spec.Annotations.Name
    if (qd.ServiceList != "") {
      cont := true
      for _, s := range strings.Split(qd.ServiceList, ",") {
        if (s == srvName) {
          cont = false
        }
      }
      if cont {
        continue
      }
    }
    replicas := 0
    if s.Spec.Mode.Replicated != nil {
      replicas = int(*s.Spec.Mode.Replicated.Replicas)
    }
    srvImage := s.Spec.TaskTemplate.ContainerSpec.Image
    sc := NewStackConf(srvImage, replicas)
    qd.UpdateServiceConf(srvName, sc)
  }
  return services, nil
}

func (qd QnibDocker) PrintServices() {
  srvForm := " %-15s %-10s %-40s %-40s\n"
  if ! qd.NoPrint {
    tm.Printf(srvForm, "Name", "Replicas", "Image", "Tag")
  }
  for _,s := range qd.Services {
    replicas := 0
    if s.Spec.Mode.Replicated != nil {
      replicas = int(*s.Spec.Mode.Replicated.Replicas)
    }
    srvName := s.Spec.Annotations.Name
    srvImage := s.Spec.TaskTemplate.ContainerSpec.Image
    ic := NewImageConf(srvImage)
    strReplicas := strconv.Itoa(replicas)
    if strReplicas == "0" {
      strReplicas = "Global"
    }
    if ! qd.NoPrint {
      tm.Printf(srvForm, srvName, strReplicas, ic.PrintImage(), ic.PrintTag())
    }
    qd.PrintTasks(s)
  }
}

func (qd QnibDocker) UpdateServiceConf(srvName string, sc StackConf) {
  _, srv := qd.SrvConf[srvName]
  if ! srv {
    qd.SrvConf[srvName] = sc
  }
}

func (qd QnibDocker) UpdateTaskList() (map[string][]TaskConf) {
  qt := make(map[string][]TaskConf)
  for _,s := range qd.Services {
    srvName := s.Spec.Annotations.Name
    if (qd.ServiceList != "") {
      cont := true
      for _, s := range strings.Split(qd.ServiceList, ",") {
        if (s == srvName) {
          cont = false
        }
      }
      if cont {
        continue
      }
    }
    srvImage := s.Spec.TaskTemplate.ContainerSpec.Image
    ic := NewImageConf(srvImage)
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
      nTask := NewTaskConf(t, tic, qd.ServiceTimeout, qd.NodeMap[t.NodeID])
      imgEq := ic.IsEqual(nTask.Image)
      if imgEq {
        nTask = nTask.SetImgUpdated(true)
      }
      qt[srvName] = append(qt[srvName], nTask)

    }

  }
  return qt
}

func (qd QnibDocker) PrintTasks(s swarm.Service) (error) {
  srv := s.Spec.Annotations.Name
  taskForm := "   >> %-7s %-30s %-10s %-10s %-12s %-25s %-25s %v\n"
  if s.Spec.Mode.Replicated == nil {
    taskForm = "   >> %-30s %-10s %-10s %-12s %-25s %-25s %v\n"
  }
  if ! qd.NoPrint {
    if s.Spec.Mode.Replicated != nil {
      tm.Printf(taskForm, "Slot", "Node", "TaskState", "SecSince", "CntStatus", "Image", "Tag", "Labels")
    } else {
      tm.Printf(taskForm, "Node", "TaskState", "SecSince", "CntStatus", "Image", "Tag", "Labels")
    }
  }
  for _, t := range qd.SrvTasks[srv] {
    if s.Spec.Mode.Replicated != nil {
      if ! qd.NoPrint {
        tm.Printf(taskForm, strconv.Itoa(t.Slot), qd.NodeMap[t.NodeID], t.State, fmt.Sprintf("%-5.1f", t.CntElapseSec), t.CntStatus, t.Image.PrintImageName(), t.Image.PrintTag(), t.PrintLabels(qd.LabelReg))
      }
    } else {
      if ! qd.NoPrint {
        tm.Printf(taskForm, qd.NodeMap[t.NodeID], t.State, fmt.Sprintf("%-5.1f", t.CntElapseSec), t.CntStatus, t.Image.PrintImageName(), t.Image.PrintTag(), t.PrintLabels(qd.LabelReg))
      }
    }
  }

  return nil
}

func (qd QnibDocker) CheckRUFinish() (bool, int){
  allUpdated := true
  allHealthy := true
  someFaulty := false
  for _,srv := range qd.Services {
    srvName := srv.Spec.Annotations.Name
    for _, t := range qd.SrvTasks[srvName] {
      if t.Faulty {
        someFaulty = true
      }
      if t.CntStatus != "healthy" {
        allHealthy = false
      }
      if ! t.ImgUpdated {
        allUpdated = false
      }
    }
  }
  if (qd.TaskCountOK() && allUpdated && allHealthy) && ! someFaulty {
    return true, 0
  }
  if someFaulty {
    return true, 1
  }
  return false, 0
}

func (qd QnibDocker) TaskCountOK() (bool) {
  allCntOK := true
  for _,srv := range qd.Services {
    replicas := 0
    if srv.Spec.Mode.Replicated != nil {
      replicas = int(*srv.Spec.Mode.Replicated.Replicas)
    }
    srvName := srv.Spec.Annotations.Name
    if (qd.ServiceList != "") {
      cont := true
      for _, s := range strings.Split(qd.ServiceList, ",") {
        if (s == srvName) {
          cont = false
        }
      }
      if cont {
        continue
      }
    }
    if srv.Spec.Mode.Replicated != nil && replicas != 0 && len(qd.SrvTasks[srvName]) != replicas {
        allCntOK = false
    }
  }
  return allCntOK
}

func (qd QnibDocker) CheckTaskHealth(task TaskConf) (string, float64, bool) {
  hReg := regexp.MustCompile("(healthy|healthy|starting)")
  defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
  dHost := fmt.Sprintf("tcp://%s:2376", qd.NodeMap[task.NodeID])
  tm.Println(dHost)
  tempCli, err := client.NewClient(dHost, "v1.24", nil, defaultHeaders)
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
    cStatus = hReg.FindString(c.Status)
    if (cStatus != "healthy") && (qd.ServiceTimeout != 0) && (float64(qd.ServiceTimeout) < cElapse) {
      faulty = true
    }

  } else {
    msg := fmt.Sprintf("No container found for ID '%s' on Host '%s'", task.ContainerID, qd.NodeMap[task.NodeID])
    tm.Println(msg)
    qd.AddLog(msg)

  }
  return cStatus, cElapse, faulty
}
