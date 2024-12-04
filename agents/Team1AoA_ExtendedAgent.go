package agents

/* Contains functions relevant to agent functionality when the AoA is Team1AoA */

import (
	"log"
	"math"
	"sort"

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
	// Step 1. Gather boundary proposals
	mi.team1_GatherRankBoundaryProposals()

	// Step 2. Generate candidates using statistical analysis of individual
	// agent proposals, providing 3 options to choose from
	cands := mi.team1_GenerateRankBoundaryCandidates()

	// Step 3. Conduct a vote on the candidates
	elected := mi.team1_VoteOnRankBoundaries(cands)

	return elected
}

/**
* Go through all the agents, and ask them what they would like the rank
* boundaries to be. Their returned value is completely up to the agent
* strategy.
 */
func (mi *ExtendedAgent) team1_GatherRankBoundaryProposals() {

	// Same request for all agents
	req := &common.Team1RankBoundaryRequestMessage{
		BaseMessage: mi.CreateBaseMessage(),
	}

	// Clear temp variable - this is just the location that the chair will add
	// all the data to as it comes into requests
	mi.team1RankBoundaryProposals = mi.team1RankBoundaryProposals[0:]

	// Iterate over all agents and ask them for their proposals. We do not
	// store who each vote came from to enforce anonymity
	for _, agentID := range mi.Server.GetAgentsInTeam(mi.TeamID) {
		mi.SendSynchronousMessage(req, agentID)
	}
}

/**
* Generate the candidates based off the proposals expressed by the agents
 */
func (mi *ExtendedAgent) team1_GenerateRankBoundaryCandidates() [3][5]int {
	// This function will take the values stored in team1RankBoundaryProposals
	// (which should have been collected just before this).
	if mi.team1RankBoundaryProposals == nil {
		log.Fatal("Tried generating candidates without asking for agent proposals! (nil pointer)")
		return [3][5]int{}
	}

	// Transpose the proposals matrix - This converts arrays of proposals of
	// each agent to arrays of proposals for each rank
	rows := len(mi.team1RankBoundaryProposals)
	cols := 5 // for 5 ranks

	// pre-allocate
	boundaries := [5][]int{}
	for i := 0; i < len(boundaries); i++ {
		boundaries[i] = make([]int, rows)
	}

	// fill
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			boundaries[j][i] = mi.team1RankBoundaryProposals[i][j]
		}
	}

	lower_candidate := [5]int{} // lower quartile of each bound
	avg_candidate := [5]int{}   // median of each bound
	upper_candidate := [5]int{} // upper quartile of each bound

	// Go through each rank boundary, and add it to the respective candidate
	for pos, b := range boundaries {
		q1, q2, q3 := CalculateQuartiles(b)
		lower_candidate[pos] = q1
		avg_candidate[pos] = q2
		upper_candidate[pos] = q3
	}

	/* There is a rare edge case here where the median for a higher quartile is
	 * lower than one for a lower quartile, if you have unprogressive or
	 * bad-actor agents. We account for this here by strictly enforcing that the
	 * options are non-decreasing. */
	for i := 1; i < 5; i++ {
		if prev := lower_candidate[i-1]; prev > lower_candidate[i] {
			lower_candidate[i] = lower_candidate[i-1]
		}
		if prev := avg_candidate[i-1]; prev > avg_candidate[i] {
			avg_candidate[i] = avg_candidate[i-1]
		}
		if prev := upper_candidate[i-1]; prev > upper_candidate[i] {
			upper_candidate[i] = upper_candidate[i-1]
		}
	}

	cands := [3][5]int{lower_candidate, avg_candidate, upper_candidate}
	return cands
}

/**
* Conduct a vote on the expressed candidates, and elect a winner
 */
func (mi *ExtendedAgent) team1_VoteOnRankBoundaries(cands [3][5]int) [5]int {
	// Default behaviour just returns the first candidate
	return cands[0]
}

/**
* Calculate the median, upper and lower quartiles for a range of data. This is
* used for generating candidates based on individual rank boundaries provided
* by agents. Calculations are done based on floats but results are rounded to
* integers before being returned.
 */
func CalculateQuartiles(values []int) (q1, q2, q3 int) {
	n := len(values)
	sort.Ints(values)

	// Calculate each quartile
	q1 = calculateMedian(values[:n/2]) // lower quartile
	q2 = calculateMedian(values)       // median

	/* Upper quartile needs a little more thought, since you need to account
	   for the case where the size is even. This is not a problem for the lower
	   quartile because n/2 gets floored down if it is a float. But in this case
	   we need to think whether we consider the median or not. */

	if n%2 == 0 {
		q3 = calculateMedian(values[n/2:])
	} else {
		q3 = calculateMedian(values[n/2+1:])
	}

	return
}

/**
* Calculate the median of a set of values. Returns an integer, rounding to the
* nearest whole number (not truncation)
 */
func calculateMedian(sortedValues []int) int {
	n := len(sortedValues)
	var median float64 = 0.0
	if n%2 == 0 {
		median = float64(sortedValues[n/2-1]+sortedValues[n/2]) / 2.0
	} else {
		median = float64(sortedValues[n/2])
	}
	return int(math.Round(median))
}

/**
* BASE IMPLEMENTATION - Always returns the same boundaries.
* OVERRIDE - You get a say in what you think the rank boundaries should be!
 */
func (mi *ExtendedAgent) Team1_BoundaryProposalRequestHandler(msg *common.Team1RankBoundaryRequestMessage) {
	bounds := [5]int{10, 20, 30, 40, 50}

	resp := &common.Team1RankBoundaryResponseMessage{
		BaseMessage: mi.CreateBaseMessage(),
		Bounds:      bounds,
	}

	mi.SendSynchronousMessage(resp, msg.GetSender())
}

func (mi *ExtendedAgent) Team1_BoundaryProposalResponseHandler(msg *common.Team1RankBoundaryResponseMessage) {
	log.Printf("Chair %v received rank boundary proposal %v from %v", mi.GetID(), msg.Bounds, msg.GetSender())
	mi.team1RankBoundaryProposals = append(mi.team1RankBoundaryProposals, msg.Bounds)
	mi.SignalMessagingComplete()
}

// Returns the non-exported function for testing
func (mi *ExtendedAgent) TestableGatherFunc() func() {
	return mi.team1_GatherRankBoundaryProposals
}

func (mi *ExtendedAgent) TestableCandidateGenFunc(proposals [][5]int) func() [3][5]int {
	mi.team1RankBoundaryProposals = proposals
	return mi.team1_GenerateRankBoundaryCandidates
}
