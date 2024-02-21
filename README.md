# Zookeeper
## Summary
I am working on my own implementation of Zookeeper based on the original [white paper](www.usenix.org/legacy/event/atc10/tech/full_papers/Hunt.pdf). I'll be starting with a simple API for a system with just 1 server, and I plan on adding replication and failure recovery as outlined in the white paper and other resources online. 
Depending on how well this goes I'll try implementing some of the applications described in the white paper on top of the this to see how well it performs. I'm thinking probably some sort of distributed locking or a basic Kafka / Yahoo! Message Broker (YMB) implementation.

** **This is purely an educational exercise. There will almost certainly be bugs that I haven't caught and should not be used for a real production system.** **
