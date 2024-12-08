package main

import (
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	baseServer "github.com/MattSScott/basePlatformSOMAS/v2/pkg/server"

	agents "github.com/ADimoska/SOMASExtended/agents"
	common "github.com/ADimoska/SOMASExtended/common"
	gameRecorder "github.com/ADimoska/SOMASExtended/gameRecorder"
	envServer "github.com/ADimoska/SOMASExtended/server"
)

func main() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Create log file with timestamp in name
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile, err := os.OpenFile("logs/log_"+timestamp+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Create a MultiWriter to write to both the log file and stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Set log output to multiWriter
	log.SetOutput(multiWriter)

	// Remove date and time prefix from log entries
	log.SetFlags(0)

	log.Println("main function started.")

	// agent configuration
	agentConfig := agents.AgentConfig{
		InitScore:    0,
		VerboseLevel: 10,
	}

	team4_evil := agents.Team4Config{
		Chaoticness: 1, // from 1 to 3, 3 being most chaotic
		Evilness:    3, // from 1 to 3, 3 being most evil
	}

	team4_good := agents.Team4Config{
		Chaoticness: 1, // from 1 to 3, 3 being most chaotic
		Evilness:    1, // from 1 to 3, 3 being most evil
	}

	team4_neutral := agents.Team4Config{
		Chaoticness: 1, // from 1 to 3, 3 being most chaotic
		Evilness:    2, // from 1 to 3, 3 being most evil
	}

	serv := &envServer.EnvironmentServer{
		// note: the zero turn is used for team forming
		BaseServer: baseServer.CreateBaseServer[common.IExtendedAgent](
			3,                   //  iterations
			100,                 //  turns per iteration
			50*time.Millisecond, //  max duration
			10),                 //  message bandwidth
		Teams: make(map[uuid.UUID]*common.Team),
	}
	serv.Init(
		3, // turns to apply threshold once
	)
	serv.SetGameRunner(serv)

	const numAgents int = 20

	// store agent population for all experiments
	agentPopulations := [][]common.IExtendedAgent{}

	// experiment 0, baseline lawful neutral
	agentPopulation := []common.IExtendedAgent{}
	for i := 0; i < 20; i++ {
		agentPopulation = append(agentPopulation, agents.Team4_CreateAgent(serv, agentConfig, team4_neutral))
	}
	agentPopulations = append(agentPopulations, agentPopulation)

	// experiment 1-12, evil vs good ()
	// prop increases in 10% increments
	for evilProp := 0; evilProp < 11; evilProp++ {
		numAgentsEvil := int(numAgents * evilProp / 10)

		for i := 0; i < numAgentsEvil; i++ {
			agentPopulation = append(agentPopulation, agents.Team4_CreateAgent(serv, agentConfig, team4_evil))
		}
		for i := 0; i < numAgents-numAgentsEvil; i++ {
			agentPopulation = append(agentPopulation, agents.Team4_CreateAgent(serv, agentConfig, team4_good))
		}
		agentPopulations = append(agentPopulations, agentPopulation)
	}

	// for i := 0; i < numAgents-2; i++ {
	// 	// Add mostly honest agents
	// 	agentPopulation = append(agentPopulation, agents.Create_Team1Agent(serv, agentConfig, agents.Honest))
	// }

	// // Add a short term and long term cheater agent from team 1
	// team1ShortTermCheater := agents.Create_Team1Agent(serv, agentConfig, agents.CheatShortTerm)
	// log.Printf("Team1 %v is of type CheatShortTerm", team1ShortTermCheater.GetID())
	// agentPopulation = append(agentPopulation, team1ShortTermCheater)

	// team1LongTermCheater := agents.Create_Team1Agent(serv, agentConfig, agents.CheatLongTerm)
	// log.Printf("Team1 %v is of type CheatLongTerm", team1LongTermCheater.GetID())
	// agentPopulation = append(agentPopulation, team1LongTermCheater)

	for experiment := 10; experiment < len(agentPopulations); experiment++ {
		agentPopulation := agentPopulations[experiment]
		for i, agent := range agentPopulation {
			agent.SetName(i)
			serv.AddAgent(agent)
		}

		//serv.ReportMessagingDiagnostics()
		serv.Start()

		// custom function to see agent result
		serv.LogAgentStatus()
		serv.LogTeamStatus()

		// record data
		// serv.DataRecorder.GamePlaybackSummary()
		gameRecorder.ExportToCSV(serv.DataRecorder, "visualization_output/csv_data/experiment_"+strconv.Itoa(experiment))
	}
}
