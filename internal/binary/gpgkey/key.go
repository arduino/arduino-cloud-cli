package gpgkey

import (
	_ "embed"
)

var (
	//go:embed index_public.gpg.key
	IndexPublicKey []byte
)
