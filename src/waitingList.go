package main

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Waiter is a person who is on the waiting list for the next game
// (they used the "/next" Discord command)
type Waiter struct {
	DiscordMention  string
	DatetimeExpired time.Time
}

func waitingListAlert(g *Game, creator string) {
	if len(waitingList) == 0 {
		return
	}

	// Build a list of everyone on the waiting list
	mentionList := ""
	for _, waiter := range waitingList {
		if waiter.DatetimeExpired.After(time.Now()) {
			mentionList += waiter.DiscordMention + ", "
		}
	}
	mentionList = strings.TrimSuffix(mentionList, ", ")

	// Empty the waiting list
	waitingList = make([]*Waiter, 0)

	// Alert all of the people on the waiting list
	alert := creator + " created a table. (" + variants[g.Options.Variant] + ")\n" + mentionList
	discordSend(discordListenChannels[0], "", alert) // Assume that the first channel listed in the "discordListenChannels" slice is the main channel
}

func waitingListAdd(m *discordgo.MessageCreate) {
	// Get the Discord guild object
	var guild *discordgo.Guild
	if v, err := discord.Guild(discordListenChannels[0]); err != nil { // Assume that the first channel ID is the same as the server ID
		log.Error("Failed to get the Discord guild.")
	} else {
		guild = v
	}

	// Get their custom nickname for the Discord server, if any
	var username string
	for _, member := range guild.Members {
		if member.User.ID != m.Author.ID {
			continue
		}

		if member.Nick == "" {
			username = member.User.Username
		} else {
			username = member.Nick
		}
	}

	// Search through the waiting list to see if they are already on it
	for _, waiter := range waitingList {
		if waiter.DiscordMention == m.Author.Mention() {
			// Update their expiry time
			waiter.DatetimeExpired = time.Now().Add(idleWaitingListTimeout)

			// Let them know
			msg := username + ", you are already on the waiting list."
			discordSend(m.ChannelID, "", msg)
			return
		}
	}

	msg := username + ", I will ping you when the next table opens."
	discordSend(m.ChannelID, "", msg)
	waiter := &Waiter{
		DiscordMention:  m.Author.Mention(),
		DatetimeExpired: time.Now().Add(idleWaitingListTimeout),
	}
	waitingList = append(waitingList, waiter)
}