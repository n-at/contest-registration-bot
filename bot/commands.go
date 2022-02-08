package bot

import (
	"contest-registration-bot/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"strings"
)

func (bot *Bot) processCommand(update *tgbotapi.Update) error {
	if !update.Message.IsCommand() {
		return bot.msg(update, esc("Пожалуйста, введите команду. Для справки введите /help"))
	}

	switch update.Message.Command() {
	case "start":
		return bot.commandStart(update)
	case "help":
		return bot.commandHelp(update)
	case "contests":
		return bot.commandContests(update)
	case "registration":
		return bot.commandRegistration(update)
	default:
		return bot.msg(update, esc("Не знаю такой команды :("))
	}
}

// commandStart First run message
func (bot *Bot) commandStart(update *tgbotapi.Update) error {
	return bot.msg(update, esc("Этот бот поможет зарегистрироваться на олимпиаду. Для справки введите /help"))
}

// commandHelp Help message
func (bot *Bot) commandHelp(update *tgbotapi.Update) error {
	message := strings.Builder{}
	message.WriteString(esc("Этот бот поможет зарегистрироваться на контест. Доступные команды:\n"))
	message.WriteString(esc("/help - справка\n"))
	message.WriteString(esc("/contests - список контестов и сведения о регистрации\n"))
	message.WriteString(esc("/registration - регистрация на контест\n"))
	return bot.msg(update, message.String())
}

// commandContests List all current contests
func (bot *Bot) commandContests(update *tgbotapi.Update) error {
	contests, err := storage.GetContests()
	if err != nil {
		log.Errorf("/contests: unable to get contests: %s", err)
		return bot.msg(update, esc("Не удалось найти контесты :("))
	}

	participation, err := storage.GetContestParticipantParticipation(update.Message.Chat.ID)
	if err != nil {
		log.Errorf("/contests: unable to get participation: %s", err)
		return bot.msg(update, esc("Не удалось найти регистрации на контесты :("))
	}
	participants := make(map[uint64]storage.ContestParticipant)
	for _, participant := range participation {
		participants[participant.ContestId] = participant
	}

	message := strings.Builder{}
	message.WriteString("Найдены контесты:\n")

	contestsFound := false

	for _, contest := range contests {
		if contest.Hidden {
			continue
		}

		contestsFound = true

		message.WriteRune('\n')
		message.WriteString("*" + esc(contest.Name) + "*\n")
		message.WriteString("*Что:* " + esc(contest.Description) + "\n")
		message.WriteString("*Где:* " + esc(contest.Where) + "\n")
		message.WriteString("*Когда:* " + esc(contest.When) + "\n")
		if contest.Closed {
			message.WriteString("_Регистрация на контест закрыта_\n")
		}

		participant, ok := participants[contest.Id]
		if ok {
			message.WriteString("_Есть регистрация на контест_\n")
			message.WriteString("*Имя:* " + esc(participant.Name) + "\n")
			message.WriteString("*Школа/ВУЗ:* " + esc(participant.School) + "\n")
			message.WriteString("*Логин:* `" + esc(participant.Login) + "`\n")
			message.WriteString("*Пароль:* `" + esc(participant.Password) + "`\n")

			notifications, err := storage.GetContestNotifications(contest.Id)
			if err != nil {
				log.Errorf("/contest: unable to get notifications of contest %d: %s", contest.Id, err)
			} else if len(notifications) > 0 {
				message.WriteString("_Оповещения участников:_\n")
				for _, notification := range notifications {
					message.WriteString(esc(">>> "+notification.Message) + "\n")
				}
			}
		}
	}

	if !contestsFound {
		return bot.msg(update, "Сейчас контестов нет")
	}

	return bot.msg(update, message.String())
}

// commandRegistration Start contest choose dialog
func (bot *Bot) commandRegistration(update *tgbotapi.Update) error {
	state := &storage.DialogState{
		ParticipantId: update.Message.Chat.ID,
		DialogType:    DialogTypeChooseContest,
		DialogStep:    ChooseContestStepZero,
	}
	return bot.processDialog(update, state)
}
