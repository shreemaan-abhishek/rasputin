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
	client             *clientv3.Client
	electionSession    *concurrency.Session
	election           *concurrency.Election
	ctx                *context.Context
	prefix             string
	val                string
	leadershipDuration time.Duration
	currentLeaderKey   []byte
	statusCh           chan bool
}

var isLeader bool = false

func Commission(client *clientv3.Client, leaseTimeToLive int, electionPrefix string, electionContext *context.Context, value string, leadershipDuration time.Duration) (*Rasputin, error) {

	s, err := concurrency.NewSession(client, concurrency.WithTTL(leaseTimeToLive))
	if err != nil {
		return nil, err
	}
	e := concurrency.NewElection(s, electionPrefix)
	statusCh := make(chan bool)
	r := &Rasputin{
		client:             client,
		electionSession:    s,
		election:           e,
		ctx:                electionContext,
		prefix:             electionPrefix,
		val:                value,
		leadershipDuration: leadershipDuration,
		statusCh:           statusCh,
	}

	go r.waitCleanup(electionContext)
	go r.observe()

	fmt.Println("Rasputin!")
	return r, nil
}

// Observe leadership status changes and notify client via a channel
func (r *Rasputin) observe() {
	cres := r.election.Observe(*r.ctx)
	for response := range cres {
		r.currentLeaderKey = response.Kvs[0].Key
		// isLeader denots previous leadership status as it hasn't been updated yet
		// (r.election.Key() == string(r.currentLeaderKey)) denotes current leadership status
		if !isLeader && (r.election.Key() == string(r.currentLeaderKey)) {
			r.statusCh <- true
		}
		if isLeader && (r.election.Key() != string(r.currentLeaderKey)) {
			r.statusCh <- false
		}
		isLeader = (r.election.Key() == string(r.currentLeaderKey))
	}
}

func (r *Rasputin) IsLeader() bool {
	return isLeader
}

// Participate in leader election, statusCh will produce true when the current instance
// becomes a leader and false when the current instance loses leadership.
func (r *Rasputin) Participate() (<-chan bool, <-chan error) {
	cherr := make(chan error)
	go func() {
		if err := r.election.Campaign(*r.ctx, r.val); err != nil {
			cherr<- err
		}
		r.giveUpLeadershipAfterDelay(r.leadershipDuration)
	}()

	return r.statusCh, cherr
}

func (r *Rasputin) giveUpLeadershipAfterDelay(delay time.Duration) {
	time.Sleep(delay)
	if err := r.election.Resign(*r.ctx); err != nil {
		log.Printf("Failed to give up leadership due to error: %s", err)
	}
	// re-participate as a candidate after giving up leadership
	r.Participate()
}

func (r *Rasputin) Close() {
	log.Println("Closing rasputin, freeing resources")
	r.client.Close()
	r.electionSession.Close()
	close(r.statusCh)
}

// Waits for context cancellation to cleanup resources
func (r *Rasputin) waitCleanup(ctx *context.Context) {
	log.Println("Watching...")
	<-(*ctx).Done()
	r.Close()
}
