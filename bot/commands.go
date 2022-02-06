package bot

import (
	"contest-registration-bot/storage"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func (bot *Bot) processCommand(update *tgbotapi.Update) error {
	if !update.Message.IsCommand() {
		return bot.sendMessageToUpdate(update, esc("Пожалуйста, введите команду. Для справки введите /help"))
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
	case "register":
		return bot.commandRegister(update)
	default:
		return bot.sendMessageToUpdate(update, esc("Не знаю такой команды :("))
	}
}

// commandStart First run message
func (bot *Bot) commandStart(update *tgbotapi.Update) error {
	return bot.sendMessageToUpdate(update, esc("Этот бот поможет зарегистрироваться на олимпиаду. Для справки введите /help"))
}

// commandHelp Help message
func (bot *Bot) commandHelp(update *tgbotapi.Update) error {
	message := strings.Builder{}
	message.WriteString(esc("Этот бот поможет зарегистрироваться на олимпиаду. Доступные команды:\n"))
	message.WriteString(esc("/help - справка\n"))
	message.WriteString(esc("/contests - список контестов и сведения о регистрации\n"))
	message.WriteString(esc("/registration - регистрация на контест\n"))
	return bot.sendMessageToUpdate(update, message.String())
}

// commandContests List all current contests
func (bot *Bot) commandContests(update *tgbotapi.Update) error {
	contests, err := storage.GetContests()
	if err != nil {
		log.Errorf("/contests: unable to get contests: %s", err)
		return bot.sendMessageToUpdate(update, esc("Не удалось найти контесты :("))
	}

	participation, err := storage.GetContestParticipantParticipation(update.Message.Chat.ID)
	if err != nil {
		log.Errorf("/contests: unable to get participation: %s", err)
		return bot.sendMessageToUpdate(update, esc("Не удалось найти регистрации на контесты :("))
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
		}
	}

	if !contestsFound {
		return bot.sendMessageToUpdate(update, "Сейчас контестов нет")
	}

	return bot.sendMessageToUpdate(update, message.String())
}

// commandRegistration List contests available for registration
func (bot *Bot) commandRegistration(update *tgbotapi.Update) error {
	contests, err := storage.GetContests()
	if err != nil {
		log.Errorf("/registration: unable to get contests: %s", err)
		return bot.sendMessageToUpdate(update, esc("Не удалось найти контесты :("))
	}

	var contestButtons []tgbotapi.KeyboardButton

	for _, contest := range contests {
		if contest.Hidden || contest.Closed {
			continue
		}
		callback := fmt.Sprintf("/register %d %s", contest.Id, contest.Name)
		button := tgbotapi.NewKeyboardButton(callback)
		contestButtons = append(contestButtons, button)
	}

	if len(contestButtons) > 0 {
		message := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите доступный для регистрации контест")
		keyboard := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(contestButtons...))
		keyboard.OneTimeKeyboard = true
		message.ReplyMarkup = keyboard
		_, err := bot.api.Send(message)
		return err
	} else {
		return bot.sendMessageToUpdate(update, "Доступных для регистрации контестов нет")
	}
}

// commandRegister Begin contest registration
func (bot *Bot) commandRegister(update *tgbotapi.Update) error {
	commandArguments := strings.TrimSpace(update.Message.CommandArguments())
	arguments := strings.Split(commandArguments, " ")
	if len(arguments) == 0 {
		log.Errorf("/register: format mismatch")
		return bot.sendMessageToUpdate(update, esc("Не удалось распознать команду :("))
	}

	contestId, err := strconv.ParseUint(arguments[0], 10, 64)
	if err != nil {
		log.Errorf("/register contestId parsing error: %s", err)
		return bot.sendMessageToUpdate(update, esc("Не удалось определить, на какой контест идет регистрация :("))
	}

	contest, err := storage.GetContest(contestId)
	if err != nil {
		log.Errorf("/register: contest get error: %s", err)
		return bot.sendMessageToUpdate(update, esc("Не удалось найти контест :("))
	}
	if contest.Hidden {
		log.Errorf("/register: request of hidden contest %d", contestId)
		return bot.sendMessageToUpdate(update, esc("Этот контест больше не существует :("))
	}
	if contest.Closed {
		log.Errorf("/register: request of closed contest %d", contestId)
		return bot.sendMessageToUpdate(update, esc("Регистрация на этот контест закрыта :("))
	}

	participantId := update.Message.Chat.ID
	participation, err := storage.GetContestParticipantParticipation(participantId)
	if err != nil {
		log.Errorf("/register: unable to find participant's contests: %s", err)
		return bot.sendMessageToUpdate(update, esc("Что-то пошло не так :("))
	}
	for _, participant := range participation {
		if participant.ContestId == contestId {
			log.Infof("/register: double registration of %d to %d", participantId, contestId)
			return bot.sendMessageToUpdate(update, "На этот контест уже есть регистрация")
		}
	}

	registrationState := &storage.RegistrationState{
		ParticipantId: participantId,
		ContestId:     contestId,
		Step:          storage.RegistrationStepZero,
	}
	err = storage.SaveRegistrationState(registrationState)
	if err != nil {
		log.Errorf("/register: unable to save new registration state: %s", err)
		return bot.sendMessageToUpdate(update, esc("Не удалось начать регистрацию :("))
	}

	return bot.processRegistration(update, registrationState)
}
