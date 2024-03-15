# TODO List

- Initial service implementation on 1 process
  - Fix race conditions
    - Make all the structures in our server thread safe
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
- Add an async version of the server
  - Maybe just have an async client that calls each method in a goroutine?
  - For FIFO client order we can use channels to implement this. And use blocking vs non-blocking channels for implementing sync / async