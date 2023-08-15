package loadbalance

import (
	"strings"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

var _ base.PickerBuilder = (*Picker)(nil)

type Picker struct {
	mu        sync.RWMutex
	leader    balancer.SubConn
	followers []balancer.SubConn
	current   uint64
}

func (p *Picker) Build(buildInfo base.PickerBuildInfo) balancer.Picker {
	p.mu.Lock()
	defer p.mu.Unlock()

	var followers []balancer.SubConn
	for sc, scInfo := range buildInfo.ReadySCs {
		isLeader := scInfo.Address.Attributes.Value("is_leader").(bool)
		if isLeader {
			p.leader = sc
			continue
		}
		followers = append(followers, sc)
	}
	p.followers = followers
	return p
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result balancer.PickResult
	if strings.Contains(info.FullMethodName, "Produce") || len(p.followers) == 0 { // Produce 메서드는 Leader 로 보낸다.
		result.SubConn = p.leader
	} else if strings.Contains(info.FullMethodName, "Consume") { // Consume 메서드는 Follower 로 보낸다.
		result.SubConn = p.nextFollower()
	}

	if result.SubConn == nil {
		return result, balancer.ErrNoSubConnAvailable
	}
	return result, nil
}

// 라운드 로빈 방식으로 다음 follower 고른다.
func (p *Picker) nextFollower() balancer.SubConn {
	cur := atomic.AddUint64(&p.current, uint64(1))
	len := uint64(len(p.followers))
	idx := int(cur % len)
	return p.followers[idx]
}

// gRPC 에 등록
func init() {
	balancer.Register(base.NewBalancerBuilder(Name, &Picker{}, base.Config{}))
}
