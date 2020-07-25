package main

import (
	"fmt"
	"wishlist-bot/scrapers/amazonscraper"

	"github.com/bwmarrin/discordgo"
)

func cutoffString(input string, max int, replacement string) string {
	if len(input) > max {
		return input[:max] + replacement
	}
	return input
}

func ToEmbed(product *amazonscraper.Product) *discordgo.MessageEmbed {
	const maxContentLength = 150
	const replacementContent = "..."

	priceText := fmt.Sprintf("%.2f", product.Price)
	if product.OriginalPrice > 0 {
		savings := product.OriginalPrice - product.Price
		percentOff := savings / product.OriginalPrice * 100
		priceText = fmt.Sprintf("~~%0.2f~~\n**%0.2f**\n*%.2f (%.0f%%) off*", product.OriginalPrice, product.Price, savings, percentOff)
	}
	embed := discordgo.MessageEmbed{
		Title:       product.Title,
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
	return &embed
}
