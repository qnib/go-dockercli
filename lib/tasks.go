package dockerlib

import (
  "time"
  "regexp"


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
  ContainerID   string
  State         swarm.TaskState
  StateTime     time.Time
  DesiredState  swarm.TaskState
  CntStatus      string
  CntElapseSec  float64
  Faulty        bool
  HealthTimeout int
}

func NewTaskConf(task swarm.Task, img ImageConf, healthTimeout int) (TaskConf) {

  tc := TaskConf{
    ID: task.ID,
    UpdatedAt: task.Meta.UpdatedAt,
    CreatedAt: task.Meta.CreatedAt,
    Version: task.Meta.Version,
    Slot: task.Slot,
    Image: img,
    ImgUpdated: false,
    NodeID: task.NodeID,
    ContainerID: task.Status.ContainerStatus.ContainerID,
    State: task.Status.State,
    StateTime: task.Status.Timestamp,
    DesiredState: task.DesiredState,
    CntStatus: "",
    CntElapseSec: 0.0,
    Faulty: false,
    HealthTimeout: healthTimeout,
  }
  tc.CntStatus, tc.CntElapseSec, tc.Faulty = tc.CheckTaskHealth()
  return tc
}

func (tc TaskConf) CheckTaskHealth() (string, float64, bool) {
  hReg := regexp.MustCompile("(healthy|healthy|starting|no healthcheck)")
  tempCli, err := client.NewEnvClient()
  if err != nil {
    panic(err)
  }
  cfilter := filters.NewArgs()
  cfilter.Add("id", tc.ContainerID)
  containers, _ := tempCli.ContainerList(context.Background(), types.ContainerListOptions{Filter: cfilter})
  var cElapse float64
  var cStatus string
  faulty := false
  if len(containers) == 1 {
    c := containers[0]
    cTime := time.Unix(c.Created,0)
    cElapse = time.Since(cTime).Seconds()
    cStatus = hReg.FindString(c.Status)
    if (cStatus != "healthy") && (tc.HealthTimeout != 0) && (float64(tc.HealthTimeout) < cElapse) {
      faulty = true
    }

  }
  return cStatus, cElapse, faulty
}
