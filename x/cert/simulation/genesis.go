package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/certikfoundation/shentu/x/cert/internal/types"
)

// RandomizedGenState creates a random genesis state for module simulation.
func RandomizedGenState(simState *module.SimulationState) {
	r := simState.Rand
	gs := types.GenesisState{}

	numOfCertifiers := r.Intn(100)
	for i := 0; i < numOfCertifiers; i++ {
		gs.Certifiers = append(gs.Certifiers, GenerateACertifier(r))
	}

	numOfValidators := r.Intn(10)
	for i := 0; i < numOfValidators; i++ {
		gs.Validators = append(gs.Validators, GenerateAValidator(r))
	}

	numOfPlatforms := r.Intn(10)
	for i := 0; i < numOfPlatforms; i++ {
		gs.Platforms = append(gs.Platforms, GenerateAPlatform(r))
	}

	numOfLibrary := r.Intn(20)
	for i := 0; i < numOfLibrary; i++ {
		gs.Libraries = append(gs.Libraries, GenerateALibrary(r))
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(gs)
}

// GenerateACertifier returns an object of Certifier with random field values.
func GenerateACertifier(r *rand.Rand) types.Certifier {
	return types.Certifier{
		Address:     simulation.RandomAccounts(r, 1)[0].Address,
		Proposer:    simulation.RandomAccounts(r, 1)[0].Address,
		Description: simulation.RandStringOfLength(r, 50),
	}
}

// GenerateAValidator returns an object of Validator with random field values.
func GenerateAValidator(r *rand.Rand) types.Validator {
	randomAccount := simulation.RandomAccounts(r, 1)[0]
	return types.Validator{
		PubKey:    randomAccount.PubKey,
		Certifier: randomAccount.Address,
	}
}

// GenerateALibrary returns an object of Library with random field values.
func GenerateALibrary(r *rand.Rand) types.Library {
	return types.Library{
		Address:   simulation.RandomAccounts(r, 1)[0].Address,
		Publisher: simulation.RandomAccounts(r, 1)[0].Address,
	}
}

// GenerateAPlatform returns an object of Platform with random field values.
func GenerateAPlatform(r *rand.Rand) types.Platform {
	return types.Platform{
		Address:     sdk.GetConsAddress(simulation.RandomAccounts(r, 1)[0].PubKey),
		Description: simulation.RandStringOfLength(r, 10),
	}
}
