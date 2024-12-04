package common

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

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

func (t *Team4) GetAuditCost(commonPool int) int {
	return 1
}

// Punishment Voting System
func (t *Team4) AoA4HandlePunishmentVote(punishmentVoteMap map[uuid.UUID]map[int]int) int {
	punishmentGrades := make(map[int][]int)

	for _, votes := range punishmentVoteMap {
		for punishment, grade := range votes {
			punishmentGrades[punishment] = append(punishmentGrades[punishment], grade)
		}
	}

	// Calculate median for each punishment
	medianGrades := make(map[int]int)
	for punishment, grades := range punishmentGrades {
		medianGrades[punishment] = calculateMedian(grades)
	}

	// Determine punishment with the highest median grade
	var selectedPunishment int
	highestMedian := -1
	for punishment, median := range medianGrades {
		if median > highestMedian {
			highestMedian = median
			selectedPunishment = punishment
		}
	}

	// Return punishment score (you can define values for each punishment)
	return getPunishmentScore(selectedPunishment)
}

func getPunishmentScore(punishment int) int {
	switch punishment {
	case 0:
		return 0 // No punishment
	case 1:
		return 10 // Small fine
	case 2:
		return 25 // Moderate fine
	case 3:
		return 50 // Large fine
	case 4:
		return 100 // Severe punishment
	default:
		return 0
	}
}

func calculateMedian(grades []int) int {
	if len(grades) == 0 {
		return 0
	}

	sort.Ints(grades)

	mid := len(grades) / 2
	if len(grades)%2 == 0 {
		// Average of middle grades if even
		return (grades[mid-1] + grades[mid]) / 2
	}
	// Middle grade if odd
	return grades[mid]
}

func (t *Team4) AoA4SetRankUp(rankUpVoteMap map[uuid.UUID]map[uuid.UUID]int) {
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

func (t *Team4) AoA4RunProposedWithdrawalVote(proposedWithdrawalMap map[uuid.UUID]int, withdrawalVoteMap map[uuid.UUID]map[uuid.UUID]int) {
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
