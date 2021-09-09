package accounting

import (
	"sync"
)

const DefaultL3FlowCollectionSize = 16
const DefaultL4FlowCollectionSize = 64
const DefaultFlowCollectionHistorySize = 300

type FlowFingerprint struct {
	SrcAddr  string
	DstAddr  string
	SrcPort  uint16
	DstPort  uint16
	Protocol string
}

type Flow struct {
	FlowFingerprint
	InboundBytes     int64
	InboundPackets   int64
	InboundDuration  int64
	OutboundBytes    int64
	OutboundPackets  int64
	OutboundDuration int64
}

type FlowTimestamp struct {
	Start int64
	End   int64
}

func (t *FlowTimestamp) Offset(offset int64) (ts FlowTimestamp) {
	ts.Start = t.Start + offset
	ts.End = t.End + offset

	return
}

type FlowCollection struct {
	InterfaceName string
	FlowTimestamp
	L3FlowMap map[FlowFingerprint]*Flow
	L4FlowMap map[FlowFingerprint]*Flow
	Mu        *sync.Mutex
}

func NewFlowCollection(ifaceName string) (flowCol *FlowCollection) {
	flowCol = &FlowCollection{
		InterfaceName: ifaceName,
		L3FlowMap:     make(map[FlowFingerprint]*Flow, DefaultL3FlowCollectionSize),
		L4FlowMap:     make(map[FlowFingerprint]*Flow, DefaultL4FlowCollectionSize),
		Mu:            &sync.Mutex{},
	}

	return
}

func (c *FlowCollection) SetTimestamp(start int64, end int64) {
	c.Start = start
	c.End = end
}

func (c *FlowCollection) UpdateL3Inbound(flowFp FlowFingerprint, numBytes int64, numPkts int64, duration int64) {
	flow, ok := c.L3FlowMap[flowFp]
	if !ok {
		c.Mu.Lock()
		c.L3FlowMap[flowFp] = &Flow{
			FlowFingerprint: flowFp,
			InboundBytes:    numBytes,
			InboundPackets:  numPkts,
			InboundDuration: duration,
		}
		c.Mu.Unlock()
	} else {
		c.Mu.Lock()
		flow.InboundBytes += numBytes
		flow.InboundPackets += numPkts
		flow.InboundDuration += duration
		c.Mu.Unlock()
	}
}

func (c *FlowCollection) UpdateL3Outbound(flowFp FlowFingerprint, numBytes int64, numPkts int64, duration int64) {
	flow, ok := c.L3FlowMap[flowFp]
	if !ok {
		c.Mu.Lock()
		c.L3FlowMap[flowFp] = &Flow{
			FlowFingerprint:  flowFp,
			OutboundBytes:    numBytes,
			OutboundPackets:  numPkts,
			OutboundDuration: duration,
		}
		c.Mu.Unlock()
	} else {
		c.Mu.Lock()
		flow.OutboundBytes += numBytes
		flow.OutboundPackets += numPkts
		flow.OutboundDuration += duration
		c.Mu.Unlock()
	}
}

func (c *FlowCollection) UpdateL4Inbound(flowFp FlowFingerprint, numBytes int64, numPkts int64, duration int64) {
	flow, ok := c.L4FlowMap[flowFp]
	if !ok {
		c.Mu.Lock()
		c.L4FlowMap[flowFp] = &Flow{
			FlowFingerprint: flowFp,
			InboundBytes:    numBytes,
			InboundPackets:  numPkts,
			InboundDuration: duration,
		}
		c.Mu.Unlock()
	} else {
		c.Mu.Lock()
		flow.InboundBytes += numBytes
		flow.InboundPackets += numPkts
		flow.InboundDuration += duration
		c.Mu.Unlock()
	}
}

func (c *FlowCollection) UpdateL4Outbound(flowFp FlowFingerprint, numBytes int64, numPkts int64, duration int64) {
	flow, ok := c.L4FlowMap[flowFp]
	if !ok {
		c.Mu.Lock()
		c.L4FlowMap[flowFp] = &Flow{
			FlowFingerprint:  flowFp,
			OutboundBytes:    numBytes,
			OutboundPackets:  numPkts,
			OutboundDuration: duration,
		}
		c.Mu.Unlock()
	} else {
		c.Mu.Lock()
		flow.OutboundBytes += numBytes
		flow.OutboundPackets += numPkts
		flow.OutboundDuration += duration
		c.Mu.Unlock()
	}
}

func (c *FlowCollection) UpdateByFlowCol(fc *FlowCollection) {
	c.Mu.Lock()
	for _, f := range fc.L3FlowMap {
		flow, ok := c.L3FlowMap[f.FlowFingerprint]
		if !ok {
			ff := *f
			c.L3FlowMap[f.FlowFingerprint] = &ff
		} else {
			flow.InboundBytes += f.InboundBytes
			flow.InboundPackets += f.InboundPackets
			flow.InboundDuration += f.InboundDuration
			flow.OutboundBytes += f.OutboundBytes
			flow.OutboundPackets += f.OutboundPackets
			flow.OutboundDuration += f.OutboundDuration
		}
	}

	for _, f := range fc.L4FlowMap {
		flow, ok := c.L4FlowMap[f.FlowFingerprint]
		if !ok {
			ff := *f
			c.L4FlowMap[f.FlowFingerprint] = &ff
		} else {
			flow.InboundBytes += f.InboundBytes
			flow.InboundPackets += f.InboundPackets
			flow.InboundDuration += f.InboundDuration
			flow.OutboundBytes += f.OutboundBytes
			flow.OutboundPackets += f.OutboundPackets
			flow.OutboundDuration += f.OutboundDuration
		}
	}
	c.Mu.Unlock()
}

func (c *FlowCollection) Reset() {
	c.Mu.Lock()
	c.L3FlowMap = make(map[FlowFingerprint]*Flow, DefaultL3FlowCollectionSize)
	c.L4FlowMap = make(map[FlowFingerprint]*Flow, DefaultL4FlowCollectionSize)
	c.Mu.Unlock()
}

type FlowCollectionHistory struct {
	InterfaceName  string
	HistCollection map[FlowTimestamp]*FlowCollection
	LastTimestamp  FlowTimestamp
	Mu             *sync.Mutex
}

func NewFlowCollectionHistory(ifaceName string) (flowColHist *FlowCollectionHistory) {
	flowColHist = &FlowCollectionHistory{
		InterfaceName:  ifaceName,
		HistCollection: make(map[FlowTimestamp]*FlowCollection, DefaultFlowCollectionHistorySize),
		Mu:             &sync.Mutex{},
	}

	return
}

func (h *FlowCollectionHistory) SetLastTimestamp(ts FlowTimestamp) {
	h.LastTimestamp = ts
}

func (h *FlowCollectionHistory) Retention(before int64) {
	for k := range h.HistCollection {
		if k.End < before {
			delete(h.HistCollection, k)
		}
	}
}

/*
Assume duration = 5, flow timestamp list is aggregated as below:
10, 11, | 12, 13, 14, 15, 16, | 17, 18, 19, 20, 21, | 22, 23, 24, 25, 26(LastTimestamp.End)
*/
func (h *FlowCollectionHistory) AggregationByDuration(duration int64) (fc *FlowCollection) {
	fc = NewFlowCollection(h.InterfaceName)

	for ts := h.LastTimestamp; h.LastTimestamp.End-ts.End < duration; ts = ts.Offset(-1) {
		fcSample, ok := h.HistCollection[ts]
		if !ok {
			continue
		}

		fc.UpdateByFlowCol(fcSample)
	}

	return
}
