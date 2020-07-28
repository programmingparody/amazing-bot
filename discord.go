package main

import (
	"fmt"
	"wishlist-bot/scrapers/amazonscraper"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	session      *discordgo.Session
	problemEmoji string
}

func NewDiscordChatAppSession(s *discordgo.Session) *DiscordBot {
	return &DiscordBot{
		session:      s,
		problemEmoji: "🇫", //For those on dark themed editors, it's that blue [F] emoji.
	}
}

func discordMessageID(channelID string, messageID string) string {
	return fmt.Sprintf("%s-%s", channelID, messageID)
}

func (db *DiscordBot) createMessageFromDiscordMessage(s *discordgo.Session, m *discordgo.Message) *Message {
	return &Message{
		MessageIsFromThisBot: m.Author.ID == s.State.User.ID,
		Content:              m.Content,
		ID:                   discordMessageID(m.ChannelID, m.ID),
		Remove: func() error {
			error := s.ChannelMessageDelete(m.ChannelID, m.ID)
			return error
		},
		RespondWithAmazonProduct: func(p *amazonscraper.Product) (string, error) {
			embed := db.toEmbed(p)
			m, error := s.ChannelMessageSendEmbed(m.ChannelID, embed)
			s.MessageReactionAdd(m.ChannelID, m.ID, db.problemEmoji)
			if error != nil {
				return "", error
			}
			return discordMessageID(m.ChannelID, m.ID), error
		},
	}
}

func (db *DiscordBot) isProblemReaction(m *discordgo.MessageReaction) bool {
	return m.Emoji.Name == db.problemEmoji
}

func (db *DiscordBot) OnProductProblemReport(cb OnProductProblemReportCallback) error {
	db.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		if m.UserID == s.State.User.ID {
			return
		}
		if db.isProblemReaction(m.MessageReaction) {
			cb(db, discordMessageID(m.ChannelID, m.MessageID))
		}
	})
	return nil
}

func (db *DiscordBot) OnMessage(cb OnMessageCallback) error {
	db.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		cb(db, db.createMessageFromDiscordMessage(s, m.Message))
	})
	return nil
}

func cutoffString(input string, max int, replacement string) string {
	if len(input) > max {
		return input[:max] + replacement
	}
	return input
}

func (db *DiscordBot) toEmbed(product *amazonscraper.Product) *discordgo.MessageEmbed {
	const maxContentLength = 150
	const replacementContent = "..."

	title := product.Title
	if len(title) == 0 {
		title = "*Title not found*"
	}

	priceText := fmt.Sprintf("%.2f", product.Price)
	if product.OriginalPrice > 0 {
		savings := product.OriginalPrice - product.Price
		percentOff := savings / product.OriginalPrice * 100
		priceText = fmt.Sprintf("~~%0.2f~~\n**%0.2f**\n*%.2f (%.0f%%) off*", product.OriginalPrice, product.Price, savings, percentOff)
	}
	embed := discordgo.MessageEmbed{
		Title:       title,
		URL:         product.URL.String(),
		Description: cutoffString(product.Description, maxContentLength, replacementContent),

		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: product.ImageURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Price",
				Value:  priceText,
				Inline: true,
			},
			{
				Name:   "Rating",
				Value:  fmt.Sprintf("%.1f", product.Rating),
				Inline: true,
			},
			{
				Name:   "#Ratings",
				Value:  fmt.Sprintf("%v", product.RatingsCount),
				Inline: true,
			},
		},
		Color: 0xFF9900,
	}

	if product.OutOfStock {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Out Of Stock",
			Value:  "😢",
			Inline: true,
		})
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name: "\u200B",
		Value: fmt.Sprintf(
			`**Something wrong with this result?**
			React with %s to report this embed and pay respects`, db.problemEmoji),
		Inline: false,
	})
	return &embed
}
