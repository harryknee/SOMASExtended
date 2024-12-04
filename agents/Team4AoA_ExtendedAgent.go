package agents

/* Contains functions relevant to agent functionality when the AoA is Team4AoA */

import (
	"log"
	"math/rand"

	"github.com/google/uuid"

	common "github.com/ADimoska/SOMASExtended/common"
)

// ----------------------- Team 4 AoA Functions -----------------------

func (mi *ExtendedAgent) Team4_GetRankUpVote() map[uuid.UUID]int {
	return make(map[uuid.UUID]int)
}

func (mi *ExtendedAgent) Team4_GetConfession() bool {
	return false
}

func (mi *ExtendedAgent) Team4_StateConfessionToTeam() {
	// Broadcast contribution to team
	confession := mi.Team4_GetConfession()
	confessionMsg := mi.Team4_CreateConfessionMessage(confession)
	mi.BroadcastSyncMessageToTeam(confessionMsg)
}

func (mi *ExtendedAgent) Team4_HandleConfessionMessage(msg *common.Team4_ConfessionMessage) {
	if mi.VerboseLevel > 8 {
		if msg.Confession {
			log.Printf("Agent %s received confession notification from %s: I'm really sorry :(",
				mi.GetID(), msg.GetSender())
		} else {
			log.Printf("Agent %s received confession notification from %s: Noo! I'm innocent I swear!",
				mi.GetID(), msg.GetSender())
		}
	}
	// Team's agent should implement logic to store or process the reported proposed withdrawal amount as desired
}

// Get agents vote map for
func (mi *ExtendedAgent) Team4_GetProposedWithdrawalVote() map[uuid.UUID]int {
	return make(map[uuid.UUID]int)
}

func (mi *ExtendedAgent) Team4_GetProposedWithdrawal(instance common.IExtendedAgent) int {
	// first check if the agent has a team
	if !mi.HasTeam() {
		return 0
	}
	// Currently, assume stated withdrawal matches actual withdrawal
	return instance.Team4_ProposeWithdrawal()
}

func (mi *ExtendedAgent) Team4_StateProposalToTeam() {
	// Broadcast contribution to team
	proposedWithdrawal := mi.Team4_GetProposedWithdrawal(mi)
	proposalMsg := mi.Team4_CreateProposedWithdrawalMessage(proposedWithdrawal)
	mi.BroadcastSyncMessageToTeam(proposalMsg)
}

func (mi *ExtendedAgent) Team4_ProposeWithdrawal() int {
	// first check if the agent has a team
	if !mi.HasTeam() {
		return 0
	}
	if mi.Server.GetTeam(mi.GetID()).TeamAoA != nil {
		// double check if score in agent is sufficient (this should be handled by AoA though)
		commonPool := mi.Server.GetTeam(mi.GetID()).GetCommonPool()
		aoaExpectedWithdrawal := mi.Server.GetTeam(mi.GetID()).TeamAoA.GetExpectedWithdrawal(mi.GetID(), mi.GetTrueScore(), commonPool)
		if commonPool < aoaExpectedWithdrawal {
			return commonPool
		}
		return aoaExpectedWithdrawal + rand.Intn(4)
	} else {
		if mi.VerboseLevel > 6 {
			log.Printf("[WARNING] Agent %s has no AoA, withdrawing 0\n", mi.GetID())
		}
		return 0
	}
}

func (mi *ExtendedAgent) Team4_HandleProposedWithdrawalMessage(msg *common.Team4_ProposedWithdrawalMessage) {
	if mi.VerboseLevel > 8 {
		log.Printf("Agent %s received proposed withdrawal notification from %s: amount=%d\n",
			mi.GetID(), msg.GetSender(), msg.StatedAmount)
	}
	// Team's agent should implement logic to store or process the reported proposed withdrawal amount as desired
}

func (mi *ExtendedAgent) Team4_GetPunishmentVoteMap() map[int]int {
	return make(map[int]int)

}

// ----------------------- Team 1 AoA Functions -----------------------

func (mi *ExtendedAgent) Team1_ChairUpdateRanks(currentRanking map[uuid.UUID]int) map[uuid.UUID]int {
	// Chair iterates through existing rank map in team
	// and gets the new ranks of the agents in the team
	// according to AoA function
	newRanking := make(map[uuid.UUID]int)
	for agentUUID, _ := range currentRanking {
		newRank := mi.Server.GetTeam(agentUUID).TeamAoA.(*common.Team1AoA).GetAgentNewRank(agentUUID)
		newRanking[agentUUID] = newRank
	}

	// Returns a map of agent UUIDs to new Rank (int)
	return newRanking
}

func (mi *ExtendedAgent) Team1_VoteOnRankBoundaries(initialBoundaries [5]int) [5]int {
	// Default behaviour should just vote for the guideline rank boundaries
	return initialBoundaries
}
