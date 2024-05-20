package sync

import (
	"fmt"
	"testing"
	"time"

	"github.com/diiyw/nodis/pb"
)

func TestWebsocket_Publish(t *testing.T) {
	ws := NewWebsocket()
	go ws.Publish("127.0.0.1:8080", func(conn Conn) {
		fmt.Println("connected")
		conn.Send(&pb.Op{Operation: &pb.Operation{Type: pb.OpType_Set, Key: "test", Value: []byte("test")}})
	})
	time.Sleep(time.Second)
	go func() {
		err := ws.Subscribe("ws://127.0.0.1:8080", func(op *pb.Op) {
			if op.Operation.Key != "test" {
				t.Fail()
			}
		})
		if err != nil {
			t.Errorf("error: %v", err)
		}
	}()
	time.Sleep(time.Second)
}
