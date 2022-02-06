package bot

import (
	"contest-registration-bot/storage"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	defaultTimeout = 30
)

type Configuration struct {
	Token         string
	Debug         bool
	UpdateTimeout int
}

type Bot struct {
	api    *tgbotapi.BotAPI
	config Configuration
}

///////////////////////////////////////////////////////////////////////////////

// New Create new bot
func New(config Configuration) (*Bot, error) {
	if len(config.Token) == 0 {
		return nil, errors.New("bot token required")
	}
	if config.UpdateTimeout == 0 {
		config.UpdateTimeout = defaultTimeout
	}

	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}

	api.Debug = config.Debug

	return &Bot{
		api:    api,
		config: config,
	}, nil
}

// Start Process new messages
func (bot *Bot) Start() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = bot.config.UpdateTimeout

	updates := bot.api.GetUpdatesChan(updateConfig)

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}
			if err := bot.processUpdate(&update); err != nil {
				log.Errorf("update error: %s", err)
			}
		}
	}()
}

func (bot *Bot) processUpdate(update *tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	participantChatId := update.Message.Chat.ID

	registrationState, err := storage.GetRegistrationState(participantChatId)
	if err != nil {
		return err
	}
	if registrationState != nil {
		return bot.processRegistration(update, registrationState)
	}

	return bot.processCommand(update)
}

func (bot *Bot) SendNotifications(contestId uint64, text string) error {
	contest, err := storage.GetContest(contestId)
	if err != nil {
		return err
	}

	participants, err := storage.GetContestParticipants(contestId)
	if err != nil {
		return nil
	}

	messageBuilder := strings.Builder{}
	messageBuilder.WriteString("*Оповещение для участников контеста \"" + esc(contest.Name) + "\"*:\n\n")
	messageBuilder.WriteString(esc(text))
	messageText := messageBuilder.String()

	go func() {
		for _, participant := range participants {
			if participant.ParticipantId == 0 {
				continue
			}
			message := tgbotapi.NewMessage(participant.ParticipantId, messageText)
			message.ParseMode = tgbotapi.ModeMarkdownV2
			_, err := bot.api.Send(message)
			if err != nil {
				log.Errorf("unable to send contest %d notification to %d", contestId, participant.ParticipantId)
			}
		}
	}()

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// Utility methods

// sendMessageToUpdate Set plain text message to update's channel
func (bot *Bot) sendMessageToUpdate(update *tgbotapi.Update, message string) error {
	response := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	response.ParseMode = tgbotapi.ModeMarkdownV2
	_, err := bot.api.Send(response)
	return err
}

func esc(text string) string {
	return tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, text)
}

func trimString(text string, maxLength int) string {
	runes := []rune(text)
	length := min(len(runes), maxLength)
	return string(runes[0:length])
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
