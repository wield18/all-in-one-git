package videobuf

type vBufOption func(*VideoBufServer)

func WithMaxSize(maxSize int) vBufOption {
	return vBufOption(func(vbs *VideoBufServer) {
		vbs.maxSize = maxSize
	})
}
