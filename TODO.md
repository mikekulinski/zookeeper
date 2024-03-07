# TODO List

- Initial service implementation on 1 process
  - Fix race conditions
    - Make all the structures in our server thread safe
- Add an async version of the server
  - Maybe just have an async client that calls each method in a goroutine?
  - For FIFO client order we can use channels to implement this. And use blocking vs non-blocking channels for implementing sync / async
- Add replication for the Zookeeper server
  - Create a write-ahead log (WAL) that we can use as the history of all changes to the ZNodes
    - Maybe move this to disk at some point once we have multiple different processes running
    - Figure out how to implement snapshotting so we don't have a permanent gigantic log
  - Implement some sort of leader election
  - Redirect all writes to the leader
  - Use two-phase commit for replication
  - Implement atomic broadcast (ZAB)
- Split Zookeeper into a multiple processes
  - Set up a way to run integration tests