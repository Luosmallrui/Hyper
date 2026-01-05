package snowflake

import "github.com/bwmarrin/snowflake"

var node *snowflake.Node

func init() {
	node, _ = snowflake.NewNode(1)
}

func GenUserID() int64 {
	return node.Generate().Int64()
}
func GenID() int64 {
	return node.Generate().Int64()
}
