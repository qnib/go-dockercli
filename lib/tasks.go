
package dockerlib

import (
  "time"
  "regexp"
  "fmt"
  "strings"

  "github.com/docker/engine-api/types/swarm"
  "github.com/docker/engine-api/client"
  "github.com/docker/engine-api/types"
  "github.com/docker/engine-api/types/filters"
  "golang.org/x/net/context"
)

type QnibTasks struct {
  Tasks   []TaskConf
}

func NewQnibTasks() (QnibTasks) {
  return QnibTasks{
    Tasks: []TaskConf{},
  }
}

func (qt QnibTasks) IsItem(tc TaskConf) (bool) {
  for _, i := range qt.Tasks {
    if i.ID == tc.ID {
      return true
    }
  }
  return false
}

type TaskConf struct {
  ID            string
  UpdatedAt     time.Time
  CreatedAt     time.Time
  Version       swarm.Version
  Slot          int
  Image         ImageConf
  ImgUpdated    bool
  NodeID        string
  HostName      string
  ContainerID   string
  State         swarm.TaskState
  StateTime     time.Time
  DesiredState  swarm.TaskState
  CntStatus     string
  Labels        map[string]string
  CntCreatedAt  int
  CntElapseSec  float64
  Faulty        bool
  HealthTimeout int
}

func NewTaskConf(task swarm.Task, img ImageConf, healthTimeout int, hostName string) (TaskConf) {

  tc := TaskConf{
    ID: task.ID,
    UpdatedAt: task.Meta.UpdatedAt,
    CreatedAt: task.Meta.CreatedAt,
    Version: task.Meta.Version,
    Slot: task.Slot,
    Image: img,
    ImgUpdated: false,
    NodeID: task.NodeID,
    HostName: hostName,
    ContainerID: task.Status.ContainerStatus.ContainerID,
    State: task.Status.State,
    StateTime: task.Status.Timestamp,
    DesiredState: task.DesiredState,
    CntStatus: "",
    Labels: make(map[string]string),
    CntCreatedAt: 0,
    CntElapseSec: 0.0,
    Faulty: false,
    HealthTimeout: healthTimeout,
  }
  tc.CntStatus, tc.CntElapseSec, tc.CntCreatedAt, tc.Labels, tc.Faulty = tc.CheckTaskHealth()
  return tc
}

func (tc TaskConf) PrintLabels(lReg string) (string) {
  lr := regexp.MustCompile(lReg)
  res := []string{}
  for k,v := range tc.Labels {
    if lr.MatchString(k) {
      res = append(res, fmt.Sprintf("%s=%s", k, v))
    }
  }
  return strings.Join(res, ",")
}

func (tc TaskConf) CheckTaskHealth() (string, float64, int, map[string]string, bool) {
  hReg := regexp.MustCompile("(healthy|healthy|starting|no healthcheck)")
  defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
  tempCli, err := client.NewClient(fmt.Sprintf("tcp://%s:2376", tc.HostName), "v1.24", nil, defaultHeaders)
  if err != nil {
    panic(err)
  }
  cfilter := filters.NewArgs()
  cfilter.Add("id", tc.ContainerID)
  containers, _ := tempCli.ContainerList(context.Background(), types.ContainerListOptions{Filter: cfilter})
  var cElapse float64
  var cStatus string
  var cTime int
  var cLab map[string]string
  faulty := false
  if len(containers) == 1 {
    c := containers[0]
    cLab = c.Labels
    cTime := time.Unix(c.Created,0)
    cElapse = time.Since(cTime).Seconds()
    cStatus = hReg.FindString(c.Status)
    if (cStatus != "healthy") && (tc.HealthTimeout != 0) && (float64(tc.HealthTimeout) < cElapse) {
      faulty = true
    }

  }
  return cStatus, cElapse, cTime, cLab, faulty
}
