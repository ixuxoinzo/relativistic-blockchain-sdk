package tests
import (
        "time"
	"strconv"

        "github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)
func CreateTestNode(id string, lat, lon float64) *types.Node {
        return &types.Node{
                ID: id,
                Position: types.Position{
                        Latitude:  lat,
                        Longitude: lon,
                        Altitude:  0,
                },
                Address: id + ".example.com:8080",
                Metadata: types.Metadata{
                        Region:   "test",
                        Provider: "test",
                        Version:  "1.0.0",
                },
                IsActive: true,
                LastSeen: time.Now().UTC(),
        }
}
func CreateTestBlock(hash, proposedBy string, lat, lon float64) *types.Block {
        return &types.Block{
                Hash:       hash,
                Timestamp:  time.Now().UTC(),
                ProposedBy: proposedBy,
                NodePosition: types.Position{
                        Latitude:  lat,
                        Longitude: lon,
                        Altitude:  0,
                },
                Data: []byte("test data"),
        }
}
func CreateTestTransaction(hash string, lat, lon float64) *types.Transaction {
        return &types.Transaction{
                Hash:      hash,
                Timestamp: time.Now().UTC(),
                NodePosition: types.Position{
                        Latitude:  lat,
                        Longitude: lon,
                        Altitude:  0,
                },
                Data: []byte("test transaction"),
        }
}
func GenerateTestNodes(count int, startLat, startLon float64) []*types.Node {
        nodes := make([]*types.Node, count)
        for i := 0; i < count; i++ {
                nodes[i] = CreateTestNode(
                       "node-" + strconv.Itoa(i),
                        startLat+float64(i)*0.1,
                        startLon+float64(i)*0.1,
                )
        }
        return nodes
}
