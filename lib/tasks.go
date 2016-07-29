package dockerlib

import (
  "time"

  "github.com/docker/engine-api/types/swarm"
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
  ID    string
  UpdatedAt time.Time
  CreatedAt time.Time
  Version swarm.Version
  Slot    int
  Image   ImageConf
  NodeID    string
  ContainerID string
  State   swarm.TaskState
  StateTime time.Time
  DesiredState swarm.TaskState
  CntState string
}

func NewTaskConf(task swarm.Task, img ImageConf) (TaskConf) {
  return TaskConf{
    ID: task.ID,
    UpdatedAt: task.Meta.UpdatedAt,
    CreatedAt: task.Meta.CreatedAt,
    Version: task.Meta.Version,
    Slot: task.Slot,
    Image: img,
    NodeID: task.NodeID,
    ContainerID: task.Status.ContainerStatus.ContainerID,
    State: task.Status.State,
    StateTime: task.Status.Timestamp,
    DesiredState: task.DesiredState,
    CntState: "",
  }
}
