package memwallet

import (
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1"
	"github.com/decred/dcrd/dcrutil"
)
// keyToAddr maps the passed private to corresponding p2pkh address.
func keyToAddr(key *secp256k1.PrivateKey, net *chaincfg.Params) (dcrutil.Address, error) {
	pubKey := (*secp256k1.PublicKey)(&key.PublicKey)
	serializedKey := pubKey.SerializeCompressed()
	pubKeyAddr, err := dcrutil.NewAddressSecpPubKey(serializedKey, net)
	if err != nil {
		return nil, err
	}
	return pubKeyAddr.AddressPubKeyHash(), nil
}