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

	DialogTypeRegistration  = "registration"
	DialogTypeChooseContest = "choose_contest"

	RegistrationStepZero      = "zero"
	RegistrationStepName      = "name"
	RegistrationStepSchool    = "school"
	RegistrationStepContacts  = "contacts"
	RegistrationStepLanguages = "languages"

	ChooseContestStepZero   = "zero"
	ChooseContestStepChoice = "choice"
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

type DialogAction func(bot *Bot, update *tgbotapi.Update, state *storage.DialogState) (bool, error)
type DialogSteps map[string]DialogAction

var dialogs map[string]DialogSteps

func init() {
	dialogs = map[string]DialogSteps{
		DialogTypeRegistration:  registrationSteps,
		DialogTypeChooseContest: chooseContestSteps,
	}
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
			if err := bot.processUpdate(&update); err != nil {
				log.Errorf("update error: %s", err)
			}
		}
	}()
}

// SendNotifications Send notifications to all participants of the contest
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
	messageBuilder.WriteString("*Оповещение участников контеста \"" + esc(contest.Name) + "\"*:\n\n")
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

func (bot *Bot) processUpdate(update *tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	participantChatId := update.Message.Chat.ID

	dialogState, err := storage.GetDialogState(participantChatId)
	if err != nil {
		return err
	}
	if dialogState != nil {
		return bot.processDialog(update, dialogState)
	} else {
		return bot.processCommand(update)
	}
}

func (bot *Bot) processDialog(update *tgbotapi.Update, dialogState *storage.DialogState) error {
	dialogSteps, ok := dialogs[dialogState.DialogType]
	if !ok {
		log.Errorf("found unknown dialog type: %s", dialogState.DialogType)
		if err := storage.DeleteDialogState(dialogState.ParticipantId); err != nil {
			log.Errorf("unable to delete dialog state: %d: %s", dialogState.ParticipantId, err)
			return bot.msg(update, esc("Произошла ошибка :( Попробуйте еще раз"))
		}
	}

	dialogAction, ok := dialogSteps[dialogState.DialogStep]
	if !ok {
		log.Errorf("found unknown dialog step: %s.%s", dialogState.DialogType, dialogState.DialogStep)
		if err := storage.DeleteDialogState(dialogState.ParticipantId); err != nil {
			log.Errorf("unable to delete dialog state: %d: %s", dialogState.ParticipantId, err)
			return bot.msg(update, esc("Произошла ошибка :( Попробуйте еще раз"))
		}
	}

	if update.Message.Text == "/cancel" {
		if err := storage.DeleteDialogState(dialogState.ParticipantId); err != nil {
			log.Errorf("unable to delete dialog state %d: %s", dialogState.ParticipantId, err)
			return bot.msg(update, esc("Произошла ошибка :("))
		} else {
			return bot.msg(update, esc("Отменено"))
		}
	}

	done, err := dialogAction(bot, update, dialogState)
	if err != nil {
		log.Errorf("dialog step error: %s", err)
	}

	if done {
		if err := storage.DeleteDialogState(dialogState.ParticipantId); err != nil {
			log.Errorf("unable to delete dialog state %d: %s", dialogState.ParticipantId, err)
			return bot.msg(update, esc("Произошла ошибка :("))
		}
	} else {
		if err := storage.SaveDialogState(dialogState); err != nil {
			log.Errorf("unable to sage dialog state %d: %s", dialogState.ParticipantId, err)
			return bot.msg(update, esc("Не удалось сохранить данные :(\nПопробуйте еще раз"))
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// Utility methods

// msg Set plain text message to update's channel
func (bot *Bot) msg(update *tgbotapi.Update, message string) error {
	response := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	response.ParseMode = tgbotapi.ModeMarkdownV2
	_, err := bot.api.Send(response)
	return err
}

func esc(text string) string {
	return tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, text)
}

func trim(text string, maxLength int) string {
	trimmed := strings.TrimSpace(text)
	runes := []rune(trimmed)
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
