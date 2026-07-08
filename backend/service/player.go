package service

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"minecraft-manager/model"
	"minecraft-manager/pkg/rcon"
)

type PlayerService struct {
	RCON *rcon.RCONClient
	DB   *gorm.DB
}

func NewPlayerService(rconClient *rcon.RCONClient, db *gorm.DB) *PlayerService {
	return &PlayerService{RCON: rconClient, DB: db}
}

type PlayerInfo struct {
	Name   string `json:"name"`
	UUID   string `json:"uuid"`
	Online bool   `json:"online"`
}

// GetOnlinePlayers returns the list of currently online players via RCON.
func (s *PlayerService) GetOnlinePlayers() ([]PlayerInfo, error) {
	result, err := s.RCON.ExecuteWithRetry("list", 2)
	if err != nil {
		return nil, fmt.Errorf("rcon list failed: %w", err)
	}

	return parsePlayerList(result), nil
}

// KickPlayer kicks a player from the server.
func (s *PlayerService) KickPlayer(playerName, reason string) (string, error) {
	cmd := fmt.Sprintf("kick %s", playerName)
	if reason != "" {
		cmd += " " + reason
	}
	return s.RCON.ExecuteWithRetry(cmd, 2)
}

// BanPlayer bans a player and records it.
func (s *PlayerService) BanPlayer(playerName, reason string) (string, error) {
	cmd := fmt.Sprintf("ban %s", playerName)
	if reason != "" {
		cmd += " " + reason
	}
	result, err := s.RCON.ExecuteWithRetry(cmd, 2)
	if err != nil {
		return "", err
	}

	// Record the ban
	ban := model.BanRecord{
		UUID:       playerName, // RCON ban uses player name; UUID lookup via Mojang API is an extension
		PlayerName: playerName,
		Reason:     reason,
	}
	s.DB.Create(&ban)

	return result, nil
}

// OpPlayer grants operator status.
func (s *PlayerService) OpPlayer(playerName string) (string, error) {
	cmd := fmt.Sprintf("op %s", playerName)
	return s.RCON.ExecuteWithRetry(cmd, 2)
}

// DeopPlayer revokes operator status.
func (s *PlayerService) DeopPlayer(playerName string) (string, error) {
	cmd := fmt.Sprintf("deop %s", playerName)
	return s.RCON.ExecuteWithRetry(cmd, 2)
}

// parsePlayerList parses the Minecraft "list" command output.
// Format: "There are X of a max of Y players online: player1, player2, player3"
// Or: "There are 0 of a max of 20 players online:"
func parsePlayerList(output string) []PlayerInfo {
	var players []PlayerInfo

	// Find the colon separator
	idx := strings.LastIndex(output, ":")
	if idx == -1 {
		return players
	}

	namesPart := strings.TrimSpace(output[idx+1:])
	if namesPart == "" {
		return players
	}

	// Split player names
	re := regexp.MustCompile(`,\s*`)
	names := re.Split(namesPart, -1)

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			players = append(players, PlayerInfo{
				Name:   name,
				UUID:   "", // UUID lookup via Mojang API is an extension point
				Online: true,
			})
		}
	}

	return players
}
