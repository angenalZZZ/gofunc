package shortid

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/angenalZZZ/gofunc/data/random"
)

// DefaultABC is the default URL-friendly alphabet.
const DefaultABC = "0129abcCDEijImk4567lnuJKMvLwxFGHyzABpqNrOPtoQsRSUdYefTghVW38ZX"

// Abc represents a shuffled alphabet used to generate the Ids and provides methods to
// encode data.
type Abc struct {
	alphabet []rune
}

// Id type represents a short Id generator working with a given alphabet.
type Id struct {
	abc    Abc
	worker uint
	epoch  time.Time  // ids can be generated for 34 years since this date
	ms     uint       // ms since epoch for the last id
	count  uint       // request count within the same ms
	mx     sync.Mutex // locks access to ms and count
}

var id *Id

func init() {
	id = MustNew(0, DefaultABC, 1)
}

// GetDefault retrieves the default short Id generator initialised with the default alphabet,
// worker=0 and seed=1. The default can be overwritten using SetDefault.
func GetDefault() *Id {
	return (*Id)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&id))))
}

// SetDefault overwrites the default generator.
func SetDefault(sid *Id) {
	target := (*unsafe.Pointer)(unsafe.Pointer(&id))
	source := unsafe.Pointer(sid)
	atomic.SwapPointer(target, source)
}

// Generate generates an Id using the default generator.
func Generate() (string, error) {
	return id.Generate()
}

// MustGenerate acts just like Generate, but panics instead of returning errors.
func MustGenerate() string {
	id, err := Generate()
	if err == nil {
		return id
	}
	panic(err)
}

// New constructs an instance of the short Id generator for the given worker number [0,31], alphabet
// (64 unique symbols) and seed value (to shuffle the alphabet). The worker number should be
// different for multiple or distributed processes generating Ids into the same data space. The
// seed, on contrary, should be identical.
func New(worker uint8, alphabet string, seed uint64) (*Id, error) {
	if worker > 31 {
		return nil, errors.New("expected worker in the range [0,31]")
	}
	abc, err := NewAbc(alphabet, seed)
	if err == nil {
		sid := &Id{
			abc:    abc,
			worker: uint(worker),
			epoch:  time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC),
			ms:     0,
			count:  0,
		}
		return sid, nil
	}
	return nil, err
}

// MustNew acts just like New, but panics instead of returning errors.
func MustNew(worker uint8, alphabet string, seed uint64) *Id {
	sid, err := New(worker, alphabet, seed)
	if err == nil {
		return sid
	}
	panic(err)
}

// Generate generates a new short Id.
func (sid *Id) Generate() (string, error) {
	return sid.GenerateInternal(nil, sid.epoch)
}

// MustGenerate acts just like Generate, but panics instead of returning errors.
func (sid *Id) MustGenerate() string {
	id, err := sid.Generate()
	if err == nil {
		return id
	}
	panic(err)
}

// GenerateInternal should only be used for testing purposes.
func (sid *Id) GenerateInternal(tm *time.Time, epoch time.Time) (string, error) {
	ms, count := sid.getMsAndCounter(tm, epoch)
	ids := make([]rune, 9)
	if tmp, err := sid.abc.Encode(ms, 8, 5); err == nil {
		copy(ids, tmp) // first 8 symbols
	} else {
		return "", err
	}
	if tmp, err := sid.abc.Encode(sid.worker, 1, 5); err == nil {
		ids[8] = tmp[0]
	} else {
		return "", err
	}
	if count > 0 {
		if s, err := sid.abc.Encode(count, 0, 6); err == nil {
			// only extend if really need it
			ids = append(ids, s...)
		} else {
			return "", err
		}
	}
	return string(ids), nil
}

func (sid *Id) getMsAndCounter(tm *time.Time, epoch time.Time) (uint, uint) {
	sid.mx.Lock()
	defer sid.mx.Unlock()
	var ms uint
	if tm != nil {
		ms = uint(tm.Sub(epoch).Nanoseconds() / 1000000)
	} else {
		ms = uint(time.Now().Sub(epoch).Nanoseconds() / 1000000)
	}
	if ms == sid.ms {
		sid.count++
	} else {
		sid.count = 0
		sid.ms = ms
	}
	return sid.ms, sid.count
}

// String returns a string representation of the short Id generator.
func (sid *Id) String() string {
	return fmt.Sprintf("Id(worker=%v, epoch=%v, abc=%v)", sid.worker, sid.epoch, sid.abc)
}

// Abc returns the instance of alphabet used for representing the Ids.
func (sid *Id) Abc() Abc {
	return sid.abc
}

// Epoch returns the value of epoch used as the beginning of millisecond counting (normally
// 2016-01-01 00:00:00 local time)
func (sid *Id) Epoch() time.Time {
	return sid.epoch
}

// Worker returns the value of worker for this short Id generator.
func (sid *Id) Worker() uint {
	return sid.worker
}

// NewAbc constructs a new instance of shuffled alphabet to be used for Id representation.
func NewAbc(alphabet string, seed uint64) (Abc, error) {
	runes := []rune(alphabet)
	if len(runes) != len(DefaultABC) {
		return Abc{}, fmt.Errorf("alphabet must contain %v unique characters", len(DefaultABC))
	}
	if nonUnique(runes) {
		return Abc{}, errors.New("alphabet must contain unique characters only")
	}
	abc := Abc{alphabet: nil}
	abc.shuffle(alphabet, seed)
	return abc, nil
}

// MustNewAbc acts just like NewAbc, but panics instead of returning errors.
func MustNewAbc(alphabet string, seed uint64) Abc {
	res, err := NewAbc(alphabet, seed)
	if err == nil {
		return res
	}
	panic(err)
}

func nonUnique(runes []rune) bool {
	found := make(map[rune]struct{})
	for _, r := range runes {
		if _, seen := found[r]; !seen {
			found[r] = struct{}{}
		}
	}
	return len(found) < len(runes)
}

func (abc *Abc) shuffle(alphabet string, seed uint64) {
	source := []rune(alphabet)
	for len(source) > 1 {
		seed = (seed*9301 + 49297) % 233280
		i := int(seed * uint64(len(source)) / 233280)

		abc.alphabet = append(abc.alphabet, source[i])
		source = append(source[:i], source[i+1:]...)
	}
	abc.alphabet = append(abc.alphabet, source[0])
}

// Encode encodes a given value into a slice of runes of length symbols. In case symbols==0, the
// length of the result is automatically computed from data. Even if fewer symbols is required to
// encode the data than symbols, all positions are used encoding 0 where required to guarantee
// uniqueness in case further data is added to the sequence. The value of digits [4,6] represents
// represents n in 2^n, which defines how much randomness flows into the algorithm: 4 -- every value
// can be represented by 4 symbols in the alphabet (permitting at most 16 values), 5 -- every value
// can be represented by 2 symbols in the alphabet (permitting at most 32 values), 6 -- every value
// is represented by exactly 1 symbol with no randomness (permitting 64 values).
func (abc *Abc) Encode(val, symbols, digits uint) ([]rune, error) {
	if digits < 4 || 6 < digits {
		return nil, fmt.Errorf("allowed digits range [4,6], found %v", digits)
	}

	var computedSize uint = 1
	if val >= 1 {
		computedSize = uint(math.Log2(float64(val)))/digits + 1
	}
	if symbols == 0 {
		symbols = computedSize
	} else if symbols < computedSize {
		return nil, fmt.Errorf("cannot accommodate data, need %v digits, got %v", computedSize, symbols)
	}

	mask := 1<<digits - 1

	rdm := make([]int, int(symbols))
	// no random component if digits == 6
	if digits < 6 {
		copy(rdm, maskedRandomInt(len(rdm), 0x3f-mask))
	}

	res := make([]rune, int(symbols))
	for i := range res {
		shift := digits * uint(i)
		index := (int(val>>shift) & mask) | rdm[i]
		res[i] = abc.alphabet[index]
	}
	return res, nil
}

// MustEncode acts just like Encode, but panics instead of returning errors.
func (abc *Abc) MustEncode(val, size, digits uint) []rune {
	res, err := abc.Encode(val, size, digits)
	if err == nil {
		return res
	}
	panic(err)
}

func maskedRandomInt(size, mask int) []int {
	s := make([]int, size)
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err == nil {
		for i, b := range bytes {
			s[i] = int(b) & mask
		}
	} else {
		for i := range s {
			s[i] = random.R.Intn(0xff) & mask
		}
	}
	return s
}

// String returns a string representation of the Abc instance.
func (abc Abc) String() string {
	return fmt.Sprintf("Abc{alphabet='%v')", abc.Alphabet())
}

// Alphabet returns the alphabet used as an immutable string.
func (abc Abc) Alphabet() string {
	return string(abc.alphabet)
}
