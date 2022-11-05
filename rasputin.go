package rasputin

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	name string
	cli *clientv3.Client
	electionSession *concurrency.Session
	election *concurrency.Election
	ctx *context.Context
	pfx string
	val string
)

func Commission(candidateName string, client *clientv3.Client, leaseTimeToLive int, electionPrefix string, electionContext context.Context, value string) {
	name = candidateName
	cli = client
	pfx = electionPrefix
	ctx = &electionContext
	val = value

	s, err := concurrency.NewSession(cli, concurrency.WithTTL(leaseTimeToLive))
	if err != nil {
	  log.Fatal(err)
	}
	electionSession = s
	election = concurrency.NewElection(electionSession, pfx)

	fmt.Println("Rasputin!")
}

func participate() {
	if err := election.Campaign(*ctx, pfx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("leader election for", name)
	fmt.Println("Do some work in", name)
	time.Sleep(500 * time.Second)  
	if err := election.Resign(*ctx); err != nil {
	  log.Fatal(err)
	}
	fmt.Println("resign ", name)
}