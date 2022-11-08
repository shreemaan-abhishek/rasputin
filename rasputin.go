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
	ldrshipDuration time.Duration
)

func Commission(candidateName string, client *clientv3.Client, leaseTimeToLive int, electionPrefix string, electionContext *context.Context, value string, leadershipDuration time.Duration) {
	name = candidateName
	cli = client
	pfx = electionPrefix
	ctx = electionContext
	val = value
	ldrshipDuration = leadershipDuration

	s, err := concurrency.NewSession(cli, concurrency.WithTTL(leaseTimeToLive))
	if err != nil {
	  log.Fatal(err)
	}
	electionSession = s
	election = concurrency.NewElection(electionSession, pfx)

	go watch(electionContext)

	fmt.Println("Rasputin!")
}

func Participate() {
	if err := election.Campaign(*ctx, val); err != nil {
		log.Fatal(err)
	}
	log.Printf("%s acquired leadership status", name)
	go giveUpLeadershipAfterDelay(ldrshipDuration)
}

func giveUpLeadershipAfterDelay(delay time.Duration) {
	time.Sleep(delay)
	if err := election.Resign(*ctx); err != nil {
		log.Printf("Failed to give up leadership for %s due to error: %s", name, err)
	}
	log.Println("Gave up leadership:", name)
	// re-participate as a candidate after giving up leadership
	Participate()
}

func Close() {
	log.Println("Closing rasputin")
	cli.Close()
	electionSession.Close()
}

func watch(ctx *context.Context) {
	log.Println("Watching...")
	<-(*ctx).Done()
	Close()
}