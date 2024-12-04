package agents

import (
	"log"
	"math/rand"

	"github.com/google/uuid"

	gameRecorder "github.com/ADimoska/SOMASExtended/gameRecorder"

	common "github.com/ADimoska/SOMASExtended/common"

	// TODO:

	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/agent"
	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/message"
)

type ExtendedAgent struct {
	*agent.BaseAgent[common.IExtendedAgent]
	Server common.IServer
	Score  int
	TeamID uuid.UUID

	// private
	LastScore int

	// debug
	VerboseLevel int

	// AoA vote
	AoARanking []int

	/* Team Ranking - Tracks the teams that the agent would like to join (if /
	 * when it is an orphan) in order of preference. Lowest index = highest
	 * priority. If this is empty, the server will not attempt to allocate the
	 * orphan to a team. */
	TeamRanking []uuid.UUID

	LastTeamID uuid.UUID // Tracks the last team the agent was part of

	// for recording purpose
	TrueSomasTeamID int // your true team id! e.g. team 4 -> 4. Override this in your agent constructor
}

type AgentConfig struct {
	InitScore    int
	VerboseLevel int
}

func GetBaseAgents(funcs agent.IExposedServerFunctions[common.IExtendedAgent], configParam AgentConfig) *ExtendedAgent {
	return &ExtendedAgent{
		BaseAgent:    agent.CreateBaseAgent(funcs),
		Server:       funcs.(common.IServer), // Type assert the server functions to IServer interface
		Score:        configParam.InitScore,
		VerboseLevel: configParam.VerboseLevel,
		AoARanking:   []int{0},
		TeamRanking:  []uuid.UUID{},
	}
}

// ----------------------- Interface implementation -----------------------

// Get the agent's current team ID
func (mi *ExtendedAgent) GetTeamID() uuid.UUID {
	return mi.TeamID
}

// Get the agent's last team ID
func (mi *ExtendedAgent) GetLastTeamID() uuid.UUID {
	return mi.LastTeamID
}

// Get the agent's current score
// Can only be called by the server (otherwise other agents will see their true score)
func (mi *ExtendedAgent) GetTrueScore() int {
	return mi.Score
}

// Get the agent's true team ID
func (mi *ExtendedAgent) GetTrueSomasTeamID() int {
	return mi.TrueSomasTeamID
}

// Setter for the server to call, in order to set the true score for this agent
func (mi *ExtendedAgent) SetTrueScore(score int) {
	mi.Score = score
}

// custom function: ask for rolling the dice
func (mi *ExtendedAgent) StartRollingDice(instance common.IExtendedAgent) {
	if mi.VerboseLevel > 10 {
		log.Printf("%s is rolling the Dice\n", mi.GetID())
	}
	if mi.VerboseLevel > 9 {
		log.Println("---------------------")
	}
	// TODO: implement the logic in environment, do a random of 3d6 now with 50% chance to stick
	mi.LastScore = -1
	rounds := 1
	turnScore := 0

	willStick := false

	// loop until not stick
	for !willStick {
		// debug add score directly
		currentScore := Roll3Dice()

		// check if currentScore is higher than lastScore
		if currentScore > mi.LastScore {
			turnScore += currentScore
			mi.LastScore = currentScore
			willStick = instance.StickOrAgain(turnScore, currentScore)
			if willStick {
				mi.DecideStick() //used just for debugging
				break
			}
			mi.DecideRollAgain() //used just for debugging
		} else {
			// burst, lose all turn score
			if mi.VerboseLevel > 4 {
				log.Printf("%s **BURSTED!** round: %v, current score: %v\n", mi.GetID(), rounds, currentScore)
			}
			turnScore = 0
			break
		}

		rounds++
	}

	// add turn score to total score
	mi.Score += turnScore

	if mi.VerboseLevel > 4 {
		log.Printf("%s's turn score: %v, total score: %v\n", mi.GetID(), turnScore, mi.Score)
	}
}

// stick or again
func (mi *ExtendedAgent) StickOrAgain(accumulatedScore int, prevRoll int) bool {
	// if mi.verboseLevel > 8 {
	// 	log.Printf("%s is deciding to stick or again\n", mi.GetID())
	// }
	return rand.Intn(2) == 0
}

// decide to stick
func (mi *ExtendedAgent) DecideStick() {
	if mi.VerboseLevel > 6 {
		log.Printf("%s decides to [STICK], last score: %v\n", mi.GetID(), mi.LastScore)
	}
}

// decide to roll again
func (mi *ExtendedAgent) DecideRollAgain() {
	if mi.VerboseLevel > 6 {
		log.Printf("%s decides to ROLL AGAIN, last score: %v\n", mi.GetID(), mi.LastScore)
	}
}

// TODO: TO BE IMPLEMENTED BY TEAM'S AGENT
// get the agent's actual contribution to the common pool
// This function MUST return the same value when called multiple times in the same turn
func (mi *ExtendedAgent) GetActualContribution(instance common.IExtendedAgent) int {
	if mi.HasTeam() {
		contribution := instance.DecideContribution()
		if mi.VerboseLevel > 6 {
			log.Printf("%s is contributing %d to the common pool and thinks the common pool size is %d\n", mi.GetID(), contribution, mi.Server.GetTeam(mi.GetID()).GetCommonPool())
		}
		return contribution
	} else {
		if mi.VerboseLevel > 6 {
			log.Printf("%s has no team, skipping contribution\n", mi.GetID())
		}
		return 0
	}
}

func (mi *ExtendedAgent) DecideContribution() int {
	// first check if the agent has a team
	if !mi.HasTeam() {
		return 0
	}
	// MVP: contribute exactly as defined in AoA
	if mi.Server.GetTeam(mi.GetID()).TeamAoA != nil {
		aoaExpectedContribution := mi.Server.GetTeam(mi.GetID()).TeamAoA.GetExpectedContribution(mi.GetID(), mi.GetTrueScore())
		// double check if score in agent is sufficient (this should be handled by AoA though)
		if mi.GetTrueScore() < aoaExpectedContribution {
			return mi.GetTrueScore() // give all score if less than expected
		}
		return aoaExpectedContribution
	} else {
		if mi.VerboseLevel > 6 {
			// should not happen!
			log.Printf("[WARNING] Agent %s has no AoA, contributing 0\n", mi.GetID())
		}
		return 0
	}
}

// get the agent's stated contribution to the common pool
// TODO: the value returned by this should be broadcasted to the team via a message
// This function MUST return the same value when called multiple times in the same turn
func (mi *ExtendedAgent) GetStatedContribution(instance common.IExtendedAgent) int {
	// first check if the agent has a team
	if !mi.HasTeam() {
		return 0
	}

	// Hardcoded stated
	statedContribution := instance.GetActualContribution(instance)
	return statedContribution
}

// make withdrawal from common pool
func (mi *ExtendedAgent) GetActualWithdrawal(instance common.IExtendedAgent) int {
	// first check if the agent has a team
	if !mi.HasTeam() {
		return 0
	}
	currentPool := mi.Server.GetTeam(mi.GetID()).GetCommonPool()
	withdrawal := instance.DecideWithdrawal()
	log.Printf("%s is withdrawing %d from the common pool of size %d\n", mi.GetID(), withdrawal, currentPool)
	return withdrawal
}

// The value returned by this should be broadcasted to the team via a message
// This function MUST return the same value when called multiple times in the same turn
func (mi *ExtendedAgent) GetStatedWithdrawal(instance common.IExtendedAgent) int {
	// first check if the agent has a team
	if !mi.HasTeam() {
		return 0
	}
	// Currently, assume stated withdrawal matches actual withdrawal
	return instance.DecideWithdrawal()
}

// Decide the withdrawal amount based on AoA and current pool size
func (mi *ExtendedAgent) DecideWithdrawal() int {
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
		return aoaExpectedWithdrawal
	} else {
		if mi.VerboseLevel > 6 {
			log.Printf("[WARNING] Agent %s has no AoA, withdrawing 0\n", mi.GetID())
		}
		return 0
	}
}

/*
Provide agentId for memory, current accumulated score
(to see if above or below predicted threshold for common pool contribution)
And previous roll in case relevant
*/
func (mi *ExtendedAgent) StickOrAgainFor(agentId uuid.UUID, accumulatedScore int, prevRoll int) int {
	// random chance, to simulate what is already implemented
	return rand.Intn(2)
}

// dev function
func (mi *ExtendedAgent) LogSelfInfo() {
	log.Printf("[Agent %s] score: %v\n", mi.GetID(), mi.Score)
}

// Agent returns their preference for an audit on contribution
// 0: No preference
// 1: Prefer audit
// -1: Prefer no audit
func (mi *ExtendedAgent) GetContributionAuditVote() common.Vote {
	return common.CreateVote(0, mi.GetID(), uuid.Nil)
}

// Agent returns their preference for an audit on withdrawal
// 0: No preference
// 1: Prefer audit
// -1: Prefer no audit
func (mi *ExtendedAgent) GetWithdrawalAuditVote() common.Vote {
	return common.CreateVote(0, mi.GetID(), uuid.Nil)
}

func (mi *ExtendedAgent) SetAgentContributionAuditResult(agentID uuid.UUID, result bool) {}

func (mi *ExtendedAgent) SetAgentWithdrawalAuditResult(agentID uuid.UUID, result bool) {}

// ----Withdrawal------- Messaging functions -----------------------

func (mi *ExtendedAgent) HandleTeamFormationMessage(msg *common.TeamFormationMessage) {
	log.Printf("Agent %s received team forming invitation from %s\n", mi.GetID(), msg.GetSender())

	// Already in a team - reject invitation
	if mi.TeamID != (uuid.UUID{}) {
		if mi.VerboseLevel > 6 {
			log.Printf("Agent %s rejected invitation from %s - already in team %v\n",
				mi.GetID(), msg.GetSender(), mi.TeamID)
		}
		return
	}

	// Handle team creation/joining based on sender's team status
	sender := msg.GetSender()
	if mi.Server.CheckAgentAlreadyInTeam(sender) {
		existingTeamID := mi.Server.AccessAgentByID(sender).GetTeamID()
		mi.joinExistingTeam(existingTeamID)
	} else {
		mi.createNewTeam(sender)
	}
}

func (mi *ExtendedAgent) HandleContributionMessage(msg *common.ContributionMessage) {
	if mi.VerboseLevel > 8 {
		log.Printf("Agent %s received contribution notification from %s: amount=%d\n",
			mi.GetID(), msg.GetSender(), msg.StatedAmount)
	}

	// Team's agent should implement logic to store or process the reported contribution amount as desired
}

func (mi *ExtendedAgent) HandleScoreReportMessage(msg *common.ScoreReportMessage) {
	if mi.VerboseLevel > 8 {
		log.Printf("Agent %s received score report from %s: score=%d\n",
			mi.GetID(), msg.GetSender(), msg.TurnScore)
	}

	// Team's agent should implement logic to store or process score of other agents as desired
}

func (mi *ExtendedAgent) HandleWithdrawalMessage(msg *common.WithdrawalMessage) {
	if mi.VerboseLevel > 8 {
		log.Printf("Agent %s received withdrawal notification from %s: amount=%d\n",
			mi.GetID(), msg.GetSender(), msg.StatedAmount)
	}

	// Team's agent should implement logic to store or process the reported withdrawal amount as desired
}

func (mi *ExtendedAgent) HandleAgentOpinionRequestMessage(msg *common.AgentOpinionRequestMessage) {
	// Team's agent should implement logic to respond to opinion request as desired
	log.Printf("Agent %s received opinion request from %s\n", mi.GetID(), msg.AgentID)
	opinion := 70
	opinionResponseMsg := mi.CreateAgentOpinionResponseMessage(msg.AgentID, opinion)
	log.Printf("Sending opinion response to %s\n", msg.AgentID)
	mi.SendMessage(opinionResponseMsg, msg.AgentID) // Sent asynchronously, because this is "extra information"
}

func (mi *ExtendedAgent) HandleAgentOpinionResponseMessage(msg *common.AgentOpinionResponseMessage) {
	// Team's agent should implement logic to store or process opinion response as desired
	log.Printf("Agent %s received opinion response from %s: opinion=%d\n", mi.GetID(), msg.GetSender(), msg.AgentOpinion)
}

func (mi *ExtendedAgent) BroadcastSyncMessageToTeam(msg message.IMessage[common.IExtendedAgent]) {
	// Send message to all team members synchronously
	agentsInTeam := mi.Server.GetAgentsInTeam(mi.TeamID)
	for _, agentID := range agentsInTeam {
		if agentID != mi.GetID() {
			mi.SendSynchronousMessage(msg, agentID)
		}
	}
}

func (mi *ExtendedAgent) StateContributionToTeam(instance common.IExtendedAgent) {
	// Broadcast contribution to team
	statedContribution := instance.GetStatedContribution(instance)
	contributionMsg := mi.CreateContributionMessage(statedContribution)
	mi.BroadcastSyncMessageToTeam(contributionMsg)
}

func (mi *ExtendedAgent) StateWithdrawalToTeam(instance common.IExtendedAgent) {
	// Broadcast withdrawal to team
	statedWithdrawal := instance.GetStatedWithdrawal(instance)
	withdrawalMsg := mi.CreateWithdrawalMessage(statedWithdrawal)
	mi.BroadcastSyncMessageToTeam(withdrawalMsg)
}

// ----------------------- Info functions -----------------------
func (mi *ExtendedAgent) GetExposedInfo() common.ExposedAgentInfo {
	return common.ExposedAgentInfo{
		AgentUUID:   mi.GetID(),
		AgentTeamID: mi.TeamID,
	}
}

func (mi *ExtendedAgent) CreateScoreReportMessage() *common.ScoreReportMessage {
	return &common.ScoreReportMessage{
		BaseMessage: mi.CreateBaseMessage(),
		TurnScore:   mi.LastScore,
	}
}

func (mi *ExtendedAgent) CreateContributionMessage(statedAmount int) *common.ContributionMessage {
	return &common.ContributionMessage{
		BaseMessage:  mi.CreateBaseMessage(),
		StatedAmount: statedAmount,
	}
}

func (mi *ExtendedAgent) CreateWithdrawalMessage(statedAmount int) *common.WithdrawalMessage {
	return &common.WithdrawalMessage{
		BaseMessage:  mi.CreateBaseMessage(),
		StatedAmount: statedAmount,
	}
}

func (mi *ExtendedAgent) CreateAgentOpinionRequestMessage(agentID uuid.UUID) *common.AgentOpinionRequestMessage {
	return &common.AgentOpinionRequestMessage{
		BaseMessage: mi.CreateBaseMessage(),
		AgentID:     agentID,
	}
}

func (mi *ExtendedAgent) CreateAgentOpinionResponseMessage(agentID uuid.UUID, opinion int) *common.AgentOpinionResponseMessage {
	return &common.AgentOpinionResponseMessage{
		BaseMessage:  mi.CreateBaseMessage(),
		AgentID:      agentID,
		AgentOpinion: opinion,
	}
}

// ----------------------- Debug functions -----------------------

func Roll3Dice() int {
	// row 3d6
	total := 0
	for i := 0; i < 3; i++ {
		total += rand.Intn(6) + 1
	}
	return total
}

// func Debug_StickOrAgainJudgement() bool {
// 	// 50% chance to stick
// 	return rand.Intn(2) == 0
// }

// ----------------------- Team forming functions -----------------------
func (mi *ExtendedAgent) StartTeamForming(instance common.IExtendedAgent, agentInfoList []common.ExposedAgentInfo) {
	// TODO: implement team forming logic
	if mi.VerboseLevel > 6 {
		log.Printf("%s is starting team formation\n", mi.GetID())
	}

	chosenAgents := instance.DecideTeamForming(agentInfoList)
	mi.SendTeamFormingInvitation(chosenAgents)
	mi.SignalMessagingComplete()
}

func (mi *ExtendedAgent) DecideTeamForming(agentInfoList []common.ExposedAgentInfo) []uuid.UUID {
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

	// random choice from the invitation list
	rand.Shuffle(len(invitationList), func(i, j int) { invitationList[i], invitationList[j] = invitationList[j], invitationList[i] })
	if len(invitationList) == 0 {
		return []uuid.UUID{}
	}
	chosenAgent := invitationList[0]

	// Return a slice containing the chosen agent
	return []uuid.UUID{chosenAgent}
}

func (mi *ExtendedAgent) SendTeamFormingInvitation(agentIDs []uuid.UUID) {
	for _, agentID := range agentIDs {
		invitationMsg := &common.TeamFormationMessage{
			BaseMessage: mi.CreateBaseMessage(),
			AgentInfo:   mi.GetExposedInfo(),
			Message:     "Would you like to form a team?",
		}
		// Debug print to check message contents
		log.Printf("Sending invitation: sender=%v, teamID=%v, receiver=%v\n", mi.GetID(), mi.GetTeamID(), agentID)
		mi.SendSynchronousMessage(invitationMsg, agentID)
	}
}

func (mi *ExtendedAgent) createNewTeam(senderID uuid.UUID) {
	log.Printf("Agent %s is creating a new team\n", mi.GetID())
	teamIDs := []uuid.UUID{mi.GetID(), senderID}
	newTeamID := mi.Server.CreateAndInitTeamWithAgents(teamIDs)

	if newTeamID == (uuid.UUID{}) {
		if mi.VerboseLevel > 6 {
			log.Printf("Agent %s failed to create a new team\n", mi.GetID())
		}
		return
	}

	mi.TeamID = newTeamID
	if mi.VerboseLevel > 6 {
		log.Printf("Agent %s created a new team with ID %v\n", mi.GetID(), newTeamID)
	}
}

func (mi *ExtendedAgent) joinExistingTeam(teamID uuid.UUID) {
	mi.TeamID = teamID
	mi.Server.AddAgentToTeam(mi.GetID(), teamID)
	if mi.VerboseLevel > 6 {
		log.Printf("Agent %s joined team %v\n", mi.GetID(), teamID)
	}
}

// SetTeamID assigns a new team ID to the agent
// Parameters:
//   - teamID: The UUID of the team to assign to this agent
func (mi *ExtendedAgent) SetTeamID(teamID uuid.UUID) {
	// Store the previous team ID
	mi.LastTeamID = mi.TeamID
	mi.TeamID = teamID
}

func (mi *ExtendedAgent) HasTeam() bool {
	return mi.TeamID != (uuid.UUID{})
}

// In RunStartOfIteration, the server loops through each agent in each team
// and sets the teams AoA by majority vote from the agents in that team.
func (mi *ExtendedAgent) SetAoARanking(Preferences []int) {
	mi.AoARanking = Preferences
}

func (mi *ExtendedAgent) GetAoARanking() []int {
	return mi.AoARanking
}

/*
* Decide whether to allow an agent into the team. This will be part of the group
* strategy, and should be implemented by individual groups. During testing this
* function is mocked.
 */
func (mi *ExtendedAgent) VoteOnAgentEntry(candidateID uuid.UUID) bool {
	// TODO: Implement strategy for accepting an agent into the team.
	// Return true to accept them, false to not accept them.
	return true
}

// Return the team ranking
func (mi *ExtendedAgent) GetTeamRanking() []uuid.UUID {
	return mi.TeamRanking
}

// Set the team ranking of which teams this agent would like to join - lower
// index = higher priority. This can be updated as the game goes on, the server
// will only act on this information when the agent is orphaned.
func (mi *ExtendedAgent) SetTeamRanking(teamRanking []uuid.UUID) {
	mi.TeamRanking = teamRanking
}

// ----------------------- Data Recording Functions -----------------------
func (mi *ExtendedAgent) RecordAgentStatus(instance common.IExtendedAgent) gameRecorder.AgentRecord {
	record := gameRecorder.NewAgentRecord(
		instance.GetID(),
		instance.GetTrueSomasTeamID(),
		instance.GetTrueScore(),
		instance.GetStatedContribution(instance),
		instance.GetActualContribution(instance),
		instance.GetActualWithdrawal(instance),
		instance.GetStatedWithdrawal(instance),
		instance.GetTeamID(),
	)
	return record
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