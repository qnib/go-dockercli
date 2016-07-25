package dockerlib

type StackConf struct {
  ImageName string
  ImageTag  string
  Replicas  int
}

func NewStackConf(name string, tag string, rep int) (StackConf){
  return StackConf{
    ImageName: name,
    ImageTag: tag,
    Replicas: rep,
  }

}
