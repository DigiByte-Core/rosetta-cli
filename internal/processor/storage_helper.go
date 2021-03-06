// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package processor

import (
	"context"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/fetcher"
	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/reconciler"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// BlockStorageHelper implements the storage.Helper
// interface.
type BlockStorageHelper struct {
	network *types.NetworkIdentifier
	fetcher *fetcher.Fetcher

	// Configuration settings
	lookupBalanceByBlock bool
	exemptAccounts       []*reconciler.AccountCurrency
}

// NewBlockStorageHelper returns a new BlockStorageHelper.
func NewBlockStorageHelper(
	network *types.NetworkIdentifier,
	fetcher *fetcher.Fetcher,
	lookupBalanceByBlock bool,
	exemptAccounts []*reconciler.AccountCurrency,
) *BlockStorageHelper {
	return &BlockStorageHelper{
		network:              network,
		fetcher:              fetcher,
		lookupBalanceByBlock: lookupBalanceByBlock,
		exemptAccounts:       exemptAccounts,
	}
}

// AccountBalance attempts to fetch the balance
// for a missing account in storage. This is necessary
// for running the "check" command at an arbitrary height
// instead of syncing from genesis.
func (h *BlockStorageHelper) AccountBalance(
	ctx context.Context,
	account *types.AccountIdentifier,
	currency *types.Currency,
	block *types.BlockIdentifier,
) (*types.Amount, error) {
	if !h.lookupBalanceByBlock {
		return &types.Amount{
			Value:    "0",
			Currency: currency,
		}, nil
	}

	// In the case that we are syncing from arbitrary height,
	// we may need to recover the balance of an account to
	// perform validations.
	_, value, err := reconciler.GetCurrencyBalance(
		ctx,
		h.fetcher,
		h.network,
		account,
		currency,
		types.ConstructPartialBlockIdentifier(block),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to get currency balance in storage helper", err)
	}

	return &types.Amount{
		Value:    value,
		Currency: currency,
	}, nil
}

// Asserter returns a *asserter.Asserter.
func (h *BlockStorageHelper) Asserter() *asserter.Asserter {
	return h.fetcher.Asserter
}

// ExemptFunc returns a parser.ExemptOperation.
func (h *BlockStorageHelper) ExemptFunc() parser.ExemptOperation {
	return func(op *types.Operation) bool {
		return reconciler.ContainsAccountCurrency(
			h.exemptAccounts,
			&reconciler.AccountCurrency{
				Account:  op.Account,
				Currency: op.Amount.Currency,
			},
		)
	}
}
