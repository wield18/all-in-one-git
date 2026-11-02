package consumer

import "golang.org/x/sync/semaphore"

func NewSignal[key comparable](IOPower int) *Signal[key] {
	return &Signal[key]{
		computingPower: int64(IOPower),
		Weighted:       semaphore.NewWeighted(int64(IOPower)),
	}
}

// func (s *Signal[k]) toMap() map[k]*Signal[k] {
// 	res := make(map[k]*Signal[k])
// 	cur := s
// 	for cur != nil {
// 		res[cur.k] = cur
// 		cur = cur.next
// 	}
// 	return res
// }

func (s *Signal[k]) WithK(key k) *Signal[k] {
	s.Key = key
	return s
}

func (s *Signal[k]) WithComputingPower(IOPower int64) *Signal[k] {
	s.computingPower = IOPower
	s.Weighted = semaphore.NewWeighted(int64(IOPower))
	return s
}

func (s *Signal[k]) WithNext(next *Signal[k], IOPowerExtend int64) *Signal[k] {
	s.next = next
	s.computingPowerExtend = IOPowerExtend
	return s
}
