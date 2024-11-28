package environmentServer

import (
	"SOMAS_Extended/agents"
	"SOMAS_Extended/common"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/server"
)

type EnvironmentServer struct {
	*server.BaseServer[common.IExtendedAgent]

	teamsMutex    sync.RWMutex
	agentInfoList []common.ExposedAgentInfo
	teams         map[uuid.UUID]common.Team

	roundScoreThreshold int
	deadAgents          []common.IExtendedAgent

	// set of options for team strategies (agents rank these options)
	aoaMenu []*common.ArticlesOfAssociation
}

// overrides that requires implementation
func (cs *EnvironmentServer) RunTurn(i, j int) {
	fmt.Printf("\nIteration %v, Turn %v, current agent count: %v\n", i, j, len(cs.GetAgentMap()))
	if j == 1 {
		// take votes at team level and allocate Strategy.
		time.Sleep(10 * time.Second)

		cs.AllocateAoAs()
	}
	if j == 0 {
		cs.StartAgentTeamForming()

	} else { // debug roll dice for agents
		for _, agent := range cs.GetAgentMap() {
			if !cs.IsAgentDead(agent.GetID()) { // only agents that are alive can roll dice
				agent.StartRollingDice()
				team := cs.teams[agent.GetTeamID()]
				agentContribution := agent.ContributeToCommonPool()
				team.CommonPool += agentContribution
				cs.teams[agent.GetTeamID()] = team
				agent.SetTrueScore(agent.GetTrueScore() - agentContribution)
			}
		}
		cs.UpdateCommonPools()
		for _, agent := range cs.GetAgentMap() {
			if !cs.IsAgentDead(agent.GetID()) {
				team := cs.teams[agent.GetTeamID()]
				agentWithdrawal := agent.WithdrawFromCommonPool()
				team.CommonPool -= agentWithdrawal
				cs.teams[agent.GetTeamID()] = team
				agent.SetTrueScore(agent.GetTrueScore() + agentWithdrawal)
			}
		}
	}
}

func (cs *EnvironmentServer) RunStartOfIteration(iteration int) {
	fmt.Printf("--------Start of iteration %v---------\n", iteration)
	cs.CreateNewRoundScoreThreshold()
	// start team forming

}

// Infers pairwise outcomes from rankings
// Alternative would be needing a mapping of 15 individual pairwise comparisons
func runCopelandVote(team *common.Team, cs *EnvironmentServer) []int {

	pairwiseWins := make(map[string]int)
	copelandScores := make(map[byte]float64)

	fmt.Printf("Starting Copeland Vote for Team %s with %d members.\n", team.TeamID, len(team.Agents))
	// Loop through each agent in the team

	for _, agent := range team.Agents {

		agentRanking := cs.GetAgentMap()[agent].GetAoARanking()

		fmt.Printf("Agent %s has the following AoA rankings:\n", agent)
		fmt.Println(agentRanking)

		// Loop through each pair of ranked candidates and perform pairwise comparison
		for i := 0; i < len(agentRanking); i++ {
			for j := i + 1; j < len(agentRanking); j++ {
				if agentRanking[i] < agentRanking[j] {

					pair := []int{agentRanking[i], agentRanking[j]}

					pairKey := fmt.Sprintf("%d%d", pair[0], pair[1])

					fmt.Printf("Agent %s: Comparing candidates %d and %d. Winner: %d\n", agent, pair[0], pair[1], pair[0])

					pairwiseWins[pairKey]++
				} else {

					pair := []int{agentRanking[j], agentRanking[i]}

					pairKey := fmt.Sprintf("%d%d", pair[0], pair[1])

					fmt.Printf("Agent %s: Comparing candidates %d and %d. Winner: %d\n", agent, pair[1], pair[0], pair[1])

					pairwiseWins[pairKey] -= 1
				}

			}
		}
	}

	fmt.Println(pairwiseWins)
	for pair, score := range pairwiseWins {
		// Subtract ASCII value of 0
		candidate1 := pair[0] - 48
		candidate2 := pair[1] - 48

		fmt.Printf("Processing pair %s (candidate 1: %d, candidate 2: %d), score: %d\n", pair, candidate1, candidate2, score)

		if score > 0 {
			copelandScores[candidate1] += 1
			fmt.Printf("Candidate %d wins, Copeland score updated: %v\n", candidate1, copelandScores[candidate1])

		} else if score < 0 {
			copelandScores[candidate2] += 1
			fmt.Printf("Candidate %d wins, Copeland score updated: %v\n", candidate2, copelandScores[candidate2])
		} else {
			copelandScores[candidate1] += 0.5
			copelandScores[candidate2] += 0.5
			fmt.Printf("It's a tie! Copeland scores updated: %v, %v\n", copelandScores[candidate1], copelandScores[candidate2])

		}
	}
	fmt.Println(copelandScores)

	var maxScore float64
	var maxCandidates []int
	for key, score := range copelandScores {
		candidate := int(key)
		if score > maxScore {
			maxScore = score
			maxCandidates = []int{candidate}
		} else if score == maxScore {
			maxCandidates = append(maxCandidates, candidate)
		}
	}

	fmt.Printf("\nWinning candidates for Team %s: %v\n", team.TeamID, maxCandidates)

	return maxCandidates
}

// Aggregates scores for candidates returns all candidates who have the highest score
func runBordaVote(team *common.Team, aoaCandidates []int, cs *EnvironmentServer) []int {

	aoaCandidatesSet := make(map[int]struct{})
	for _, candidate := range aoaCandidates {
		aoaCandidatesSet[candidate] = struct{}{}
	}

	voteSum := make(map[int]int) // key = AoA candidate, value = total votes
	n := len(aoaCandidates)      // Could explicitly do n := 6, right now each points allocation is off by len(all_candidates) - len(aoaCandidates)
	for _, agent := range team.Agents {

		agentRanking := cs.GetAgentMap()[agent].GetAoARanking()
		fmt.Printf("Agent %s has the following AoA rankings:\n", agent)
		fmt.Println((agentRanking))

		// Check if the current AoA is a candidate
		// May be better to loop on candidates instead
		for vote, aoa := range agentRanking {
			if _, exists := aoaCandidatesSet[aoa]; exists {
				points := n - vote - 1
				voteSum[aoa] += points
				fmt.Printf("Agent %s votes for AoA %d with %d point\n", agent, aoa, points)
			}
		}
	}

	fmt.Println("\nCandidates scores:")
	fmt.Println(voteSum)
	var filtered []int

	if len(voteSum) == 1 {
		return filtered
	}

	// Initialize maxVotes to the first candidate's score
	maxVotes := voteSum[aoaCandidates[0]]

	// Find the max score and filter candidates with the max score in one pass
	for candidate, score := range voteSum {
		if score > maxVotes {
			maxVotes = score
			// Reset filtered list with the new max score
			filtered = []int{candidate}
		} else if score == maxVotes {
			filtered = append(filtered, candidate)
		}

		fmt.Printf("Processing candidate %d with score %d\n", candidate, score) // Debugging print
	}

	// Remove candidates below a threshold (check if there are ties)
	fmt.Println("\nFiltered candidates after tie removal:")
	fmt.Println(filtered)

	return filtered
}

func (cs *EnvironmentServer) AllocateAoAs() {
	for _, team := range cs.teams {
		winners := runCopelandVote(&team, cs)
		if len(winners) > 1 {
			fmt.Println("Multiple winners detected. Running Borda Vote.")
			winners = runBordaVote(&team, winners, cs)
		}
		// Select random AoA if still tied, else select 'winner'
		if len(winners) > 0 {

			// Create a random number generator with a seed based on current time
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			// Generate random index
			randomI := r.Intn(len(winners))

			team.TeamAoA = cs.aoaMenu[winners[randomI]]

			fmt.Printf("Team %v has AoA: %v\n", team.TeamID, winners[randomI])

		}
	}
}

func (cs *EnvironmentServer) RunEndOfIteration(int) {
	for _, agent := range cs.GetAgentMap() {
		cs.KillAgentBelowThreshold(agent.GetID())
	}
}

// custom override
func (cs *EnvironmentServer) Start() {
	// steal method from package...
	cs.BaseServer.Start()

	// TODO
}

// constructor
func MakeEnvServer(numAgent int, iterations int, turns int, maxDuration time.Duration, maxThread int, agentConfig agents.AgentConfig) *EnvironmentServer {
	serv := &EnvironmentServer{
		BaseServer: server.CreateBaseServer[common.IExtendedAgent](iterations, turns, maxDuration, maxThread),
		teams:      make(map[uuid.UUID]common.Team),
	}
	serv.SetGameRunner(serv)

	// create agents
	// example: Base Agent & MI_256 from team 4

	// dummy agents (base agent)
	for i := 0; i < numAgent; i++ {
		base_agent := agents.GetBaseAgents(serv, agentConfig)
		serv.AddAgent(base_agent)
		serv.AddAgent(base_agent)
		base_agent1 := agents.GetBaseAgents1(serv, agentConfig)
		serv.AddAgent(base_agent1)
		base_agent2 := agents.GetBaseAgents2(serv, agentConfig)
		serv.AddAgent(base_agent2)
		// TEAM 1
		// TEAM 2
		// TEAM 3
		// TEAM 4
		// example: MI_256 from team 4
		team4_agent := agents.Team4_CreateAgent(serv, agentConfig)
		serv.AddAgent(team4_agent)
		// TEAM 5
		// TEAM 6
	}

	serv.aoaMenu = []*common.ArticlesOfAssociation{nil, nil, nil, nil}

	// for now, menu just has 4 choices of AoA. TBC.
	serv.aoaMenu[0] = common.CreateArticlesOfAssociation(
		common.CreateFixedContributionRule(10),
		common.CreateFixedWithdrawalRule(10),
		common.CreateFixedAuditCost(10),
		common.CreateFixedPunishment(10),
	)

	serv.aoaMenu[1] = common.CreateArticlesOfAssociation(
		common.CreateFixedContributionRule(20),
		common.CreateFixedWithdrawalRule(20),
		common.CreateFixedAuditCost(20),
		common.CreateFixedPunishment(20),
	)

	serv.aoaMenu[2] = common.CreateArticlesOfAssociation(
		common.CreateFixedContributionRule(30),
		common.CreateFixedWithdrawalRule(30),
		common.CreateFixedAuditCost(30),
		common.CreateFixedPunishment(30),
	)

	serv.aoaMenu[3] = common.CreateArticlesOfAssociation(
		common.CreateFixedContributionRule(40),
		common.CreateFixedWithdrawalRule(40),
		common.CreateFixedAuditCost(40),
		common.CreateFixedPunishment(40),
	)

	return serv
}

// debug log printing
func (cs *EnvironmentServer) LogAgentStatus() {
	// log agent count, and their scores
	fmt.Printf("Agent count: %v\n", len(cs.GetAgentMap()))
	for _, agent := range cs.GetAgentMap() {
		agent.LogSelfInfo()
	}
	for _, agent := range cs.deadAgents {
		fmt.Printf("Agent %v is dead\n", agent.GetID())
	}
}

// pretty logging to show all team status
func (cs *EnvironmentServer) LogTeamStatus() {
	for _, team := range cs.teams {
		fmt.Printf("Team %v: %v\n", team.TeamID, team.Agents)
	}
	// log agents that have no team
	for _, agent := range cs.GetAgentMap() {
		if agent.GetTeamID() == uuid.Nil {
			fmt.Printf("Agent %v has no team\n", agent.GetID())
		}
	}
}

func (cs *EnvironmentServer) UpdateAndGetAgentExposedInfo() []common.ExposedAgentInfo {
	// clear the list
	cs.agentInfoList = nil
	for _, agent := range cs.GetAgentMap() {
		cs.agentInfoList = append(cs.agentInfoList, agent.GetExposedInfo())
	}
	return cs.agentInfoList
}

// create a new round score threshold
func (cs *EnvironmentServer) CreateNewRoundScoreThreshold() {
	// random one between 10 to 20 (TODO)
	cs.roundScoreThreshold = rand.Intn(10) + 10
	fmt.Printf("[server] New round score threshold: %v\n", cs.roundScoreThreshold)
}

// check agent score
func (cs *EnvironmentServer) KillAgentBelowThreshold(agentID uuid.UUID) int {
	agent := cs.GetAgentMap()[agentID]
	score := agent.GetTrueScore()
	if score < cs.roundScoreThreshold {
		cs.KillAgent(agentID)
	}
	return score
}

// kill agent
func (cs *EnvironmentServer) KillAgent(agentID uuid.UUID) {
	cs.deadAgents = append(cs.deadAgents, cs.GetAgentMap()[agentID])
	cs.RemoveAgent(cs.GetAgentMap()[agentID])
	fmt.Printf("[server] Agent %v killed\n", agentID)
}

// is agent dead
func (cs *EnvironmentServer) IsAgentDead(agentID uuid.UUID) bool {
	for _, deadAgent := range cs.deadAgents {
		if deadAgent.GetID() == agentID {
			return true
		}
	}
	return false
}

// team forming

func (cs *EnvironmentServer) StartAgentTeamForming() {
	// Clear existing teams at the start of team formation
	cs.teamsMutex.Lock()
	cs.teams = make(map[uuid.UUID]common.Team)
	cs.teamsMutex.Unlock()

	// Get updated agent info and let agents form teams
	agentInfo := cs.UpdateAndGetAgentExposedInfo()

	fmt.Printf("------------- [server] Starting team formation -------------\n\n")

	// Launch team formation for each agent
	for _, agent := range cs.GetAgentMap() {
		agent.StartTeamForming(agentInfo)
	}
}

func (cs *EnvironmentServer) CreateTeam() {
	cs.teams = make(map[uuid.UUID]common.Team)
}

func (cs *EnvironmentServer) AddAgentToTeam(agentID uuid.UUID, teamID uuid.UUID) {
	cs.teamsMutex.Lock()
	defer cs.teamsMutex.Unlock()

	// Check if agent is already in this team
	team := cs.teams[teamID]
	for _, existingAgent := range team.Agents {
		if existingAgent == agentID {
			return // Skip if agent already exists
		}
	}

	team.Agents = append(team.Agents, agentID)
	cs.teams[teamID] = team
}

func (cs *EnvironmentServer) GetAgentsInTeam(teamID uuid.UUID) []uuid.UUID {
	cs.teamsMutex.RLock()
	defer cs.teamsMutex.RUnlock()
	return cs.teams[teamID].Agents
}

func (cs *EnvironmentServer) CheckAgentAlreadyInTeam(agentID uuid.UUID) bool {
	cs.teamsMutex.RLock()
	defer cs.teamsMutex.RUnlock()

	for _, team := range cs.teams {
		for _, agent := range team.Agents {
			if agent == agentID {
				return true
			}
		}
	}
	return false
}

func (cs *EnvironmentServer) CreateAndInitTeamWithAgents(agentIDs []uuid.UUID) uuid.UUID {
	// Skip if no agents provided
	if len(agentIDs) == 0 {
		return uuid.UUID{}
	}

	// check if any agent is already in a team
	for _, agentID := range agentIDs {
		if cs.CheckAgentAlreadyInTeam(agentID) {
			fmt.Printf("[server] Agent %v is already in a team\n", agentID)
			return uuid.UUID{}
		}
	}

	// Generate team ID first
	teamID := uuid.New()

	// Protect map write with mutex
	cs.teamsMutex.Lock()
	cs.teams[teamID] = common.Team{
		TeamID: teamID,
		Agents: agentIDs,
	}
	cs.teamsMutex.Unlock()

	// Update each agent's team ID
	for _, agentID := range agentIDs {
		if agent, exists := cs.GetAgentMap()[agentID]; exists {
			agent.SetTeamID(teamID)
		}
	}

	fmt.Printf("[server] Created team %v with agents %v\n", teamID, agentIDs)
	return teamID
}

// agent get team
func (cs *EnvironmentServer) GetTeam(agentID uuid.UUID) common.Team {
	// cs.teamsMutex.RLock()
	// defer cs.teamsMutex.RUnlock()
	return cs.teams[cs.GetAgentMap()[agentID].GetTeamID()]
}

/*
* Update each agent's Common Pool value. For each team, check the value of its
* pool, and update that value in each of the agents part of that team.
 */
func (cs *EnvironmentServer) UpdateCommonPools() {

	// acquire mutex
	cs.teamsMutex.Lock()
	defer cs.teamsMutex.Unlock()

	agent_map := cs.GetAgentMap()
	for _, team := range cs.teams {
		// Get the value of the common pool
		pool := team.CommonPool
		// Distribute it amongst all the agents
		for _, agentID := range team.Agents {
			agent_map[agentID].SetCommonPoolValue(pool)
		}
	}
}
