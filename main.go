package main

import (
	"fmt"
	"github.com/modmuss50/FactorioWrapper/utils"
	"github.com/modmuss50/FactorioWrapper/config"
	"os"
	"os/exec"
	"bufio"
	"io"
	"log"
	"strings"
	"syscall"
)

func main() {
	fmt.Println("Starting wrapper")

	dataDir := "./data/"
	if !utils.FileExists(dataDir){
		utils.MakeDir(dataDir)
	}

	fmt.Println("Loading config")
	config.LoadConfig()


	version := config.FactorioVersion
	tarBal := fmt.Sprintf("%vfactorio_headless_x64_%v.tar.xz", dataDir, version)
	gameDir := dataDir
	proccessDir := "/data/factorio"


	if !utils.FileExists(dataDir) || ! utils.FileExists(tarBal) {
		utils.MakeDir(dataDir)
		if !utils.DownloadURL(fmt.Sprintf("https://www.factorio.com/get-download/%v/headless/linux64", version), tarBal) {
			fmt.Println("Failed to download")
			os.Exit(1)
		}
		utils.ExtractTarXZ(tarBal, gameDir)
		fmt.Println("Done.")
	}

	fmt.Println("Starting game...")
	factorioProcess := getExec(proccessDir)
	fmt.Println("Getting input for game")
	factorioInput, err := factorioProcess.StdinPipe()
	utils.TextInput = factorioInput
	if err != nil {
		log.Fatal(err)
	}
	defer factorioInput.Close()
	factorioOutput, _ := factorioProcess.StdoutPipe()


	scanner := bufio.NewScanner(factorioOutput)
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			if strings.Contains(text, "ping") {
				io.WriteString(factorioInput, "Pong!\n")
			}
			if strings.Contains(text, "changing state from(CreatingGame) to(InGame)") {
				utils.LoadDiscord(config.DiscordToken)
				utils.SendStringToDiscord("Server started on factorio version " + version, config.DiscordChannel)
			}
			if strings.Contains(text, "[JOIN]") {
				utils.SendStringToDiscord(text[26:], config.DiscordChannel)
			}
			if strings.Contains(text, "[CHAT]") {
				if !strings.Contains(text, " [CHAT] <server>:") {
					utils.SendStringToDiscord(text[26:], config.DiscordChannel)
				}
			}
			if strings.Contains(text, "[LEAVE]") {
				utils.SendStringToDiscord(text[27:], config.DiscordChannel)
			}
			if strings.Contains(text, "Goodbye") {
				utils.SendStringToDiscord("Server closed", config.DiscordChannel)
				utils.DiscordClient.Close()
				os.Exit(0)
			}
			fmt.Printf("\t > %s\n", text)
		}
	}()

	fmt.Println("Launching process")
	er := factorioProcess.Run()
	if er != nil {
		log.Fatal(er)
		os.Exit(1)
	}

	//ticker := time.NewTicker(time.Second * 10)
	//go func() {
	//	for range ticker.C {
	//		//io.WriteString(factorioInput, "hello is this working?\n")
	//	}
	//}()

	readInput(factorioProcess)

}

func readInput(cmd *exec.Cmd) {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	if strings.HasPrefix(text, "stop") {
		fmt.Println("Stopping server...")
		syscall.Kill(cmd.Process.Pid, syscall.SIGINT)

	}
	if strings.HasPrefix(text, "fstop") {
		fmt.Println("Stopping server")
		cmd.Process.Kill()
		return
	}
	fmt.Println("Command not found!")
	readInput(cmd)
}

func getExec(dir string) *exec.Cmd {
	fullDir := "." + dir + "/bin/x64/factorio"
	fmt.Println(fullDir)
	factorioExec := exec.Command(fullDir, "--start-server", "." + dir + "/saves/" + config.FactorioSaveFileName)
	return factorioExec
}
