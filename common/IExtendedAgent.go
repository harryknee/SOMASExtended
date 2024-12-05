package common

import (
	gameRecorder "github.com/ADimoska/SOMASExtended/gameRecorder"

	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/agent"
	"github.com/google/uuid"

	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/message"
)

type IExtendedAgent interface {
	agent.IAgent[IExtendedAgent]

	// Getters
	GetTeamID() uuid.UUID
	GetLastTeamID() uuid.UUID
	GetTrueScore() int
	GetTeamRanking() []uuid.UUID

	// Functions that involve strategic decisions
	StartTeamForming(instance IExtendedAgent, agentInfoList []ExposedAgentInfo)
	StartRollingDice(instance IExtendedAgent)
	GetActualContribution(instance IExtendedAgent) int
	GetActualWithdrawal(instance IExtendedAgent) int
	GetStatedContribution(instance IExtendedAgent) int
	GetStatedWithdrawal(instance IExtendedAgent) int

	// Setters
	SetTeamID(teamID uuid.UUID)
	SetTrueScore(score int)
	SetAgentContributionAuditResult(agentID uuid.UUID, result bool)
	SetAgentWithdrawalAuditResult(agentID uuid.UUID, result bool)
	SetTeamRanking(teamRanking []uuid.UUID)
	DecideStick()
	DecideRollAgain()

	// Strategic decisions (functions that each team can implement their own)
	// NOTE: Any function calling these should have a parameter of type IExtendedAgent (instance IExtendedAgent)
	DecideTeamForming(agentInfoList []ExposedAgentInfo) []uuid.UUID
	StickOrAgain(accumulatedScore int, prevRoll int) bool
	DecideContribution() int
	DecideWithdrawal() int
	VoteOnAgentEntry(candidateID uuid.UUID) bool
	StickOrAgainFor(agentId uuid.UUID, accumulatedScore int, prevRoll int) int

	// Messaging functions
	HandleTeamFormationMessage(msg *TeamFormationMessage)
	HandleScoreReportMessage(msg *ScoreReportMessage)
	HandleWithdrawalMessage(msg *WithdrawalMessage)
	BroadcastSyncMessageToTeam(msg message.IMessage[IExtendedAgent])
	HandleContributionMessage(msg *ContributionMessage)
	StateContributionToTeam(instance IExtendedAgent)
	StateWithdrawalToTeam(instance IExtendedAgent)

	// Info
	GetExposedInfo() ExposedAgentInfo
	CreateScoreReportMessage() *ScoreReportMessage
	CreateContributionMessage(statedAmount int) *ContributionMessage
	CreateWithdrawalMessage(statedAmount int) *WithdrawalMessage
	LogSelfInfo()
	GetAoARanking() []int
	SetAoARanking(Preferences []int)
	GetContributionAuditVote() Vote
	GetWithdrawalAuditVote() Vote

	// Data Recording
	RecordAgentStatus(instance IExtendedAgent) gameRecorder.AgentRecord

	//Agent internal state update
	UpdateStateAfterContribution()
	UpdateStateAfterWithdrawal()
	UpdateStateAfterContributionAudit()
	UpdateStateAfterWithdrawalAudit()
	UpdateStateAfterRoll()
	UpdateStateTurnend()
	InitializeStartofTurn()
}
