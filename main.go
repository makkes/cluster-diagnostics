package main

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/dcos/cluster-diagnostics/diagnose"
	"github.com/dcos/cluster-diagnostics/orchestrate"
	"github.com/dcos/cluster-diagnostics/ssh"
)

func initiate(jumpUser, jumpHost, user string, hosts []string) {

	client, err := ssh.NewClient(jumpUser, jumpHost, true)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	client.TransferFile("cluster-diagnostics.linux-amd64", "cluster-diagnostics")
	stdout, stderr, err := client.Execute("./cluster-diagnostics orchestrate " + user + " " + strings.Join(hosts, " "))
	defer func() {
		_, _, err := client.Execute("rm ./cluster-diagnostics")
		if err != nil {
			log.Fatal(err)
		}
	}()
	io.Copy(os.Stdout, stdout)
	io.Copy(os.Stderr, stderr)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a mode: initiate|orchestrate|diagnose")
	}

	if os.Args[1] == "orchestrate" {
		orchestrate.Orchestrate(os.Args[2], os.Args[3:])
		os.Exit(0)
	} else if os.Args[1] == "initiate" {
		if len(os.Args) < 6 {
			log.Fatalf("Usage: %s %s USER JUMPHOST DIAGUSER DIAGHOST...", os.Args[0], os.Args[1])
		}
		initiate(os.Args[2], os.Args[3], os.Args[4], os.Args[5:])
		os.Exit(0)
	} else if os.Args[1] == "diagnose" {
		diagnose.Diagnose()
	}
}
