package ksm

// "phase":         e.Phase,
// "extrinsic_idx": e.ExtrinsicIdx,
// "type":          e.Type,
// "module_id":     call.Module.Name,
// "event_id":      e.Event.Name,
// "params":        e.Params,
// "topic":         e.Topic,
type EventModel struct {
	Phase        interface{}
	ExtrinsicIdx interface{}
	Type         interface{}
	ModuleId     interface{}
	EventId      interface{}
	Params       interface{}
	Topics       interface{}
}

func (e EventModel) GetPhaseApplyExtrinsic() int {
	return e.Phase.(int)
}

func (e EventModel) GetExtrinsicIdx() int {
	return e.ExtrinsicIdx.(int)
}

func (e EventModel) GetEventId() string {
	return e.EventId.(string)
}

type EventModelRecord []EventModel

func (e EventModelRecord) GetBalancesTransfer() EventModelRecord {
	er := make(EventModelRecord, 0)
	for _, v := range e {
		if v.ModuleId == "Balances" && v.EventId == "Transfer" {
			er = append(er, v)
		}
	}
	return er
}

func (e EventModelRecord) GetSystemExtrinsicFailed() EventModelRecord {
	er := make(EventModelRecord, 0)
	for _, v := range e {
		if v.ModuleId == "System" && v.EventId == "ExtrinsicFailed" {
			er = append(er, v)
		}
	}
	return er
}
