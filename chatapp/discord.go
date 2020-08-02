package chatapp

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	session      *discordgo.Session
	problemEmoji string
}

type DiscordMessageActions struct {
	session *discordgo.Session
	message *discordgo.Message
	discord *Discord
}

func (a *DiscordMessageActions) Remove() error {
	error := a.session.ChannelMessageDelete(a.message.ChannelID, a.message.ID)
	return error
}
func (a *DiscordMessageActions) RespondWithProduct(p *Product) (string, error) {
	embed := a.discord.toEmbed(p)
	m, error := a.session.ChannelMessageSendEmbed(a.message.ChannelID, embed)
	a.session.MessageReactionAdd(m.ChannelID, m.ID, a.discord.problemEmoji)
	if error != nil {
		return "", error
	}
	return discordMessageToID(m.ChannelID, m.ID), error
}

func NewDiscordSession(s *discordgo.Session) *Discord {
	return &Discord{
		session:      s,
		problemEmoji: "🇫", //For those on dark themed editors, it's that blue [F] emoji.
	}
}

func discordMessageToID(channelID string, messageID string) string {
	return fmt.Sprintf("%s-%s", channelID, messageID)
}

func (db *Discord) createMessageFromDiscordMessage(s *discordgo.Session, m *discordgo.Message) *Message {
	return &Message{
		MessageIsFromThisBot: m.Author.ID == s.State.User.ID,
		Content:              m.Content,
		ID:                   discordMessageToID(m.ChannelID, m.ID),
		Actions: &DiscordMessageActions{
			session: db.session,
			discord: db,
			message: m,
		},
	}
}

func (db *Discord) isProblemReaction(m *discordgo.MessageReaction) bool {
	return m.Emoji.Name == db.problemEmoji
}

func (db *Discord) OnProductProblemReport(cb OnProductProblemReportCallback) error {
	db.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		if m.UserID == s.State.User.ID {
			return
		}
		if db.isProblemReaction(m.MessageReaction) {
			cb(db, discordMessageToID(m.ChannelID, m.MessageID))
		}
	})
	return nil
}

func (db *Discord) OnMessage(cb OnMessageCallback) error {
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

func (db *Discord) toEmbed(product *Product) *discordgo.MessageEmbed {
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
