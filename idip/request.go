package idip

type RequestHead struct {
	GameAppId string
	BigArea   string
	Area      string
	Cmd       int
}

type RequestHeadOption func(*RequestHead)

func NewRequestHead(opts ...RequestHeadOption) *RequestHead {
	head := &RequestHead{}
	for _, opt := range opts {
		opt(head)
	}
	return head
}

// WithGameAppId sets the GameAppId field
func WithGameAppId(gameAppId string) RequestHeadOption {
	return func(r *RequestHead) {
		r.GameAppId = gameAppId
	}
}

// WithBigArea sets the BigArea field
func WithBigArea(bigArea string) RequestHeadOption {
	return func(r *RequestHead) {
		r.BigArea = bigArea
	}
}

// WithArea sets the Area field
func WithArea(area string) RequestHeadOption {
	return func(r *RequestHead) {
		r.Area = area
	}
}

// WithCmd sets the Cmd field
func WithCmd(cmd int) RequestHeadOption {
	return func(r *RequestHead) {
		r.Cmd = cmd
	}
}
