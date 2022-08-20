package fani

import (
	"github.com/ipfs/go-cid"
)

type ArgType struct {
	Name  string
	Codec cid.Cid
}

type FnABI struct {
	ID       string
	ByteCode cid.Cid
	Args     []ArgType
}
