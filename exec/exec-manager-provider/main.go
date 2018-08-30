package main

import (
	"fmt"
	"github.com/eclipse/che-machine-exec/exec"
)

func main() {
	fmt.Println(exec.IsDockerInfra())
	fmt.Println(exec.IsKubernetesInfra())
}
