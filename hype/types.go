package hype

import "hype-proxy/hype/client"

type RollForwardHandle func(block []*client.BlockHeader) bool
type RollTransferHandle func(block []*client.Transfer) bool
