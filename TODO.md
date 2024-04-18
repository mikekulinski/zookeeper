# TODO List

- Initial service implementation on 1 process
  - Fix race conditions
    - Make all the structures in our server thread safe
    - Do this by splitting the update operations to go through a DB that
      handles all the locking
- Request Processor
  - Implement a processor that handles write requests
  - Send transaction to leader
  - Leader will get consensus among all servers
  - Once we receive a commit message, we will commit on our local log
- Create a write-ahead log (WAL) that we can use as the history of all changes to the ZNodes
  - Maybe move this to disk at some point once we have multiple different processes running
  - Figure out how to implement snapshotting so we don't have a permanent gigantic log
- Implement atomic broadcast (ZAB)
  - https://zookeeper.apache.org/doc/r3.4.13/zookeeperInternals.html#sc_logging
  - Include zxid in each transaction and response to the client
  - Implement some sort of leader election
  - Redirect all writes to the leader
  - Use two-phase commit for replication
  - Add connect request with the last zxid we saw so that we can use to wait until the server we're connecting to is caught up
  - Figure out how to handle version validation across replicas. Can we use the local DB
    on each replica? Or do we have to do it at the leader?
- Add an async version of the server
  - Maybe just have an async client that calls each method in a goroutine?
  - For FIFO client order we can use channels to implement this. And use blocking vs non-blocking channels for implementing sync / async