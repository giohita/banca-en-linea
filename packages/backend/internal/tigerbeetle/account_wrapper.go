//go:build !ci && !docker

package tigerbeetle

import "github.com/tigerbeetle/tigerbeetle-go/pkg/types"

// AccountWrapper envuelve types.Account para implementar AccountInterface
type AccountWrapper struct {
	*types.Account
}

// Implementación de AccountInterface para AccountWrapper
func (a *AccountWrapper) GetID() uint64            { return a.ID }
func (a *AccountWrapper) GetLedger() uint32        { return a.Ledger }
func (a *AccountWrapper) GetCode() uint16          { return a.Code }
func (a *AccountWrapper) GetFlags() uint16         { return a.Flags.ToUint16() }
func (a *AccountWrapper) GetDebitsPosted() uint64  { return a.DebitsPosted }
func (a *AccountWrapper) GetCreditsPosted() uint64 { return a.CreditsPosted }