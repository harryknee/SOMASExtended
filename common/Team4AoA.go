package common

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// A bit messy right now will clean up soon and add detailed and clear description of how our AoA works.

// Handle Audit Result
// Handle Punishment + Punishment vote

func CreateTeam4AoA(team *Team) *Team4 {

	adventurers := make(map[uuid.UUID]struct {
		Rank               string
		ExpectedWithdrawal int
	})
	auditMap := make(map[uuid.UUID][]int)

	// Populate the maps based on the given team
	for _, agent := range team.Agents {
		adventurers[agent] = struct {
			Rank               string
			ExpectedWithdrawal int
		}{
			Rank:               "F",
			ExpectedWithdrawal: 1,
		}
		auditMap[agent] = []int{}
	}

	return &Team4{
		Adventurers: adventurers,
		AuditMap:    auditMap,
	}
}

type Team4 struct {
	Adventurers map[uuid.UUID]struct {
		Rank               string
		ExpectedWithdrawal int
	}
	AuditMap map[uuid.UUID][]int
}

func (t *Team4) SetRankUp(rankUpVoteMap map[uuid.UUID]map[uuid.UUID]int) {
	approvalCounts := make(map[uuid.UUID]int)

	for _, voteMap := range rankUpVoteMap {
		for votedForID, vote := range voteMap {
			if vote == 1 {
				approvalCounts[votedForID]++
			}
		}
	}
	threshold := t.GetRankUpThreshold()

	fmt.Printf("Rank Up Vote Threshold = %d approvals\n", threshold)

	for agentID, approvalCount := range approvalCounts {

		if approvalCount >= threshold {
			fmt.Printf("Agent %v: Meets threshold, ranking up!\n", agentID)

			// If the agent has enough approvals, rank them up
			t.RankUp(agentID)
			adventurer := t.Adventurers[agentID]
			fmt.Printf("Agent %v: New Rank = %s\n", agentID, adventurer.Rank)

		}
	}
}

// Use this to increment contributions for rank raises and have agents declare what they want to withdraw
func (t *Team4) SetContributionAuditResult(agentId uuid.UUID, agentScore int, agentActualContribution int, agentStatedContribution int) {
	// Check if adventurer in Team4 struct
	adventurer, exists := t.Adventurers[agentId]
	if !exists {
		// If the adventurer isn't in the Team4 struct
		// add agent with a starting contribution of 0
		adventurer = struct {
			Rank               string
			ExpectedWithdrawal int
		}{
			Rank:               "F",
			ExpectedWithdrawal: 1,
		}
	}

	// Update the adventurers contribution in the map in the map
	t.Adventurers[agentId] = adventurer

	contributionDiff := agentStatedContribution - agentActualContribution
	if agentStatedContribution > agentActualContribution || agentActualContribution < 2 {
		t.AuditMap[agentId] = append(t.AuditMap[agentId], contributionDiff)

	}

}

func (t *Team4) SetWithdrawalAuditResult(agentId uuid.UUID, agentScore int, agentActualWithdrawal int, agentStatedWithdrawal int, commonPool int) {
	withdrawalDiff := agentStatedWithdrawal - agentActualWithdrawal
	if agentStatedWithdrawal > agentActualWithdrawal || agentActualWithdrawal < 2 {
		t.AuditMap[agentId] = append(t.AuditMap[agentId], withdrawalDiff)

	}
}

func (t *Team4) GetContributionAuditResult(agentId uuid.UUID) bool {
	results := t.AuditMap[agentId]

	if len(results) == 0 {
		return false
	}

	return results[len(results)-1] != 0
}

func (t *Team4) GetWithdrawalAuditResult(agentId uuid.UUID) bool {
	results := t.AuditMap[agentId]

	if len(results) == 0 {
		return false
	}

	return results[len(results)-1] != 0
}

func (t *Team4) ResetAuditMap() {
	t.AuditMap = make(map[uuid.UUID][]int)
}

func (t *Team4) GetExpectedContribution(agentId uuid.UUID, agentScore int) int {
	return 2
}

// Can take more than this and 'lie'
func (t *Team4) GetExpectedWithdrawal(agentId uuid.UUID, agentScore int, commonPool int) int {
	adventurer, exists := t.Adventurers[agentId]
	if !exists {
		return 1
	}

	return adventurer.ExpectedWithdrawal
}

func (t *Team4) RunProposedWithdrawalVote(proposedWithdrawalMap map[uuid.UUID]int, withdrawalVoteMap map[uuid.UUID]map[uuid.UUID]int) {
	agentVoteWeightMap := make(map[uuid.UUID]int)

	for voterID, voteMap := range withdrawalVoteMap {
		// Get the agent's rank to determine their vote weight
		voter := t.Adventurers[voterID]
		voteWeight := t.GetVoteWeight(voter.Rank)

		for votedForID, vote := range voteMap {
			if vote == 1 {
				agentVoteWeightMap[votedForID] += voteWeight
			}
		}
	}

	// Get threshold for if their proposed withdrawal is accepted
	threshold := t.GetVoteThreshold()

	for agentID, totalVoteWeight := range agentVoteWeightMap {
		if totalVoteWeight >= threshold {
			proposedWithdrawal, exists := proposedWithdrawalMap[agentID]
			if exists {

				// Update the agent's expected withdrawal if their vote weight meets the threshold
				adventurer := t.Adventurers[agentID]

				fmt.Printf("Agent %v: Current ExpectedWithdrawal = %d, Proposed = %d\n", agentID, adventurer.ExpectedWithdrawal, proposedWithdrawal)

				oldWithdrawal := adventurer.ExpectedWithdrawal

				adventurer.ExpectedWithdrawal = proposedWithdrawal

				fmt.Printf("Agent %v: Proposed Withdrawal Accepted, changed from %d to %d\n", agentID, oldWithdrawal, adventurer.ExpectedWithdrawal)

				// Update the agent in the Adventurers map
				t.Adventurers[agentID] = adventurer
			}
		}
	}
}

func (t *Team4) GetAuditCost(commonPool int) int {
	return 1
}

func (t *Team4) GetVoteThreshold() int {
	totalAdventurers := len(t.Adventurers)

	threshold := totalAdventurers * 70 / 100

	return threshold
}

func (t *Team4) GetRankUpThreshold() int {
	totalAdventurers := len(t.Adventurers)

	threshold := totalAdventurers / 2

	return threshold
}

// Will change this since we want it return a slice of AgentIds, takes in different input as well
func (t *Team4) GetVoteResult(votes []Vote) uuid.UUID {
	voteMap := make(map[uuid.UUID]int)
	for _, vote := range votes {
		if vote.IsVote >= 1 {

			// Get the rank of the voter
			voter, exists := t.Adventurers[vote.VoterID]
			if !exists {
				continue
			}

			// Get the vote weight based on the voter's rank
			voteWeight := t.GetVoteWeight(voter.Rank)

			// Accumulate the vote scaled by the voter's rank
			voteMap[vote.VotedForID] += voteWeight
		}
	}

	// Calculate the vote threshold
	threshold := t.GetVoteThreshold()

	// Check if any candidate's exceed the threshold
	for votedForID, totalVotes := range voteMap {
		if totalVotes >= threshold {
			return votedForID
		}
	}

	// If no candidate exceeds the threshold, return nil
	return uuid.Nil
}

// GetWithdrawalOrder orders adventurers based on their vote weight (highest first).
func (t *Team4) GetWithdrawalOrder(agentIDs []uuid.UUID) []uuid.UUID {
	type agentWithWeight struct {
		ID     uuid.UUID
		Weight int
	}

	// Create a slice to store agent IDs along with their vote weight
	agentsWithWeight := make([]agentWithWeight, len(agentIDs))
	for i, id := range agentIDs {
		// Retrieve the adventurer's rank and calculate their vote weight
		weight := t.GetVoteWeight(t.Adventurers[id].Rank)
		agentsWithWeight[i] = agentWithWeight{
			ID:     id,
			Weight: weight,
		}
	}

	// Sort the slice by vote weight in descending order
	sort.Slice(agentsWithWeight, func(i, j int) bool {
		return agentsWithWeight[i].Weight > agentsWithWeight[j].Weight
	})

	// Extract and return the ordered agent IDs
	orderedIDs := make([]uuid.UUID, len(agentIDs))
	for i, aw := range agentsWithWeight {
		orderedIDs[i] = aw.ID
	}

	return orderedIDs
}

func (t *Team4) HandleConfessionAuditResult(confession bool, agentId uuid.UUID, agentScore int) int {
	fmt.Println("here")
	// Check if confession is true
	if confession {
		// Apply a reduced punishment
		return agentScore * 25 / 100
	}

	// Access the agent's audit history
	auditHistory, exists := t.AuditMap[agentId]
	if !exists || len(auditHistory) < 2 {
		// If no sufficient audit history, return the original score
		return agentScore
	}

	lastIndex := len(auditHistory) - 1
	if auditHistory[lastIndex] != 0 || auditHistory[lastIndex-1] != 0 {
		return agentScore * 50 / 100
	}

	return agentScore
}

func (t *Team4) RankUp(agentID uuid.UUID) {
	adventurer, exists := t.Adventurers[agentID]
	if !exists {
		return
	}
	switch adventurer.Rank {
	case "F":
		adventurer.Rank = "E"
	case "E":
		adventurer.Rank = "D"
	case "D":
		adventurer.Rank = "C"
	case "C":
		adventurer.Rank = "B"
	case "B":
		adventurer.Rank = "A"
	case "A":
		adventurer.Rank = "S"
	case "S":
		adventurer.Rank = "SS"
	case "SS":
		adventurer.Rank = "SSS"
	case "SSS":
		adventurer.Rank = "SSS"
	}
	t.Adventurers[agentID] = adventurer
}

func (t *Team4) GetVoteWeight(rank string) int {
	switch rank {
	case "SSS":
		return 10
	case "S":
		return 8
	case "A":
		return 6
	case "B":
		return 4
	case "C":
		return 3
	case "D":
		return 2
	case "E":
		return 1
	case "F":
		return 1
	default:
		return 0
	}
}

// Need to handle AUDIT truth value and storing + how it succeeds
/*Extra Functions to implement:
team.TeamAoA.SetRankUp(rankUpVoteMap) Done
team.TeamAoA.RunWithdrawalVote(proposedWithdrawalMap) Done

agent.GetStatedWithdrawal(agent) if this needs to be different for ProposedWithdrawal

agent.GetRankUpVote()
agent.GetConfession(agent)
agent.SetAgentAuditResult(agent, agentConfession)
agent.GetWithdrawalVote(agent)
*/

/* Our Turn Flow - Additional Flow marked with ***************
func (cs *EnvironmentServer) RunTurn(i, j int) {
	fmt.Printf("\n\nIteration %v, Turn %v, current agent count: %v\n", i, j, len(cs.GetAgentMap()))

	cs.teamsMutex.Lock()
	defer cs.teamsMutex.Unlock()

	for _, team := range cs.teams {
		fmt.Println("\nRunning turn for team ", team.TeamID)
		// Sum of contributions from all agents in the team for this turn
		agentContributionsTotal := 0
		for _, agentID := range team.Agents {
			agent := cs.GetAgentMap()[agentID]
			if agent.GetTeamID() == uuid.Nil || cs.IsAgentDead(agentID) {
				continue
			}
			agent.StartRollingDice(agent)
			agentActualContribution := agent.GetActualContribution(agent)
			agentContributionsTotal += agentActualContribution
			// Assume this will be broadcasted to each agent in a team via message in sync / update a map in Agent's struct
			agentStatedContribution := agent.GetStatedContribution(agent)
			agentScore := agent.GetTrueScore()
			// Update audit result for this agent
			team.TeamAoA.SetContributionAuditResult(agentID, agentScore, agentActualContribution, agentStatedContribution)
			agent.SetTrueScore(agentScore - agentActualContribution)
		}



		// Update common pool with total contribution from this team
		// 	Agents do not get to see the common pool before deciding their contribution
		//  Different to the withdrawal phase!
		team.SetCommonPool(team.GetCommonPool() + agentContributionsTotal)




		// Do AoA processing
		team.TeamAoA.RunAoAStuff()

		// Initiate Contribution Audit vote
		contributionAuditVotes := []common.Vote{}
		for _, agentID := range team.Agents {
			agent := cs.GetAgentMap()[agentID]
			vote := agent.GetContributionAuditVote()
			contributionAuditVotes = append(contributionAuditVotes, vote)
		}

		// Execute Contribution Audit if necessary
		if agentToAudit := team.TeamAoA.GetVoteResult(contributionAuditVotes); agentToAudit != uuid.Nil {
			auditResult := team.TeamAoA.GetContributionAuditResult(agentToAudit)
			for _, agentID := range team.Agents {
				agent := cs.GetAgentMap()[agentID]
				agent.SetAgentContributionAuditResult(agentToAudit, auditResult)
			}
		}

		orderedAgents := team.TeamAoA.GetWithdrawalOrder(team.Agents)
		for _, agentID := range orderedAgents {
			agent := cs.GetAgentMap()[agentID]
			if agent.GetTeamID() == uuid.Nil || cs.IsAgentDead(agentID) {
				continue
			}

			// Pass the current pool value to agent's methods
			currentPool := team.GetCommonPool()
			agentActualWithdrawal := agent.GetActualWithdrawal(agent)
			if agentActualWithdrawal > currentPool {
				agentActualWithdrawal = currentPool // Ensure withdrawal does not exceed available pool
			}
			agentStatedWithdrawal := agent.GetStatedWithdrawal(agent)
			agentScore := agent.GetTrueScore()
			// Update audit result for this agent
			team.TeamAoA.SetWithdrawalAuditResult(agentID, agentScore, agentActualWithdrawal, agentStatedWithdrawal, team.GetCommonPool())
			agent.SetTrueScore(agentScore + agentActualWithdrawal)

			// Update the common pool after each withdrawal so agents can see the updated pool before deciding their withdrawal.
			//  Different to the contribution phase!
			team.SetCommonPool(currentPool - agentActualWithdrawal)
			fmt.Printf("[server] Agent %v withdrew %v. Remaining pool: %v\n", agentID, agentActualWithdrawal, team.GetCommonPool())
		}



		// Initiate Withdrawal Audit vote
		withdrawalAuditVotes := []common.Vote{}
		for _, agentID := range team.Agents {
			agent := cs.GetAgentMap()[agentID]
			vote := agent.GetWithdrawalAuditVote()
			withdrawalAuditVotes = append(withdrawalAuditVotes, vote)
		}


		// Execute Withdrawal Audit if necessary
		if agentToAudit := team.TeamAoA.GetVoteResult(withdrawalAuditVotes); agentToAudit != uuid.Nil {
			auditResult := team.TeamAoA.GetWithdrawalAuditResult(agentToAudit)
			for _, agentID := range team.Agents {
				agent := cs.GetAgentMap()[agentID]
				agent.SetAgentWithdrawalAuditResult(agentToAudit, auditResult)
			}
		}
	}

	// TODO: Reallocate agents who left their teams during the turn
}

*/
