package main

import (
	"fmt"
	"log"
	"os/exec"
)

func execKubectlWithManifest(fileUrl string) int {
	log.Printf("Executing kubectl with definition file '%s'\n", fileUrl)
	cmd := exec.Command("kubectl", "apply", "-f", fmt.Sprintf("\"%s\"", fileUrl))

	if err := cmd.Run(); err != nil {
		log.Printf("Exec for command '%s' failed with error : %s", cmd.String(), err.Error())
		return 0
	}
	return 1
}
