package agents

import (
	"fmt"
	"math"
	"math/rand"

	common "github.com/ADimoska/SOMASExtended/common"

	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/agent"
	"github.com/google/uuid"
)

type MI_256_v1 struct {
	*ExtendedAgent

	//character matrix:
	chaoticness int // from 1 to 3, 3 being most chaotic
	evilness    int // from 1 to 3, 3 being most evil

	//opinions
	affinity map[uuid.UUID]int // has opinions for each agent in the game
	mood     int               // starting from 0

	// store of other character's states (what the id of the agents in the team )
	teamAgentsDeclaredRolls        map[uuid.UUID]int
	teamAgentsDeclaredContribution map[uuid.UUID]int
	teamAgentsDeclaredWithdraw     map[uuid.UUID]int
	teamAgentsExpectedScore        map[uuid.UUID]int

	teamAgentsExpectedContribution map[uuid.UUID]int
	teamAgentsExpectedWithdraw     map[uuid.UUID]int

	// store of the other character's states during an audit
	lastAuditTarget  uuid.UUID
	lastVotes        map[uuid.UUID]bool
	lastAuditStarter uuid.UUID

	// TODO: internal states

	// Information I need
	Common_pool int

	//AOA parameterss
	AOAOpinion              map[uuid.UUID]int
	AoAExpectedContribution int
	AoAExpectedWithdrawal   int
	AoAAuditCost            int
	AoAPunishment           int

	//Intended Withdrawal and Contribution
	IntendedWithdrawal   int
	declaredWithdrawal   int
	intendedContribution int
	declaredcontribution int

	isThereCheatThisRound bool
	haveIlied             bool
	IcaughtLying          bool
}

// constructor for MI_256_v1
func Team4_CreateAgent(funcs agent.IExposedServerFunctions[common.IExtendedAgent], agentConfig AgentConfig) *MI_256_v1 {
	mi_256 := &MI_256_v1{
		ExtendedAgent: GetBaseAgents(funcs, agentConfig),
	}
	mi_256.trueSomasTeamID = 4 // IMPORTANT: add your team number here!
	return mi_256
}

// ----functions to calculate the data of other team members------------
func (mi *MI_256_v1) UpdateTeamDeclaredRolls() {
}

//------functions to calculated the expected AOA contributions and stuff for all agents -------------------------------

func (mi *MI_256_v1) CalcAOAContibution() int {

	mi.AoAExpectedContribution = 0
	return mi.AoAExpectedContribution

}
func (mi *MI_256_v1) CalcAOAWithdrawal() {

}
func (mi *MI_256_v1) CalcAOAAuditCost() {

}
func (mi *MI_256_v1) CalcAOAPunishment() {

}
func (mi *MI_256_v1) CalcAOAOpinion() {

}

// ----------------------- Strategies -----------------------

func (mi *MI_256_v1) AnyoneCheatedAfterContribute() {
	// get common pool before and after contribution, compare with the total expected contribution, to see if there is a cheat this round

}
func (mi *MI_256_v1) AnyoneCheatedAfterWithdrawal() {

}

// Team-forming Strategy
func (mi *MI_256_v1) DecideTeamForming(agentInfoList []common.ExposedAgentInfo) []uuid.UUID {
	invitationList := []uuid.UUID{}
	for _, agentInfo := range agentInfoList {
		// exclude the agent itself
		if agentInfo.AgentUUID == mi.GetID() {
			continue
		}
		if agentInfo.AgentTeamID == (uuid.UUID{}) {
			invitationList = append(invitationList, agentInfo.AgentUUID)
		}

	}

	// TODO: implement team forming logic
	// random choice from the invitation list
	rand.Shuffle(len(invitationList), func(i, j int) { invitationList[i], invitationList[j] = invitationList[j], invitationList[i] })
	chosenAgent := invitationList[0]

	// Return a slice containing the chosen agent
	return []uuid.UUID{chosenAgent}
}

// Dice Strategy
func (mi *MI_256_v1) StickOrAgain(accumulatedScore int, prevRoll int) bool {
	fmt.Printf("Called overriden StickOrAgain\n")

	// the higher the mood, the more risky the roll dice strategy will be

	threshLow := 8
	threshMid := 11
	threshHigh := 14
	if mi.mood > 5 { // be greedy
		if mi.LastScore < threshHigh {
			return true
		} else {
			return false
		}
	} else if (-5 < mi.mood) && (mi.mood < 5) {
		if mi.LastScore < threshMid {
			return true
		} else {
			return false
		}
	} else {
		if mi.LastScore < threshLow {
			return true
		} else {
			return false
		}
	}

}

// !!! NOTE: name and signature of functions below are subject to change by the infra team !!!

// definition of Evilness: The more he would value his own benifit over others.
// definition of Chaoticness: The willingness to take risk to breach the AoA
// definition of Mood: The willingness to take actions which diverges from abosolute neutral
// Contribution Strategy
func (mi *MI_256_v1) DecideContribution() int {
	if mi.Score == 0 {
		return 0
	}

	//parameters: AoAExpectedContribution, Mood, Lawfulness,Evilness，  CurrentScore
	mi.intendedContribution = 0
	contribute_percentage := 0.0
	// CurrentScore = mi.Score

	mood_modifier := float64(1.0 / 32.0 * float64(mi.mood))
	neutral_mean := float64(mi.AoAExpectedContribution / mi.Score) //normalize from 0-1
	//we model evil and good with different standard deviations, with chaotic being high standard deviation,
	lawful_evil_mean := neutral_mean * 7 / 8
	neutral_evil_mean := neutral_mean / 2
	neutral_good_mean := (1-neutral_mean)/2 + neutral_mean
	lawful_good_mean := float64(7.0 / 8)
	neutral_standard_deviation := float64(1.0 / 12)
	lawful_standard_deviation := 1.0 / 16.0
	chaotic_standard_deviation := 1.0 / 8
	chaotic_good_standard_deviation := 3.0 / 8 * (1 - neutral_mean)
	chaotic_evil_standard_deviation := 3.0 / 8 * neutral_mean
	neutral_good_standard_deviation := 1.0 / 8 * (1 - neutral_mean)
	neutral_evil_standard_deviation := 1.0 / 8 * neutral_mean
	lawful_good_standard_deviation := 1.0 / 16 * (1 - neutral_mean)
	lawful_evil_standard_deviation := 1.0 / 16 * neutral_mean

	Apply_mood := func(contribute_percentage float64) float64 {
		if contribute_percentage > float64(neutral_mean) {
			return contribute_percentage + mood_modifier
		} else {
			return contribute_percentage - mood_modifier
		}
	}
	// if the agent is a lawful agent, he would tend to not break the AoA
	// if the agent is a good agent, he would tend to contribute more than he should
	// absolute neutral: donating as much as the AoA asks.

	// we set a linear specturm for the strategies and calculate where we would end up in this scale
	//
	// Each personality would have a mean on that scale, and would have standard deviations based on the chaoticness

	if mi.evilness == 2 { // if agent has neutral evilness, use neutral mean
		if mi.chaoticness == 1 { // if lawful neutral

			contribute_percentage = (rand.NormFloat64() * float64(lawful_standard_deviation)) + float64(neutral_mean)

		} else if mi.chaoticness == 2 { // abs neutral
			contribute_percentage = (rand.NormFloat64() * float64(neutral_standard_deviation)) + float64(neutral_mean)

		} else if mi.chaoticness == 3 { //chaotic neutral
			contribute_percentage = (rand.NormFloat64() * float64(chaotic_standard_deviation)) + float64(neutral_mean)

		}

	} else if mi.evilness == 1 { // if agent is good
		if mi.chaoticness == 1 { // if lawful good

			contribute_percentage = (rand.NormFloat64() * float64(lawful_good_standard_deviation)) + float64(lawful_good_mean)

		} else if mi.chaoticness == 2 { // neutral good
			contribute_percentage = (rand.NormFloat64() * float64(neutral_good_standard_deviation)) + float64(neutral_good_mean)

		} else if mi.chaoticness == 3 { //chaotic good
			contribute_percentage = (rand.NormFloat64() * float64(chaotic_good_standard_deviation)) + float64(neutral_good_mean)

		}

	} else if mi.evilness == 3 { // if agent is evil
		if mi.chaoticness == 1 { // if lawful evil

			contribute_percentage = (rand.NormFloat64() * float64(lawful_evil_standard_deviation)) + float64(lawful_evil_mean)

		} else if mi.chaoticness == 2 { // neutral evil
			contribute_percentage = (rand.NormFloat64() * float64(neutral_evil_standard_deviation)) + float64(neutral_evil_mean)

		} else if mi.chaoticness == 3 { //chaotic evil
			contribute_percentage = (rand.NormFloat64() * float64(chaotic_evil_standard_deviation)) + float64(neutral_evil_mean)

		}
	}
	contribute_percentage = Apply_mood(contribute_percentage)

	mi.intendedContribution = max(int(math.Round(float64(contribute_percentage)*float64(mi.Score))), mi.Score)

	// how much to declare:
	// if you contributed less, there is no point to lie ( if audition checks against the expected contribution)
	mi.declaredcontribution = max(mi.intendedContribution, mi.AoAExpectedContribution)

	return mi.intendedContribution
}

func (mi *MI_256_v1) GetStatedContribution(instance common.IExtendedAgent) int {
	return mi.declaredcontribution
}

// Withdrawal Strategy
func (mi *MI_256_v1) DecideWithdrawal() int {
	mi.IntendedWithdrawal = 0
	Withdrawal_percentage := 0.0
	// CurrentScore = mi.Score

	mood_modifier := float64(1.0 / 32.0 * float64(mi.mood))
	// we scale from 0-2*AoAWithdrawal
	neutral_mean := float64(mi.AoAExpectedWithdrawal / (2*mi.AoAExpectedWithdrawal + 1)) //normalize from 0-1
	//we model evil and good with different standard deviations, with chaotic being high standard deviation,
	lawful_evil_mean := neutral_mean * 9 / 8
	neutral_evil_mean := neutral_mean/4 + neutral_mean
	neutral_good_mean := neutral_mean * 7.0 / 8
	lawful_good_mean := neutral_mean * 3.0 / 4
	neutral_standard_deviation := float64(1.0 / 12)
	lawful_standard_deviation := 1.0 / 16.0
	chaotic_standard_deviation := 1.0 / 8
	chaotic_good_standard_deviation := 2.0 / 8 * (1 - neutral_mean)
	chaotic_evil_standard_deviation := 2.0 / 8 * neutral_mean
	neutral_good_standard_deviation := 1.0 / 8 * (1 - neutral_mean)
	neutral_evil_standard_deviation := 1.0 / 8 * neutral_mean
	lawful_good_standard_deviation := 1.0 / 16 * (1 - neutral_mean)
	lawful_evil_standard_deviation := 1.0 / 16 * neutral_mean

	Apply_mood := func(Withdrawal_percentage float64) float64 {
		if Withdrawal_percentage > float64(neutral_mean) {
			return Withdrawal_percentage + mood_modifier
		} else {
			return Withdrawal_percentage - mood_modifier
		}
	}
	if mi.evilness == 2 { // if agent has neutral evilness, use neutral mean
		if mi.chaoticness == 1 { // if lawful neutral

			Withdrawal_percentage = (rand.NormFloat64() * float64(lawful_standard_deviation)) + float64(neutral_mean)

		} else if mi.chaoticness == 2 { // abs neutral
			Withdrawal_percentage = (rand.NormFloat64() * float64(neutral_standard_deviation)) + float64(neutral_mean)

		} else if mi.chaoticness == 3 { //chaotic neutral
			Withdrawal_percentage = (rand.NormFloat64() * float64(chaotic_standard_deviation)) + float64(neutral_mean)

		}

	} else if mi.evilness == 1 { // if agent is good
		if mi.chaoticness == 1 { // if lawful good

			Withdrawal_percentage = (rand.NormFloat64() * float64(lawful_good_standard_deviation)) + float64(lawful_good_mean)

		} else if mi.chaoticness == 2 { // neutral good
			Withdrawal_percentage = (rand.NormFloat64() * float64(neutral_good_standard_deviation)) + float64(neutral_good_mean)

		} else if mi.chaoticness == 3 { //chaotic good
			Withdrawal_percentage = (rand.NormFloat64() * float64(chaotic_good_standard_deviation)) + float64(neutral_good_mean)

		}

	} else if mi.evilness == 3 { // if agent is evil
		if mi.chaoticness == 1 { // if lawful evil

			Withdrawal_percentage = (rand.NormFloat64() * float64(lawful_evil_standard_deviation)) + float64(lawful_evil_mean)

		} else if mi.chaoticness == 2 { // neutral evil
			Withdrawal_percentage = (rand.NormFloat64() * float64(neutral_evil_standard_deviation)) + float64(neutral_evil_mean)

		} else if mi.chaoticness == 3 { //chaotic evil
			Withdrawal_percentage = (rand.NormFloat64() * float64(chaotic_evil_standard_deviation)) + float64(neutral_evil_mean)

		}
	}
	Withdrawal_percentage = Apply_mood(Withdrawal_percentage)

	mi.IntendedWithdrawal = max(int(math.Round(float64(Withdrawal_percentage)*float64(mi.AoAExpectedWithdrawal*2))), 0)

	// how much to declare:
	// if you withdrawed more, there is no point to lie ( if audition checks against the expected )
	mi.declaredWithdrawal = min(mi.IntendedWithdrawal, mi.AoAExpectedWithdrawal)

	// TODO: implement contribution strategy
	return mi.IntendedWithdrawal
}

func (mi *MI_256_v1) GetStatedWithdrawal(instance common.IExtendedAgent) int {
	return mi.declaredWithdrawal
}

// Audit Strategy
func (mi *MI_256_v1) DecideAudit() bool {
	// TODO: implement audit strategy
	return true
}
func (mi *MI_256_v1) DecideVote() bool {
	// TODO: implement vote strategy
	return true
}

// ----------------------- State Helpers -----------------------
// TODO: add helper functions for managing / using internal states

//get common pool resource

// mi.Server.GetTeam(mi.GetID()).GetCommonPool()

// //get ids of people in my team
// mi.Server.GetTeam(mi.GetID()).Agents

// //get the voting status (to implement)

// mi.Server.GetTeam(mi.GetID()).

// ---------------------------------------------------------------
func (mi *MI_256_v1) RandomizeCharacter() {
	mi.chaoticness = rand.Intn(3) + 1
	mi.evilness = rand.Intn(3) + 1
	mi.Initialize_opninions()
	mi.haveIlied = false

}

// ----------- functions that update character opinions ---------------------------------

func (mi *MI_256_v1) Initialize_opninions() {
	mi.affinity = make(map[uuid.UUID]int)
	for _, agent := range mi.Server.GetTeam(mi.GetID()).Agents {
		mi.affinity[agent] = 0
	}
	// needs to change
	mi.mood = 0

}

func (mi *MI_256_v1) startTurnUpdate() {
	// this function updates the agent states every start of turn, refreshing states if needed
	mi.UpdateMoodTurnStart()
	mi.haveIlied = false
}

/*
Mood decides wether the agent want to be more active
Mood needs to be updated at the following timepoints:

1. At the start of turn, where a random Urge is generated
2. Dice rolls, wether if you have gone bust or not
3. Comparing the expected Score of every agent after round end to you
4. If you get audited,

*/

func (mi *MI_256_v1) UpdateMoodTurnStart() {
	//generate random float from -1 to 1 and multiply by chaoticness

	randomUrge := (rand.Float64()*2 - 1) * float64(mi.chaoticness)
	randomUrge = math.Round(randomUrge)
	mi.mood += int(randomUrge)
}
func (mi *MI_256_v1) UpdateMoodAfterRoll() {

	bustMoodModifier := 1
	// if the roll has gone bust
	if mi.LastScore == 0 {
		mi.mood += bustMoodModifier * mi.chaoticness
	}
}

func (mi *MI_256_v1) UpdateMoodAfterRoundEnd() {
	// you calculate everyone's declared net gain,

}
func (mi *MI_256_v1) UpdateMoodAfterAuditionEnd() {
	// if you get audited and has been caught lying, then you probably want to be more greedy to survive
	if mi.IcaughtLying {
		mi.mood = 3 * mi.chaoticness
	}
}

/*
Affinity decides the likelyhood of our Agent will support other agent's decisions

Affinity should be updated when:
1. They declared their contributions and Withdrawals. The Amount
2. When they voted to audit you
3. When someone has failed an audit
4. When agents propose their Ideologies.
*/
func (mi *MI_256_v1) UpdateAffinityAoAProposal() {
	// need to see all of the AoA in detail
}

func (mi *MI_256_v1) UpdateAffinityAfterContribute() {
	//this function is called
	// we need to value if the declared contribution is fair, we do not know what everyone rolled, we only know their declared contribution against expected
	// apart from the fairness, there is also the satisfaction. Good people would tend to "understand" if they are not contributing as much, while evil people would want people to donate more
	// and of course, both sides would love people that donates more, more for good people, less for evil people because their is a cheat probability

	// contructing a socially fair structure:
	// simple, meeting the AoA Expected Contribution==fair
	for _, agent := range mi.Server.GetTeam(mi.GetID()).Agents {
		agentExpected := mi.teamAgentsExpectedContribution[agent]
		agentDeclared := mi.teamAgentsDeclaredContribution[agent]
		affinityChange := 0.0
		// fairness, if they contributed more or less than expected, less take note, more don't do anything
		fairnessThresh := 0.5
		if agentDeclared < agentExpected {
			affinityChange -= float64(agentExpected-agentDeclared) * (fairnessThresh)
		}
		// satisfaction ( not too important as if they miss they will cheat anyways so you can't tell)
		satisfactionThresh := 0.5
		if agentDeclared < agentExpected {
			amount_less := float64(agentExpected - agentDeclared)
			affinityChange -= (satisfactionThresh) * float64(mi.evilness-1) * math.Sqrt(amount_less)
		} else {
			amount_more := -float64(agentExpected - agentDeclared)
			affinityChange += (satisfactionThresh) * float64(3-(mi.evilness-1)) * math.Sqrt(amount_more)
		}
		mi.affinity[agent] += int(affinityChange)

	}

}
func (mi *MI_256_v1) UpdateAffinityAfterWithdraw() {
	//similar to contribution, there withdrawing same amount is fair, and satisfaction comes into play
	for _, agent := range mi.Server.GetTeam(mi.GetID()).Agents {
		agentExpected := mi.teamAgentsExpectedWithdraw[agent]
		agentDeclared := mi.teamAgentsDeclaredWithdraw[agent]
		affinityChange := 0.0
		// fairness, if they withdrawed more or less than expected, less take note, more don't do anything
		fairnessThresh := 0.5
		if agentDeclared > agentExpected {
			affinityChange += float64(agentExpected-agentDeclared) * (fairnessThresh)
		}
		// satisfaction ( not too important as if they miss they will cheat anyways so you can't tell)
		satisfactionThresh := 0.5
		if agentDeclared < agentExpected {
			amount_less := float64(agentExpected - agentDeclared)
			affinityChange += (satisfactionThresh) * float64(mi.evilness-1) * math.Sqrt(amount_less)
		} else {
			amount_more := -float64(agentExpected - agentDeclared)
			affinityChange -= (satisfactionThresh) * float64(3-(mi.evilness-1)) * math.Sqrt(amount_more)
		}
		mi.affinity[agent] += int(affinityChange)

	}

}
func (mi *MI_256_v1) UpdateAffinityAfterVote() {
	// whenever a vote happens, if the vote is targeting you, someone you really dislike, someone you really like, you would change affinity based on people's votes.

	if mi.lastAuditTarget == mi.GetID() { // if the target agent is yourself:
		//then you would really not like the guy that targeted you, and not like whoever voted yes to vote you out
		for _, agent := range mi.Server.GetTeam(mi.GetID()).Agents {
			affinityChange := 0
			if mi.lastAuditStarter == agent {
				affinityChange -= 5
			}
			voteAgainstAffinityChange := 3
			voteForAffinityChange := 1
			if mi.lastVotes[agent] == true {
				//affninity changes based on the chaoticness, the more chaotic, the more you cannot stand the person
				affinityChange -= voteAgainstAffinityChange * mi.chaoticness
			} else {
				affinityChange += voteForAffinityChange * mi.chaoticness
			}
			mi.affinity[agent] += affinityChange
		}

	} else if mi.affinity[mi.lastAuditTarget] <= -10 { // if the vote is against someone you really dislike, you gain little change for people agreeing to it, but dislike people who disagree with it
		for _, agent := range mi.Server.GetTeam(mi.GetID()).Agents {
			affinityChange := 0
			if mi.lastAuditStarter == agent {

				affinityChange += 2
			}

			voteAgainstAffinityChange := 2
			voteForAffinityChange := 1
			if mi.lastVotes[agent] == true {
				//affninity changes based on the chaoticness, the more chaotic, the more you cannot stand the person
				affinityChange += voteForAffinityChange * mi.chaoticness
			} else {
				affinityChange -= voteAgainstAffinityChange * mi.chaoticness
			}
			mi.affinity[agent] += affinityChange
		}
	} else if mi.affinity[mi.lastAuditTarget] <= -10 { // if the vote is against someone you really like, you gain little dislike for people agreeing to it, but like people who disagree with it
		for _, agent := range mi.Server.GetTeam(mi.GetID()).Agents {
			affinityChange := 0
			if mi.lastAuditStarter == agent {
				if mi.isThereCheatThisRound {
					affinityChange -= 2
				} else {
					affinityChange -= 3 // you would get more mad if there seems to be no cheating this round
				}

			}
			voteAgainstAffinityChange := 0
			voteForAffinityChange := 0
			if mi.isThereCheatThisRound {
				voteAgainstAffinityChange = 1
				voteForAffinityChange = 2
			} else {
				voteAgainstAffinityChange = 1
				voteForAffinityChange = 3
			}

			if mi.lastVotes[agent] == true {
				//affninity changes based on the chaoticness, the more chaotic, the more you cannot stand the person
				affinityChange -= voteForAffinityChange * mi.chaoticness
			} else {
				affinityChange += voteAgainstAffinityChange * mi.chaoticness
			}
			mi.affinity[agent] += affinityChange
		}
	}

}
func (mi *MI_256_v1) UpdateAffinityAfterAudit() {
	// not used currently, intergrated into vote

}
