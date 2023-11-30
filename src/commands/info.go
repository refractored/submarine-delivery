package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"runtime"
	"sandwich-delivery/src/database"
	"sandwich-delivery/src/models"
	"strconv"
)

type InfoCommand struct{}

func (c InfoCommand) getName() string {
	return "info"
}

func (c InfoCommand) getCommandData() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{Name: c.getName(),
		Description: "Manage data of an user.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "user",
				Description: "Lookup an user.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "The user to reset the daily timer.",
						Required:    true,
					},
				},
				Type: discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "bot",
				Description: "Lookup Bot Data",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	}
}

func (c InfoCommand) registerGuild() string {
	return ""
}

func (c InfoCommand) permissionLevel() models.UserPermissionLevel {
	return models.PermissionLevelAdmin
}

func (c InfoCommand) execute(session *discordgo.Session, event *discordgo.InteractionCreate) {
	var user models.User

	options := event.ApplicationCommandData().Options

	switch options[0].Name {
	case "bot":
		var pendingOrderCount int64
		var completedOrderCount int64
		var canceledOrderCount int64
		var userCount int64

		totalUsers := 0

		for _, guild := range session.State.Guilds {
			guild, err := session.Guild(guild.ID)
			if err != nil {
				fmt.Println("Error fetching guild information:", err)
				continue
			}
			totalUsers += guild.MemberCount
		}
		result := database.GetDB().Model(&models.Order{}).Where("status < ?", models.StatusDelivered).Count(&pendingOrderCount)
		if result.Error != nil {
			log.Println("Error counting orders:", result.Error)
		}
		result = database.GetDB().Model(&models.Order{}).Where("status = ?", models.StatusDelivered).Count(&completedOrderCount)
		if result.Error != nil {
			log.Println("Error counting orders:", result.Error)
		}
		result = database.GetDB().Model(&models.Order{}).Where("status > ?", models.StatusDelivered).Count(&canceledOrderCount)
		if result.Error != nil {
			log.Println("Error counting orders:", result.Error)
		}
		result = database.GetDB().Model(&models.User{}).Count(&userCount)
		if result.Error != nil {
			log.Println("Error counting orders:", result.Error)
		}
		session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title: "Bot Information",
						Description: "Bot Name: " + session.State.User.Username + "#" + session.State.User.Discriminator + "\n" +
							"Guilds: " + strconv.Itoa(len(session.State.Guilds)) + "\n" +
							"User Guild Count: " + strconv.Itoa(totalUsers) + "\n" +
							"Pending Orders: " + strconv.Itoa(int(pendingOrderCount)) + "\n" +
							"Completed Orders: " + strconv.Itoa(int(completedOrderCount)) + "\n" +
							"Canceled Orders: " + strconv.Itoa(int(canceledOrderCount)) + "\n" +
							"Sandwich Accounts: " + strconv.Itoa(int(userCount)) + "\n" +
							fmt.Sprintf("Library: DiscordGo (%s)", discordgo.VERSION) + "\n" +
							"Runtime: " + runtime.Version() + " " + runtime.GOARCH + "\n",

						Color: 0x00ff00,
						Footer: &discordgo.MessageEmbedFooter{
							Text:    "Executed by " + DisplayName(event),
							IconURL: GetUser(event).AvatarURL("256"),
						},
						Author: &discordgo.MessageEmbedAuthor{
							Name:    "Sandwich Delivery",
							IconURL: session.State.User.AvatarURL("256"),
						},
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: session.State.User.AvatarURL("256"),
						},
					},
				},
			},
		})
		break
	case "user":
		if event.ApplicationCommandData().Options[0].Options[0].UserValue(session).ID == session.State.User.ID {
			session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You can lookup the bots information with /info bot",
				},
			})
			return
		}
		resp := database.GetDB().First(&user, "user_id = ?", event.ApplicationCommandData().Options[0].Options[0].UserValue(session).ID)
		if resp.RowsAffected == 0 {
			user.UserID = event.ApplicationCommandData().Options[0].Options[0].UserValue(session).ID

			database.GetDB().Save(&user)
		}
		userarg, _ := session.User(user.UserID)
		var dailyclaimed string
		if user.DailyClaimedAt != nil {
			dailyclaimed = user.DailyClaimedAt.String()
		} else {
			dailyclaimed = "Never"
		}

		session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title: "User Information",
						Description: "Orders Created: " + strconv.Itoa(int(user.OrdersCreated)) + "\n" +
							"Orders Accepted: " + strconv.Itoa(int(user.OrdersAccepted)) + "\n" +
							"Credits: " + strconv.Itoa(int(user.Credits)) + "\n" +
							"DB ID: " + strconv.Itoa(int(user.ID)) + "\n" +
							"Blacklisted: " + strconv.FormatBool(user.IsBlacklisted) + "\n" +
							"Permission Level: " + strconv.Itoa(int(user.OrdersAccepted)) + "\n" +
							"Daily Claimed At: " + dailyclaimed + "\n" +
							"Sandwich Account Creation: " + user.CreatedAt.String() + "\n",
						Color: 0x00ff00,
						Footer: &discordgo.MessageEmbedFooter{
							Text:    "Executed by " + DisplayName(event),
							IconURL: GetUser(event).AvatarURL("256"),
						},
						Author: &discordgo.MessageEmbedAuthor{
							Name:    "Sandwich Delivery",
							IconURL: session.State.User.AvatarURL("256"),
						},
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: userarg.AvatarURL("256"),
						},
					},
				},
			},
		})
		break
	}
}
