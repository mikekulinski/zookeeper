package zxid

/*
The ZXID has two parts: the epoch and a counter. In our implementation the zxid is a 64-bit number.
We use the high order 32-bits for the epoch and the low order 32-bits for the counter.
Because it has two parts represent the zxid both as a number and as a pair of integers, (epoch, count).
The epoch number represents a change in leadership. Each time a new leader comes into power it will have its
own epoch number.
We have a simple algorithm to assign a unique zxid to a proposal:
- the leader simply increments the zxid to obtain a unique zxid for each proposal.
Leadership activation will ensure that only one leader uses a given epoch, so our simple algorithm guarantees that every
proposal will have a unique id. A new leader establishes a zxid to start using for new proposals by getting the epoch,
e, of the highest zxid it has seen and setting the next zxid to use to be (e+1, 0)
Taken from the official Zookeeper documentation: https://zookeeper.apache.org/doc/r3.4.13/zookeeperInternals.html#sc_guaranteesPropertiesDefinitions
*/
type ZXID int64

func NewZXID(epoch int32, counter int32) ZXID {
	// Set up the epoch and counter to be lined up with the high and low 32 bits of the zxid.
	var zxid int64 = 0
	highBits := int64(epoch) << 32
	lowBits := int64(counter)

	// Set the high and low bits of the zxid using bitwise OR.
	zxid |= highBits
	zxid |= lowBits
	return ZXID(zxid)
}

func (z ZXID) GetEpoch() int32 {
	// Get the epoch from the higher 32 bits of the zxid.
	return int32(z >> 32)
}

func (z ZXID) GetCounter() int32 {
	// Get the epoch from the lower 32 bits of the zxid. We do this by creating a bit mask of the lower 32 bits
	// and doing a bitwise AND to only get those bits.
	var maskLow32 ZXID = 0xFFFFFFFF
	return int32(z & maskLow32)
}
