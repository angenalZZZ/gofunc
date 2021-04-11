package snowid

import (
	"sync"
	"time"
)

// SnowWorkerM1 .
type SnowWorkerM1 struct {
	BaseTime                int64  //基础时间
	WorkerId                uint16 //机器码
	WorkerIdBitLength       byte   //机器码位长
	SeqBitLength            byte   //自增序列数位长
	MaxSeqNumber            uint32 //最大序列数（含）
	MinSeqNumber            uint32 //最小序列数（含）
	TopOverCostCount        uint32 //最大漂移次数
	_TimestampShift         byte
	_CurrentSeqNumber       uint32
	_LastTimeTick           int64
	_TurnBackTimeTick       int64
	_TurnBackIndex          byte
	_IsOverCost             bool
	_OverCostCountInOneTerm uint32
	_GenCountInOneTerm      uint32
	_TermIndex              uint32
	sync.Mutex
}

// NewSnowWorkerM1 .
func NewSnowWorkerM1(options *IdGeneratorOptions) ISnowWorker {
	var workerIdBitLength byte
	var seqBitLength byte
	var maxSeqNumber uint32

	var workerId = options.WorkerId

	if options.WorkerIdBitLength == 0 {
		workerIdBitLength = 6
	} else {
		workerIdBitLength = options.WorkerIdBitLength
	}
	if options.SeqBitLength == 0 {
		seqBitLength = 6
	} else {
		seqBitLength = options.SeqBitLength
	}
	if options.MaxSeqNumber > 0 {
		maxSeqNumber = options.MaxSeqNumber
	} else {
		maxSeqNumber = (1 << seqBitLength) - 1
	}
	var minSeqNumber = options.MinSeqNumber
	var topOverCostCount = options.TopOverCostCount

	var baseTime int64
	if options.BaseTime != 0 {
		baseTime = options.BaseTime
	} else {
		baseTime = 1582136402000
	}

	timestampShift := (byte)(options.WorkerIdBitLength + options.SeqBitLength)
	currentSeqNumber := options.MinSeqNumber

	return &SnowWorkerM1{
		BaseTime:          baseTime,
		WorkerId:          workerId,
		WorkerIdBitLength: workerIdBitLength,
		SeqBitLength:      seqBitLength,
		MaxSeqNumber:      maxSeqNumber,
		MinSeqNumber:      minSeqNumber,
		TopOverCostCount:  topOverCostCount,
		_TimestampShift:   timestampShift,
		_CurrentSeqNumber: currentSeqNumber}
}

// DoGenIDAction .
func (m1 *SnowWorkerM1) DoGenIDAction(arg *OverCostActionArg) {

}

// BeginOverCostAction .
func (m1 *SnowWorkerM1) BeginOverCostAction(useTimeTick int64) {

}

// EndOverCostAction .
func (m1 *SnowWorkerM1) EndOverCostAction(useTimeTick int64) {
	if m1._TermIndex > 10000 {
		m1._TermIndex = 0
	}
}

// BeginTurnBackAction .
func (m1 *SnowWorkerM1) BeginTurnBackAction(useTimeTick int64) {

}

// EndTurnBackAction .
func (m1 *SnowWorkerM1) EndTurnBackAction(useTimeTick int64) {

}

// NextOverCostID .
func (m1 *SnowWorkerM1) NextOverCostID() uint64 {
	currentTimeTick := m1.GetCurrentTimeTick()
	if currentTimeTick > m1._LastTimeTick {
		m1.EndOverCostAction(currentTimeTick)
		m1._LastTimeTick = currentTimeTick
		m1._CurrentSeqNumber = m1.MinSeqNumber
		m1._IsOverCost = false
		m1._OverCostCountInOneTerm = 0
		m1._GenCountInOneTerm = 0
		return m1.CalcID(m1._LastTimeTick)
	}
	if m1._OverCostCountInOneTerm >= m1.TopOverCostCount {
		m1.EndOverCostAction(currentTimeTick)
		m1._LastTimeTick = m1.GetNextTimeTick()
		m1._CurrentSeqNumber = m1.MinSeqNumber
		m1._IsOverCost = false
		m1._OverCostCountInOneTerm = 0
		m1._GenCountInOneTerm = 0
		return m1.CalcID(m1._LastTimeTick)
	}
	if m1._CurrentSeqNumber > m1.MaxSeqNumber {
		m1._LastTimeTick++
		m1._CurrentSeqNumber = m1.MinSeqNumber
		m1._IsOverCost = true
		m1._OverCostCountInOneTerm++
		m1._GenCountInOneTerm++

		return m1.CalcID(m1._LastTimeTick)
	}

	m1._GenCountInOneTerm++
	return m1.CalcID(m1._LastTimeTick)
}

// NextNormalID .
func (m1 *SnowWorkerM1) NextNormalID() uint64 {
	currentTimeTick := m1.GetCurrentTimeTick()
	if currentTimeTick < m1._LastTimeTick {
		if m1._TurnBackTimeTick < 1 {
			m1._TurnBackTimeTick = m1._LastTimeTick - 1
			m1._TurnBackIndex++
			// 每毫秒序列数的前5位是预留位，0用于手工新值，1-4是时间回拨次序
			// 最多4次回拨（防止回拨重叠）
			if m1._TurnBackIndex > 4 {
				m1._TurnBackIndex = 1
			}
			m1.BeginTurnBackAction(m1._TurnBackTimeTick)
		}

		time.Sleep(time.Duration(1) * time.Millisecond)
		return m1.CalcTurnBackID(m1._TurnBackTimeTick)
	}
	// 时间追平时，_TurnBackTimeTick清零
	if m1._TurnBackTimeTick > 0 {
		m1.EndTurnBackAction(m1._TurnBackTimeTick)
		m1._TurnBackTimeTick = 0
	}
	if currentTimeTick > m1._LastTimeTick {
		m1._LastTimeTick = currentTimeTick
		m1._CurrentSeqNumber = m1.MinSeqNumber
		return m1.CalcID(m1._LastTimeTick)
	}
	if m1._CurrentSeqNumber > m1.MaxSeqNumber {
		m1.BeginOverCostAction(currentTimeTick)
		m1._TermIndex++
		m1._LastTimeTick++
		m1._CurrentSeqNumber = m1.MinSeqNumber
		m1._IsOverCost = true
		m1._OverCostCountInOneTerm = 1
		m1._GenCountInOneTerm = 1

		return m1.CalcID(m1._LastTimeTick)
	}

	return m1.CalcID(m1._LastTimeTick)
}

// CalcID .
func (m1 *SnowWorkerM1) CalcID(useTimeTick int64) uint64 {
	result := uint64(useTimeTick<<m1._TimestampShift) + uint64(m1.WorkerId<<m1.SeqBitLength) + uint64(m1._CurrentSeqNumber)
	m1._CurrentSeqNumber++
	return result
}

// CalcTurnBackID .
func (m1 *SnowWorkerM1) CalcTurnBackID(useTimeTick int64) uint64 {
	result := uint64(useTimeTick<<m1._TimestampShift) + uint64(m1.WorkerId<<m1.SeqBitLength) + uint64(m1._TurnBackIndex)
	m1._TurnBackTimeTick--
	return result
}

// GetCurrentTimeTick .
func (m1 *SnowWorkerM1) GetCurrentTimeTick() int64 {
	var millis = time.Now().UnixNano() / 1e6
	return millis - m1.BaseTime
}

// GetNextTimeTick .
func (m1 *SnowWorkerM1) GetNextTimeTick() int64 {
	tempTimeTicker := m1.GetCurrentTimeTick()
	for tempTimeTicker <= m1._LastTimeTick {
		tempTimeTicker = m1.GetCurrentTimeTick()
	}
	return tempTimeTicker
}

// NextId .
func (m1 *SnowWorkerM1) NextId() uint64 {
	m1.Lock()
	defer m1.Unlock()
	if m1._IsOverCost {
		return m1.NextOverCostID()
	}
	return m1.NextNormalID()
}
