# rasputin
Rasputin is a leader election client. _To understand the basics of leader election checkout [this](https://aws.amazon.com/builders-library/leader-election-in-distributed-systems/) blog._

Rasputin sits on top of the etcd client and provides a succinct API for leader election and related chores. Using rasputin you are relieved off the load of writing code to perform simple tasks like periodic leadership shed, to check current leadership status, and to listen for leadership status changes.

## Usage:

Initialise your etcd client as you prefer and pass it to the rasputin constructor function with other parameters.

```go
	etcdClient, err := client.New(client.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		log.Fatal(err)
	}

	c := context.Background()
	ctx, cancel := context.WithCancel(c)
	rasp, err := rasputin.Commission(etcdClient, 1, "/ldr", &ctx, "val", 10 * time.Second)
	if err != nil {
		fmt.Println("Failed to commission rasputin due to error:", err)
		// recovery logic
	}
```
Start leader election:

```go
	statusCh, errCh := rasp.Participate()
```

`statusCh` gives leadership status updates and errCh gives errors if any.
Look for errors like so:

```go
	for err := range errCh {
		fmt.Println("Failed to participate for leader election due to error:", err)
		// recovery logic
	}
```

Look for leadership status changes like so:

```go
	for isLeader := range statusCh {
		if isLeader {
			log.Println("Wohooo! I became a leader")
		} else {
			log.Println("Lost the leadership :(")
		}
	}
```

You can make a process resign from the leadership status anytime by calling:

```go
rasputin.Resign()
```

To understand the basics of leader election using etcd checkout [this](https://shreemaan-abhishek.hashnode.dev/microservice-leader-election-using-etcd) blog.
