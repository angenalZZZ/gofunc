package snowid

// OverCostActionArg .
type OverCostActionArg struct {
	ActionType             int32
	TimeTick               int64
	WorkerId               uint16
	OverCostCountInOneTerm int32
	GenCountInOneTerm      int32
	TermIndex              int32
}

// OverCostActionArg .
func (ora OverCostActionArg) OverCostActionArg(workerId uint16, timeTick int64, actionType int32, overCostCountInOneTerm int32, genCountWhenOverCost int32, index int32) {
	ora.ActionType = actionType
	ora.TimeTick = timeTick
	ora.WorkerId = workerId
	ora.OverCostCountInOneTerm = overCostCountInOneTerm
	ora.GenCountInOneTerm = genCountWhenOverCost
	ora.TermIndex = index
}
