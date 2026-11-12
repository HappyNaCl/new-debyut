package snowflake

import (
	"os"
	"strconv"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	once sync.Once
	node *snowflake.Node
)

func getNode() *snowflake.Node {
	once.Do(func() {
		nodeID, err := strconv.ParseInt(os.Getenv("SNOWFLAKE_NODE_ID"), 10, 64)
		if err != nil || nodeID < 0 || nodeID > 1023 {
			nodeID = 1
		}
		node, err = snowflake.NewNode(nodeID)
		if err != nil {
			panic("failed to create snowflake node: " + err.Error())
		}
	})
	return node
}

func Generate() string {
	return getNode().Generate().String()
}
