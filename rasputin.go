package rasputin

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Rasputin struct {
	name               string
	client             *clientv3.Client
	electionSession    *concurrency.Session
	election           *concurrency.Election
	ctx                *context.Context
	prefix             string
	val                string
	leadershipDuration time.Duration
	currentLeaderKey   []byte
}

func Commission(candidateName string, client *clientv3.Client, leaseTimeToLive int, electionPrefix string, electionContext *context.Context, value string, leadershipDuration time.Duration) *Rasputin {

	s, err := concurrency.NewSession(client, concurrency.WithTTL(leaseTimeToLive))
	if err != nil {
		log.Fatal(err)
	}
	e := concurrency.NewElection(s, electionPrefix)

	r := &Rasputin{
		name: candidateName,
		client: client,
		electionSession: s,
		election: e,
		ctx: electionContext,
		prefix: electionPrefix,
		val: value,
		leadershipDuration: leadershipDuration,
	}

	go r.watch(electionContext)
	go r.observe()

	fmt.Println("Rasputin!")
	return r
}

func (r *Rasputin) observe() {
	cres := r.election.Observe(context.Background())
	for response := range cres {
		fmt.Println("received response", response)
		r.currentLeaderKey = response.Kvs[0].Key
	}
}

func (r *Rasputin) IsLeader() bool {
	//TODO: doc
	return r.election.Key() == string(r.currentLeaderKey)
}

func (r *Rasputin) Participate() {
	go func() {
		if err := r.election.Campaign(*r.ctx, r.val); err != nil {
			log.Fatal(err)
		}
		log.Printf("%s acquired leadership status", r.name)
		r.giveUpLeadershipAfterDelay(r.leadershipDuration)
	}()
}

func (r *Rasputin) giveUpLeadershipAfterDelay(delay time.Duration) {
	time.Sleep(delay)
	if err := r.election.Resign(*r.ctx); err != nil {
		log.Printf("Failed to give up leadership for %s due to error: %s", r.name, err)
	}
	log.Println("Gave up leadership:", r.name)
	// re-participate as a candidate after giving up leadership
	r.Participate()
}

func (r *Rasputin) Close() {
	log.Println("Closing rasputin")
	r.client.Close()
	r.electionSession.Close()
}

func (r *Rasputin) watch(ctx *context.Context) {
	log.Println("Watching...")
	<-(*ctx).Done()
	r.Close()
}
