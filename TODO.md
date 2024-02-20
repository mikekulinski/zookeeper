# TODO List

- Initial service implementation on 1 process
  - Basic implementation of file system ZNodes
    - Build in memory using a trie
    - Add validation for valid paths and size of ZNodes
    - Potentially differentiate between files and directories
    - Add versioning (-1 means no version checks)
      - We do need way to choose if you want versioning during create
  - Add the concept of sessions
    - To prevent the session from timing out, the ZooKeeper client library sends a heartbeat
      after the session has been idle for s/3 ms and switch to a
      new server if it has not heard from a server for 2s/3 ms,
      where s is the session timeout in milliseconds.
  - Implement ephemeral ZNodes (can't have children)
  - Set up a way to run integration tests
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
  - Update Zookeeper executables to be a separate client/server
  - Expose via localhost over RPC / HTTP