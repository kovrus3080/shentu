package gov

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/stretchr/testify/require"

	"github.com/certikfoundation/shentu/common"
	"github.com/certikfoundation/shentu/x/gov/internal/keeper"
	"github.com/certikfoundation/shentu/x/gov/internal/types"
)

func TestKeeper_ProposeAndVote(t *testing.T) {
	t.Log("Test keeper AddVote")
	input := createTestInput(t)
	err := input.bankKeeper.SetCoins(input.ctx, addrs[1], sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 80000*1e6)))
	require.Equal(t, nil, err)

	tp := gov.TextProposal{Title: "title0", Description: "desc0"}
	t.Run("Test submitting a proposal and adding a vote with yes", func(t *testing.T) {
		pp, err := input.govKeeper.SubmitProposal(input.ctx, tp, addrs[0])
		require.Equal(t, nil, err)
		vote := govTypes.NewVote(pp.ProposalID, addrs[0], govTypes.OptionYes)

		coins700 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 700*1e6))
		votingPeriodActivated, err := input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[1], coins700)
		require.Equal(t, nil, err)
		require.Equal(t, true, votingPeriodActivated)

		err = input.govKeeper.AddVote(input.ctx, vote.ProposalID, vote.Voter, vote.Option)
		require.Equal(t, nil, err)

		// the vote does not count since addr[0] is not a validator
		results := map[govTypes.VoteOption]sdk.Dec{
			govTypes.OptionYes:        sdk.ZeroDec(),
			govTypes.OptionAbstain:    sdk.ZeroDec(),
			govTypes.OptionNo:         sdk.ZeroDec(),
			govTypes.OptionNoWithVeto: sdk.ZeroDec(),
		}

		pass, veto, res := keeper.Tally(input.ctx, input.govKeeper, pp)
		require.Equal(t, false, pass)
		require.Equal(t, false, veto)
		require.Equal(t, gov.NewTallyResultFromMap(results), res)
	})

	// TODO: more tests. validator cases
}
func TestKeeper_AddDeposit(t *testing.T) {
	t.Log("Test keeper AddDeposit")
	input := createTestInput(t)
	err := input.bankKeeper.SetCoins(input.ctx, addrs[1], sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 80000*1e6)))
	require.Equal(t, nil, err)

	tp := gov.TextProposal{Title: "title0", Description: "desc0"}

	t.Run("adding deposit and proposal doesn't exist", func(t *testing.T) {
		pp, err := input.govKeeper.SubmitProposal(input.ctx, tp, addrs[0])
		require.Equal(t, nil, err)
		coins100 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 100*1e6))

		votingPeriodActivated, err := input.govKeeper.AddDeposit(input.ctx, pp.ProposalID+1, addrs[1], coins100)
		errString := fmt.Sprintf("unknown proposal: %d", pp.ProposalID+1)
		require.EqualError(t, err, errString)
		require.Equal(t, false, votingPeriodActivated)
	})

	t.Run("adding deposit not enough balance", func(t *testing.T) {
		pp, err := input.govKeeper.SubmitProposal(input.ctx, tp, addrs[0])
		require.Equal(t, nil, err)
		coins15000 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 15000*1e6))

		votingPeriodActivated, err := input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[0], coins15000)
		errString := "insufficient funds: insufficient account funds; 10000uctk < 15000000000uctk"
		require.EqualError(t, err, errString)
		require.Equal(t, false, votingPeriodActivated)
	})

	t.Run("adding deposit and waiting for more deposits", func(t *testing.T) {
		pp, err := input.govKeeper.SubmitProposal(input.ctx, tp, addrs[0])
		require.Equal(t, nil, err)
		coins100 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 100*1e6))

		votingPeriodActivated, err := input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[1], coins100)
		require.Equal(t, nil, err)
		require.Equal(t, false, votingPeriodActivated)
	})

	t.Run("adding more deposit and still waiting for more", func(t *testing.T) {
		pp, err := input.govKeeper.SubmitProposal(input.ctx, tp, addrs[0])
		require.Equal(t, nil, err)
		coins100 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 100*1e6))
		coins200 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 200*1e6))

		votingPeriodActivated, err := input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[1], coins100)
		require.Equal(t, nil, err)
		require.Equal(t, false, votingPeriodActivated)

		votingPeriodActivated, err = input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[1], coins200)
		require.Equal(t, nil, err)
		require.Equal(t, false, votingPeriodActivated)
	})

	t.Run("adding deposit and entering votingPeriod", func(t *testing.T) {
		pp, err := input.govKeeper.SubmitProposal(input.ctx, tp, addrs[0])
		require.Equal(t, nil, err)
		coins700 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 700*1e6))

		votingPeriodActivated, err := input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[1], coins700)
		require.Equal(t, nil, err)
		require.Equal(t, true, votingPeriodActivated)
	})

	t.Run("entering votingPeriod and trying to add more deposit", func(t *testing.T) {
		pp, err := input.govKeeper.SubmitProposal(input.ctx, tp, addrs[0])
		require.Equal(t, nil, err)
		coins700 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 700*1e6))
		coinsAfterAvtivated := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 1))

		votingPeriodActivated, err := input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[1], coins700)
		require.Equal(t, nil, err)
		require.Equal(t, true, votingPeriodActivated)

		votingPeriodActivated, err = input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[1], coinsAfterAvtivated)
		errString := fmt.Sprintf("proposal already active: %d", pp.ProposalID)
		require.EqualError(t, err, errString)
		require.Equal(t, false, votingPeriodActivated)
	})
}

func TestKeeper_DepositOperation(t *testing.T) {
	t.Log("Test keeper DepositOperation")
	input := createTestInput(t)

	tp := gov.TextProposal{Title: "title0", Description: "desc0"}
	err := input.bankKeeper.SetCoins(input.ctx, addrs[1], sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 80000*1e6)))
	require.Equal(t, nil, err)
	err = input.bankKeeper.SetCoins(input.ctx, addrs[2], sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 80000*1e6)))
	require.Equal(t, nil, err)
	err = input.bankKeeper.SetCoins(input.ctx, addrs[3], sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 80000*1e6)))
	require.Equal(t, nil, err)

	t.Run("refund all deposits in a specific proposal", func(t *testing.T) {
		pp, err := input.govKeeper.SubmitProposal(input.ctx, tp, addrs[0])
		require.Equal(t, nil, err)
		coins100 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 100*1e6))
		coins50 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 50*1e6))
		coins20 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 20*1e6))

		_, _ = input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[1], coins100)
		_, _ = input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[2], coins50)
		votingPeriodActivated, err := input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[3], coins20)
		require.Equal(t, nil, err)
		require.Equal(t, false, votingPeriodActivated)

		addr1Amount := input.bankKeeper.GetCoins(input.ctx, addrs[1])
		addr2Amount := input.bankKeeper.GetCoins(input.ctx, addrs[2])
		addr3Amount := input.bankKeeper.GetCoins(input.ctx, addrs[3])
		require.Equal(t, sdk.NewInt(79900*1e6).Int64(), addr1Amount.AmountOf(common.MicroCTKDenom).Int64())
		require.Equal(t, sdk.NewInt(79950*1e6).Int64(), addr2Amount.AmountOf(common.MicroCTKDenom).Int64())
		require.Equal(t, sdk.NewInt(79980*1e6).Int64(), addr3Amount.AmountOf(common.MicroCTKDenom).Int64())

		input.govKeeper.RefundDepositsByProposalID(input.ctx, pp.ProposalID)
		depositsRemaining := input.govKeeper.GetAllDeposits(input.ctx)
		require.Equal(t, types.Deposits(nil), depositsRemaining)
		addr1Amount = input.bankKeeper.GetCoins(input.ctx, addrs[1])
		addr2Amount = input.bankKeeper.GetCoins(input.ctx, addrs[2])
		addr3Amount = input.bankKeeper.GetCoins(input.ctx, addrs[3])
		require.Equal(t, sdk.NewInt(80000*1e6).Int64(), addr1Amount.AmountOf(common.MicroCTKDenom).Int64())
		require.Equal(t, sdk.NewInt(80000*1e6).Int64(), addr2Amount.AmountOf(common.MicroCTKDenom).Int64())
		require.Equal(t, sdk.NewInt(80000*1e6).Int64(), addr3Amount.AmountOf(common.MicroCTKDenom).Int64())
	})
	t.Run("delete all deposits in a specific proposal", func(t *testing.T) {
		pp, err := input.govKeeper.SubmitProposal(input.ctx, tp, addrs[0])
		require.Equal(t, nil, err)
		coins10 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 10*1e6))
		coins50 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 50*1e6))
		coins20 := sdk.NewCoins(sdk.NewInt64Coin(common.MicroCTKDenom, 20*1e6))

		_, _ = input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[1], coins10)
		_, _ = input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[2], coins20)
		votingPeriodActivated, err := input.govKeeper.AddDeposit(input.ctx, pp.ProposalID, addrs[3], coins50)
		require.Equal(t, nil, err)
		require.Equal(t, false, votingPeriodActivated)

		addr1Amount := input.bankKeeper.GetCoins(input.ctx, addrs[1])
		addr2Amount := input.bankKeeper.GetCoins(input.ctx, addrs[2])
		addr3Amount := input.bankKeeper.GetCoins(input.ctx, addrs[3])
		require.Equal(t, sdk.NewInt(79990*1e6).Int64(), addr1Amount.AmountOf(common.MicroCTKDenom).Int64())
		require.Equal(t, sdk.NewInt(79980*1e6).Int64(), addr2Amount.AmountOf(common.MicroCTKDenom).Int64())
		require.Equal(t, sdk.NewInt(79950*1e6).Int64(), addr3Amount.AmountOf(common.MicroCTKDenom).Int64())

		input.govKeeper.DeleteDepositsByProposalID(input.ctx, pp.ProposalID)
		depositsRemaining := input.govKeeper.GetAllDeposits(input.ctx)
		require.Equal(t, types.Deposits(nil), depositsRemaining)

		addr1Amount = input.bankKeeper.GetCoins(input.ctx, addrs[1])
		addr2Amount = input.bankKeeper.GetCoins(input.ctx, addrs[2])
		addr3Amount = input.bankKeeper.GetCoins(input.ctx, addrs[3])
		require.Equal(t, sdk.NewInt(79990*1e6).Int64(), addr1Amount.AmountOf(common.MicroCTKDenom).Int64())
		require.Equal(t, sdk.NewInt(79980*1e6).Int64(), addr2Amount.AmountOf(common.MicroCTKDenom).Int64())
		require.Equal(t, sdk.NewInt(79950*1e6).Int64(), addr3Amount.AmountOf(common.MicroCTKDenom).Int64())
	})
}
