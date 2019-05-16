package orchestrate

import (
	"io"
	"log"
	"os"
	"sync"

	"github.com/dcos/cluster-diagnostics/ssh"
)

func trigger(sshUsername, sshHost string) {
	client, err := ssh.NewAgentClient(sshUsername, sshHost)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	client.TransferFile("./cluster-diagnostics", "cluster-diagnostics")
	stdout, stderr, err := client.Execute("./cluster-diagnostics diagnose")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_, _, err := client.Execute("rm ./cluster-diagnostics")
		if err != nil {
			log.Fatal(err)
		}
	}()
	io.Copy(os.Stdout, stdout)
	io.Copy(os.Stderr, stderr)

}

func Orchestrate(user string, hosts []string) {

	var wg sync.WaitGroup
	for _, sshHost := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			trigger(user, host)
		}(sshHost)
	}

	wg.Wait()
}
