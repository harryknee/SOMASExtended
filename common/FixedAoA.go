package common

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type FixedAoA struct {
	auditRecord *AuditRecord
}

func (f *FixedAoA) GetExpectedContribution(agentId uuid.UUID, agentScore int) int {
	return agentScore
}

func (f *FixedAoA) SetContributionAuditResult(agentId uuid.UUID, agentScore int, agentActualContribution int, agentStatedContribution int) {
	// If the agent's actual contribution is not equal to the stated contribution, then the agent is cheating
	infraction := 0
	if agentActualContribution != agentStatedContribution {
		infraction = 1
	}
	f.auditRecord.AddRecord(agentId, infraction)
}

func (f *FixedAoA) GetContributionAuditResult(agentId uuid.UUID) bool {
	// true means agent failed the audit (cheated)
	infractions := f.auditRecord.GetAllInfractions(agentId) > 0
	f.auditRecord.ClearAllInfractions(agentId)
	return infractions
}

func (f *FixedAoA) GetExpectedWithdrawal(agentId uuid.UUID, agentScore int, commonPool int) int {
	return 2
}

func (f *FixedAoA) SetWithdrawalAuditResult(agentId uuid.UUID, agentScore int, agentActualWithdrawal int, agentStatedWithdrawal int, commonPool int) {
	// If the agent's actual withdrawal is not equal to the stated withdrawal, then the agent is cheating
	if agentActualWithdrawal != agentStatedWithdrawal {
		f.auditRecord.IncrementLastRecord(agentId)
	}
}

func (f *FixedAoA) GetWithdrawalAuditResult(agentId uuid.UUID) bool {
	// true means agent failed the audit (cheated)
	infractions := f.auditRecord.GetAllInfractions(agentId) > 0
	f.auditRecord.ClearAllInfractions(agentId)
	return infractions
}

func (f *FixedAoA) GetAuditCost(commonPool int) int {
	return f.auditRecord.GetAuditCost()
}

// MUST return UUID nil if audit should not be executed
// Otherwise, implement a voting mechanism to determine the agent to be audited
// and return its UUID
func (f *FixedAoA) GetVoteResult(votes []Vote) uuid.UUID {
	return uuid.Nil
}

func (t *FixedAoA) GetWithdrawalOrder(agentIDs []uuid.UUID) []uuid.UUID {
	// Seed the random number generator to ensure different shuffles each time
	rand.Seed(time.Now().UnixNano())

	// Create a copy of the agentIDs to avoid modifying the original list
	shuffledAgents := make([]uuid.UUID, len(agentIDs))
	copy(shuffledAgents, agentIDs)

	// Shuffle the agent list
	rand.Shuffle(len(shuffledAgents), func(i, j int) {
		shuffledAgents[i], shuffledAgents[j] = shuffledAgents[j], shuffledAgents[i]
	})

	return shuffledAgents
}

func CreateFixedAoA(duration int) IArticlesOfAssociation {
	auditRecord := NewAuditRecord(duration)
	return &FixedAoA{
		auditRecord: auditRecord,
	}
}

// Do nothing
func (t *FixedAoA) AoA4SetRankUp(rankUpVoteMap map[uuid.UUID]map[uuid.UUID]int) {
}
func (t *FixedAoA) AoA4RunProposedWithdrawalVote(map[uuid.UUID]int, map[uuid.UUID]map[uuid.UUID]int) {
}
func (t *FixedAoA) AoA4HandlePunishmentVote(map[uuid.UUID]map[int]int) int {
	return 0
}
