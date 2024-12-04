package agents

/* Contains functions relevant to agent functionality when the AoA is Team1AoA */

import (
	"log"
	"github.com/google/uuid"

	common "github.com/ADimoska/SOMASExtended/common"
)


func (mi *ExtendedAgent) Team1_ChairUpdateRanks(currentRanking map[uuid.UUID]int) map[uuid.UUID]int {
	// Chair iterates through existing rank map in team
	// and gets the new ranks of the agents in the team
	// according to AoA function
	newRanking := make(map[uuid.UUID]int)
	for agentUUID := range currentRanking {
		newRank := mi.Server.GetTeam(agentUUID).TeamAoA.(*common.Team1AoA).GetAgentNewRank(agentUUID)
		newRanking[agentUUID] = newRank
	}

	// Returns a map of agent UUIDs to new Rank (int)
	return newRanking
}


/**
* Agrees the rank boundaries as a social decision. This is called as part of
* pre-roll logic, using various helper functions below. This function will only
* be called by an elected chair
*/
func (mi *ExtendedAgent) Team1_AgreeRankBoundaries() [5]int {
    // Step 1. Gather boundary opinions
    prefs := mi.team1_GatherRankBoundaryPreferences()

    // Step 2. Generate candidates using statistical analysis of individual
    // agent opinions, providing 3 options to choose from
    cands := mi.team1_GenerateRankBoundaryCandidates(prefs)

    // Step 3. Conduct a vote on the candidates 
    elected := mi.team1_VoteOnRankBoundaries(cands)

    return elected
}

/**
* Go through all the agents, and ask them what they would like the rank
* boundaries to be. Their returned value is completely up to the agent
* strategy.
*/
func (mi *ExtendedAgent) team1_GatherRankBoundaryPreferences() [][5]int {
    opinions := make([][5]int, 0)
    // for all agents in team
    for _, agentID := range mi.Server.GetAgentsInTeam(mi.TeamID) {
        // Ask them what they think the rank boundaries should be
        req := &common.Team1RankBoundaryRequestMessage{
            BaseMessage:  mi.CreateBaseMessage(),
        }
        mi.SendSynchronousMessage(req, agentID)
    }
    // Iterate over all agents and ask them for their opinions. We do not
    // store who each vote came from to enforce anonymity
    return opinions
}

/**
* Generate the candidates based off the opinions expressed by the agents
*/
func (mi *ExtendedAgent) team1_GenerateRankBoundaryCandidates(opinions [][5]int) [3][5]int {
    print(opinions)
    return [3][5]int{{1, 2, 3, 4, 5}, {1, 2, 3, 4, 5}, {1, 2, 3, 4, 5}}
}

/**
* Conduct a vote on the expressed candidates, and elect a winner
*/
func (mi *ExtendedAgent) team1_VoteOnRankBoundaries(cands [3][5]int) [5]int {
	// Default behaviour just returns the first candidate
	return cands[0]
}

/**
* OVERRIDE
* You get a say in what you think the rank boundaries should be!
*/
func (mi *ExtendedAgent) Team1_BoundaryOpinionRequestHandler(msg *common.Team1RankBoundaryRequestMessage) {
    sender := msg.GetSender()
    resp :=  &common.Team1RankBoundaryResponseMessage{
		BaseMessage:  mi.CreateBaseMessage(),
        Bounds: [5]int{1, 2, 3, 4, 4},
	}
    mi.SendSynchronousMessage(resp, sender)
}

func (mi *ExtendedAgent) Team1_BoundaryOpinionResponseHandler(msg *common.Team1RankBoundaryResponseMessage) {
    log.Printf("Chair %v received rank boundary opinion from %v", mi.GetID(), msg.GetSender())
}
