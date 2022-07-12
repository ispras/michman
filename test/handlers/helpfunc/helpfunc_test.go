package helpfunc

import (
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/helpfunc"
	"testing"
)

var NUM = 1000000

func BenchmarkMergeSlices(b *testing.B) {
	var Ports1 []*protobuf.ServicePort
	var Ports2 []*protobuf.ServicePort

	for i := 1; i < NUM; i++ {
		Ports1 = append(Ports1, &protobuf.ServicePort{
			Port:        int32(i),
			Description: "merge_slice test",
		})
	}
	for i := NUM; i < NUM+NUM; i++ {
		Ports1 = append(Ports2, &protobuf.ServicePort{
			Port:        int32(i),
			Description: "merge_slice test",
		})
	}
	b.ResetTimer()
	helpfunc.MergeSlices(Ports1, Ports2, "Port")
}
