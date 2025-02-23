package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {

	randomAccount1 := createRandomAccount(t)
	randomAccount2 := createRandomAccount(t)
	fmt.Println(">> before:", randomAccount1.Balance, randomAccount2.Balance)

	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			ctx := context.Background()
			result, err := testStore.TransferTx(ctx, TransferTxParams{
				FromAccountID: randomAccount1.ID,
				ToAccountID:   randomAccount2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	// check results
	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, randomAccount1.ID, transfer.FromAccountID)
		require.Equal(t, randomAccount2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = testStore.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, randomAccount1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = testStore.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, randomAccount2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = testStore.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check account
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, randomAccount1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, randomAccount2.ID, toAccount.ID)

		// check accounts balance
		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)

		diff1 := randomAccount1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - randomAccount2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balances
	updateAccount1, err := testStore.GetAccount(context.Background(), randomAccount1.ID)
	require.NoError(t, err)

	updateAccount2, err := testStore.GetAccount(context.Background(), randomAccount2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updateAccount1.Balance, updateAccount2.Balance)

	require.Equal(t, randomAccount1.Balance-int64(n)*amount, updateAccount1.Balance)
	require.Equal(t, randomAccount2.Balance+int64(n)*amount, updateAccount2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {

	randomAccount1 := createRandomAccount(t)
	randomAccount2 := createRandomAccount(t)
	fmt.Println(">> before:", randomAccount1.Balance, randomAccount2.Balance)

	// run n concurrent transfer transactions
	n := 10
	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := randomAccount1.ID
		toAccountID := randomAccount2.ID

		if i%2 == 1 {
			fromAccountID = randomAccount2.ID
			toAccountID = randomAccount1.ID
		}

		go func() {
			ctx := context.Background()
			_, err := testStore.TransferTx(ctx, TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

	}

	// check the final updated balances
	updateAccount1, err := testStore.GetAccount(context.Background(), randomAccount1.ID)
	require.NoError(t, err)

	updateAccount2, err := testStore.GetAccount(context.Background(), randomAccount2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updateAccount1.Balance, updateAccount2.Balance)

	require.Equal(t, randomAccount1.Balance, updateAccount1.Balance)
	require.Equal(t, randomAccount2.Balance, updateAccount2.Balance)
}
