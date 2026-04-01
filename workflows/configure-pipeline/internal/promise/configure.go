package promise

import (
	"log"
	"os"

	cp "github.com/otiai10/copy"
)

func Configure() {
	if err := cp.Copy("/resources/dependencies", "/kratix/output"); err != nil {
		log.Printf("copy dependencies: %v", err)
		os.Exit(1)
	}
	log.Println("dependencies published - GitOps agents will provision them to destinations")
}
