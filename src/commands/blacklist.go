package commands

import (
	"github.com/bwmarrin/discordgo"
	"go-discord-bot/src/models"
	"gorm.io/gorm"
	"strings"
)

func BlacklistCommand(s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB) {
	args := strings.Split(m.Content, " ")

	if len(args) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: % blacklist <user_id>")
		return
	}

	userID := args[2]

	if IsUserBlacklisted(db, userID) {
		s.ChannelMessageSend(m.ChannelID, "User is already blacklisted.")
		return
	}

	err := db.Create(&models.BlacklistUser{UserID: userID}).Error
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error blacklisting the user.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "User blacklisted successfully.")
}
