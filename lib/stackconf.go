package dockerlib

import (
  "fmt"
  "strings"
)

type ImageConf struct {
  Registry string
  Repo     string
  Name     string
  Sha      bool
  Tag      string
}

func NewImageConf(image string) (ImageConf) {
  ic := ImageConf{
    Registry: "",
    Repo: "",
    Name: "",
    Sha: false,
    Tag: "latest",
  }
  var name string
  parts := strings.Split(image, "/")
  switch len(parts) {
  case 1:
      //fmt.Printf("Official image: %s\n", parts[0])
      name = parts[0]
  case 2:
      //fmt.Printf("Image w/o explicit registry repo:%s name:%s\n", parts[0], parts[1])
      ic.Registry = parts[0]
      name = parts[1]
  case 3:
      //fmt.Printf("Image w explicit registry:%s repo:%s name:%s\n", parts[0], parts[1], parts[2])
      ic.Registry = parts[0]
      ic.Repo = parts[1]
      name = parts[2]
  }
  parts = strings.Split(name, ":")
  if len(parts) == 1 {
    ic.Name = name
  } else {
    // It is either the tag or the sha256 hash
    ic.Tag = parts[1]
    nparts := strings.Split(parts[0], "@")
    ic.Name = nparts[0]
    // determine whether it's the tag or a hash
    if (len(nparts) == 2) {
      ic.Sha = true
    }
  }
  return ic
}

// Returns assembled image name registry/repo/name
func (ic ImageConf) PrintImage() (string) {
  s := []string{ic.Registry, ic.Repo, ic.Name}
  return strings.Join(s, "/")
}

// Returns Tag/sha256 hash
func (ic ImageConf) PrintTag() (string) {
  if ic.Sha {
    return fmt.Sprintf("@sha256:%s", ic.Tag)
  }
  return ic.Tag
}


func (ic ImageConf) IsEqual(other ImageConf) (bool) {
  return (ic.Registry == other.Registry) && (ic.Repo == other.Repo) && (ic.Name == other.Name) && (ic.Sha == other.Sha) && (ic.Tag == other.Tag)
}




type StackConf struct {
  Image     ImageConf
  Replicas  int
}

func NewStackConf(image string, repl int) (StackConf){
  return StackConf{
    Image: NewImageConf(image),
    Replicas: repl,
  }

}
