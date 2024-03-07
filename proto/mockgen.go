package zookeeper

//go:generate mockgen -destination mocks/proto_mock.go github.com/mikekulinski/zookeeper/proto Zookeeper_MessageClient,Zookeeper_MessageServer,ZookeeperClient
