package dockerlib

import (
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
  Version swarm.Version
  Slot    int
  Image   ImageConf
  Host    string
  State   string
}

func NewTaskConf(id string, ver swarm.Version, slot int, img ImageConf) (TaskConf) {
  return TaskConf{
    ID: id,
    Version: ver,
    Slot: slot,
    Image: img,
    Host: "",
    State: "",
  }
}
