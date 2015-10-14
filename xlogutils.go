package xlogutils
import (
	"fmt"
	"errors"
)

const XLOG_SEG_SIZE_16MB = 16 * 1024 * 1024
const XLOG_SEG_SIZE_DEFAULT = XLOG_SEG_SIZE_16MB

var xlog_seg_size uint = XLOG_SEG_SIZE_DEFAULT
var XLogSegmentsPerXLogId uint

var ParseError = errors.New("Parse error")

func init() {
	SetXLogSegSize(XLOG_SEG_SIZE_16MB)
}

func XLogSegSize() (uint) {
	return xlog_seg_size
}

func SetXLogSegSize(size uint) {
	xlog_seg_size = size
	XLogSegmentsPerXLogId = 0x100000000 / xlog_seg_size
}

type Location struct {
	XLogId uint
	Offset uint
}

type Filename struct {
	Timeline uint
	XLogId uint
	Segment uint
	IsHistory bool
}

func FilenameFromString(fname string) (*Filename, error) {
	ret := Filename{}
	if r, err := fmt.Sscanf(fname, "%08X%08X%08X", &ret.Timeline, &ret.XLogId, &ret.Segment); r != 3 || err != nil {
		if r, err := fmt.Sscan("%08X.history", &ret.Timeline); r == 1 && err == nil {
			ret.Segment = 0
			ret.XLogId = 0
			ret.IsHistory = true
			return &ret, nil
		} else {
			return nil, ParseError
		}
	}
	return &ret, nil
}

func LocationFromString(loc string) (*Location, error) {
	ret := Location{}
	if r, err := fmt.Sscanf(loc, "%X/%X", &ret.XLogId, &ret.Offset); r != 2 || err != nil {
		return nil, ParseError
	}

	return &ret, nil
}

func (this *Location) Filename(tli uint) *Filename {
	ret := Filename{}
	ret.Timeline = tli
	ret.XLogId = this.XLogId
	ret.Segment = this.Offset / xlog_seg_size
	return &ret
}

func (this *Location) Int() uint64 {
	return uint64(this.XLogId) * 0x100000000 + uint64(this.Offset)
}

func (this *Location) Diff(other *Location) uint64 {
	if this.Int() > other.Int() {
		return this.Int() - other.Int()
	} else {
		return other.Int() - this.Int()
	}
}

func (this *Filename) String() string {
	if (!this.IsHistory) {
		return fmt.Sprintf("%08x%08X%08X", this.Timeline, this.XLogId, this.Segment)
	} else {
		return fmt.Sprintf("%08X.history", this.Timeline)
	}
}

func (this *Filename) Location() *Location {
	ret := Location{}
	ret.XLogId = this.XLogId
	ret.Offset = this.Segment * xlog_seg_size
	return &ret
}

func (this *Filename) Next() *Filename {
	ret := Filename{}
	ret.Timeline = this.Timeline
	ret.XLogId = this.XLogId
	ret.Segment = this.Segment
	ret.Segment++
	if ret.Segment == XLogSegmentsPerXLogId {
		ret.Segment = 0
		ret.XLogId++
	}
	return &ret
}

func (this *Filename) Prev() *Filename {
	ret := Filename{}
	ret.Timeline = this.Timeline
	ret.XLogId = this.XLogId
	ret.Segment = this.Segment
	if ret.Segment == 0 {
		ret.Segment = XLogSegmentsPerXLogId
		ret.XLogId--
	}
	ret.Segment--
	return &ret
}