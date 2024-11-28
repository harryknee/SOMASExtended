package aoa

/*
Team 4 AoA: The Adventurer’s Guild
In the world of fantasy, the adventurers are a motley band of brave souls — warriors, mages,
rogues, and all united by a thirst for glory, treasure, and the thril of delving into the
unknown, despite their clashing motives and pasts.

There’s one peaceful hove for all the adventurers --- The Adventurer’s Guild!
The rules of theguild are simple:

⚫ If you are weak, take care of yourselves. If you are strong, it's up to you empathise and help others.
⚫ The guild won’t feed you for doing nothing, you must rely on your own hands.
⚫ Don’t be a cheat, anyone can spot you, you won’t make it out the guild grounds alive.

Rules:
Expected Contribution: 2 (Basic membership fee for the guild)
Expected Withdrawal: 1 (Some Food and Wine on the house)
Audit Cost: 1 (Anyone can raise a concern)
Audit Punishment: 100000000000 (The guild will gladly take your noble donation)


We incentivise everyone to play safe and take care of themselves, at the end of the day, if you
were too reckless and died to some beast in the forest, it's not our problem. Adventurers
come and go, only the careful ones remain in the world.


-- Contribution Rules:
If you ran back from the deadly battle and lost all your stuff, fine, don’t donate.
Otherwise, at least pay the little drink money at the tavern.
If you are feeling generous today and want to help your buddies, please donate more.

-- Withdrawal Rules:
You can take a little bit from the pot (1), that’s fine.
But if you want to take more than expected, you will need a majority vote from everyone in
the group of people who declared to have contributed more than expected. (Their vote is
weighted by total declared contribution).


-- Audit Cost: (Audit can happen at any time not occupied by business)
If you find someone doing something fishy at the corner, you can chuck the nearest drink jug
at him, just need to pay the jug money (1). You can think of the audition vote as a massive pub
brawl. Then people choose sides, the side with most headcount wins.
If you initiated the audition, and you lose by headcount, you need to pay everyone on their
team a drink (1). If you win, the truth spell is cast on the suspect. If you convicted a crime,
audited someone else as an accomplice, but lost the headcount vote, you are exiled.


-- Audit Punishment:
If there is indeed a crime convicted, we allow the poor criminal caught to do the following:
1. He can make up an excuse, people vote for the excuse. If people forgive him, he gets
   to keep his stuff. If not, he donates all his gold to the guild and rolls out the door.
2. He can audit another accomplice, if the accomplice is also guilty (if the previous
   headcount vote succeeds), then he is not exiled but loses all his money to the common pool.
   If not, he is exiled.

If there is no crime convicted by the suspect, the common pool pays him a drink (1).

Additional Flavors:
Bribing:
During the voting process, you can offer someone some money to cast the same vote as you
do. If they accept, after voting, if indeed the same vote is cast, the transaction is complete.
*/

import (
	"github.com/google/uuid"
)

type Team4 struct {
	Adventurers map[uuid.UUID]struct {
		Contribution int
		Rank         string
	}
	AuditMap map[uuid.UUID][]int
}

// Use this to increment contributions for rank raises
func (t *Team4) SetContributionAuditResult(agentId uuid.UUID, agentScore int, agentActualContribution int, agentStatedContribution int) {

	// Check if adventurer in Team4 struct
	adventurer, exists := t.Adventurers[agentId]
	if !exists {
		// If the adventurer isn't in the Team4 struct
		// add agent with a starting contribution of 0
		adventurer = struct {
			Contribution int
			Rank         string
		}{
			Contribution: 0,
			Rank:         "No Rank",
		}
	}

	// Increment the adventurer's rank from the previous turn
	adventurer.Rank = t.GetRank(adventurer.Contribution)

	// Increment the adventurer's contribution by the STATED contribution
	adventurer.Contribution += agentStatedContribution

	// Update the adventurers contribution in the map in the map
	t.Adventurers[agentId] = adventurer
}

func (t *Team4) GetExpectedContribution(agentId uuid.UUID, agentScore int) int {
	return 2
}

// Can take more than this and 'lie'
func (t *Team4) GetExpectedWithdrawal(agentId uuid.UUID) int {

	return 1
}

func (t *Team4) GetAuditCost(commonPool int) int {

	return 1
}

// Calculates voting threshold based off of number of adventurers
func (t *Team4) GetVoteThreshold() int {
	totalAdventurers := len(t.Adventurers)
	threshold := totalAdventurers / 2

	return threshold
}

func (t *Team4) GetVoteResult(votes []Vote) *uuid.UUID {
	voteMap := make(map[uuid.UUID]int)
	for _, vote := range votes {
		if vote.IsVote {

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
			return &votedForID
		}
	}

	// If no candidate exceeds the threshold, return nil
	return &uuid.Nil
}

func (t *Team4) GetRank(contribution int) string {
	switch {
	case contribution >= 10000:
		return "SSS"
	case contribution >= 64:
		return "S"
	case contribution >= 32:
		return "A"
	case contribution >= 16:
		return "B"
	case contribution >= 8:
		return "C"
	case contribution >= 4:
		return "D"
	case contribution >= 2:
		return "E"
	case contribution >= 1:
		return "F"
	default:
		return "No Rank"
	}
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

/* TO DO
- Audit result <- probability of success decreases on Rank.
- Bribe.
- Override with a 'drinking phase' <- aim to increase familiarity / trust with other agents.
- Rearrange Audit Vote order by rank <- messages are broadcast sync to the team.
- Keep track of who has voted for who during an audit.
- Losing side of 'audit' has to pay cleanup fees for damaging guild.
- Override how withdrawals work <- need a vote each time if they want to withdraw more than 1.
- Will be default to Random/Abdstain idk.
- Ideally GetVoteResult or a similar function used for auditing and voting on withdrawals.

*/
