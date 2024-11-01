package xhttp

type options struct {
	recordSize uint32 // 最大4MB
}

func defaultDoOptions() options {
	return options{
		recordSize: 512,
	}
}

type IDoOptions interface {
	apply(*options)
}

type DoOptionsWithRecordSize struct {
	size uint32
}

func NewDoOptionsWithRecordSize(size uint32) *DoOptionsWithRecordSize {
	return &DoOptionsWithRecordSize{size: size}
}

func (x *DoOptionsWithRecordSize) apply(o *options) {
	o.recordSize = x.size
}
