/*
* 薄雾算法 github.com/asyncins/mist
*
* 1      2                                                     48         56       64
* +------+-----------------------------------------------------+----------+----------+
* retain | incr                                                | salt     | sequence |
* +------+-----------------------------------------------------+----------+----------+
* 0      | 0000000000 0000000000 0000000000 0000000000 0000000 | 00000000 | 00000000 |
* +------+-----------------------------------------------------+------------+--------+
*
* 0. 最高位，占 1 位，保持为 0，使得值永远为正数；
* 1. 自增数，占 47 位，自增数在高位能保证结果值呈递增态势，遂低位可以为所欲为；
* 2. 随机因子一，占 8 位，上限数值 255，使结果值不可预测；
* 3. 随机因子二，占 8 位，上限数值 255，使结果值不可预测；
*
* 编号上限为百万亿级，上限值计算为 140737488355327 即 int64(1 << 47 - 1)，假设每天取值 10 亿，能使用 385+ 年
 */
package random

import (
	"crypto/rand"
	"math/big"
	"sync"
)

const saltBit = uint(8)               // 随机因子二进制位数
const saltShift = uint(8)             // 随机因子移位数
const incrShift = saltBit + saltShift // 自增数移位数
var defaultMist *Mist

type Mist struct {
	sync.Mutex        // 互斥锁
	incr       uint64 // 自增数
	saltA      uint64 // 随机因子一
	saltB      uint64 // 随机因子二
}

func init() {
	defaultMist = NewMist()
}

func Int64() uint64 {
	return defaultMist.Generate()
}

func NewMist(incr ...uint64) *Mist {
	var i uint64 = 1
	if len(incr) > 0 && incr[0] > 0 {
		i = incr[0]
	}
	return &Mist{incr: i}
}

func (c *Mist) Generate() uint64 {
	c.Lock()
	c.incr++
	// 获取随机因子数值 ｜ 使用真随机函数提高性能
	randA, _ := rand.Int(rand.Reader, big.NewInt(255))
	c.saltA = randA.Uint64()
	randB, _ := rand.Int(rand.Reader, big.NewInt(255))
	c.saltB = randB.Uint64()
	// 实现自动占位
	mist := (c.incr << incrShift) | (c.saltA << saltShift) | c.saltB
	c.Unlock()
	return mist
}
