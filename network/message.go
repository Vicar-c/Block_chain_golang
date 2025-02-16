package network

type GetBlocksMessage struct {
	From uint32
	// To 为0时需要拿到所有的blocks
	To uint32
}

type GetStatusMessage struct {
}

type StatusMessage struct {
	ID            string
	Version       uint32
	CurrentHeight uint32
}
