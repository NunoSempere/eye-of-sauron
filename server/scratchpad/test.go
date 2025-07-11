package main
import (
  "fmt"
)

func main(){
  xs := []string{"a", "b", "c"}
  for i, _ := range xs {
    fmt.Printf("%v\n", xs[0:i])
  }
  fmt.Printf("%v", xs[2:len(xs)])
}
