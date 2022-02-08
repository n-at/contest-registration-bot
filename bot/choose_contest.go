package bot

import (
	"contest-registration-bot/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

var chooseContestSteps = map[string]DialogAction{
	ChooseContestStepZero: func(bot *Bot, update *tgbotapi.Update, state *storage.DialogState) (bool, error) {
		contests, err := storage.GetContests()
		if err != nil {
			log.Errorf("choose contest: unable to get contests: %s", err)
			return true, bot.msg(update, esc("Не удалось найти контесты :("))
		}

		var contestButtons []tgbotapi.KeyboardButton

		for _, contest := range contests {
			if contest.Hidden || contest.Closed {
				continue
			}
			button := tgbotapi.NewKeyboardButton(contest.Name)
			contestButtons = append(contestButtons, button)
		}

		if len(contestButtons) == 0 {
			return true, bot.msg(update, "Доступных для регистрации контестов нет")
		}

		message := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите доступный для регистрации контест.\nНажмите на кнопку с названием контеста")
		keyboard := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(contestButtons...))
		keyboard.OneTimeKeyboard = true
		message.ReplyMarkup = keyboard
		_, err = bot.api.Send(message)

		state.DialogStep = ChooseContestStepChoice

		return false, err
	},

	ChooseContestStepChoice: func(bot *Bot, update *tgbotapi.Update, state *storage.DialogState) (bool, error) {
		message := tgbotapi.NewMessage(update.Message.Chat.ID, "...")
		message.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
		_, err := bot.api.Send(message)
		if err != nil {
			log.Errorf("choose contest: unable to remove keyboard: %s", err)
			return true, bot.msg(update, esc("Что-то пошло не так :("))
		}

		contest, err := storage.GetContestByName(update.Message.Text)
		if err != nil {
			log.Errorf("choose contest: %s", err)
			return true, bot.msg(update, esc("Не удалось найти контест с указанным именем :("))
		}
		if contest == nil {
			log.Errorf("choose contest: %s", err)
			return true, bot.msg(update, esc("Контест не найден :("))
		}
		if contest.Hidden {
			log.Errorf("choose contest: request of hidden contest %d", contest.Id)
			return true, bot.msg(update, esc("Этот контест больше не существует :("))
		}
		if contest.Closed {
			log.Errorf("choose contest: request of closed contest %d", contest.Id)
			return true, bot.msg(update, esc("Регистрация на этот контест закрыта :("))
		}

		participantId := update.Message.Chat.ID
		participation, err := storage.GetContestParticipantParticipation(participantId)
		if err != nil {
			log.Errorf("choose contest: unable to find participant's contests: %s", err)
			return true, bot.msg(update, esc("Что-то пошло не так :("))
		}
		for _, participant := range participation {
			if participant.ContestId == contest.Id {
				log.Infof("choose contest: double registration of %d to %d", participantId, contest.Id)
				return true, bot.msg(update, "На этот контест уже есть регистрация")
			}
		}

		state.DialogType = DialogTypeRegistration
		state.DialogStep = RegistrationStepZero
		state.Values = storage.DialogValues{
			"ContestId": contest.Id,
		}

		return false, bot.processDialog(update, state)
	},
}
