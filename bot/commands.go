package bot

import (
	"contest-registration-bot/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"strings"
)

// commandStart First run message
func (bot *Bot) commandStart(update *tgbotapi.Update) error {
	return bot.sendMessageToUpdate(update, "Этот бот поможет зарегистрироваться на олимпиаду. Для справки введите /help")
}

// commandHelp Help message
func (bot *Bot) commandHelp(update *tgbotapi.Update) error {
	message := strings.Builder{}
	message.WriteString("Этот бот поможет зарегистрироваться на олимпиаду. Доступные команды:\n")
	message.WriteString("/help - справка\n")
	message.WriteString("/contests - список контестов\n")
	return bot.sendMessageToUpdate(update, message.String())
}

// commandContests List all current contests
func (bot *Bot) commandContests(update *tgbotapi.Update) error {
	contests, err := storage.GetContests()
	if err != nil {
		log.Errorf("unable to get contests: %s", err)
		return bot.sendMessageToUpdate(update, "Не удалось найти контесты :(")
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
		message.WriteString("*" + bot.esc(contest.Name) + "*\n")
		message.WriteString("*Что:* " + bot.esc(contest.Description) + "\n")
		message.WriteString("*Где:* " + bot.esc(contest.Where) + "\n")
		message.WriteString("*Когда:* " + bot.esc(contest.When) + "\n")
		if contest.Closed {
			message.WriteString("*Регистрация закрыта*\n")
		}
	}

	if !contestsFound {
		return bot.sendMessageToUpdate(update, "Сейчас контестов нет")
	}

	return bot.sendMessageToUpdate(update, message.String())
}
