package bot

import (
	"contest-registration-bot/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"strings"
)

var registrationSteps = map[string]DialogAction{
	RegistrationStepZero: func(bot *Bot, update *tgbotapi.Update, state *storage.DialogState) (bool, error) {
		message := strings.Builder{}
		message.WriteString(esc("Начинаем регистрацию на контест.\n"))
		message.WriteString(esc("Чтобы отменить регистрацию, в любой момент введите /cancel\n\n"))
		message.WriteString(esc("Введите Ваши фамилию, имя и отчество\n"))
		message.WriteString("_" + esc("например: Иванов Иван Иванович") + "_")
		if err := bot.msg(update, message.String()); err != nil {
			return false, err
		}
		state.DialogStep = RegistrationStepName
		return false, nil
	},

	RegistrationStepName: func(bot *Bot, update *tgbotapi.Update, state *storage.DialogState) (bool, error) {
		name := trim(update.Message.Text, 100)
		if len(name) == 0 {
			if err := bot.msg(update, esc("Попробуйте ввести ФИО еще раз")); err != nil {
				return false, err
			}
			return false, nil
		}
		message := strings.Builder{}
		message.WriteString(esc("Введите название Вашей школы или ВУЗа, а также класс (или курс и группу)\n"))
		message.WriteString("_" + esc("например: ПсковГУ, 1 курс, группа 081-0902") + "_")
		if err := bot.msg(update, message.String()); err != nil {
			return false, err
		}
		state.Values["Name"] = name
		state.DialogStep = RegistrationStepSchool
		return false, nil
	},

	RegistrationStepSchool: func(bot *Bot, update *tgbotapi.Update, state *storage.DialogState) (bool, error) {
		school := trim(update.Message.Text, 200)
		if len(school) == 0 {
			if err := bot.msg(update, esc("Попробуйте ввести название образовательной организации еще раз")); err != nil {
				return false, err
			}
			return false, nil
		}
		message := strings.Builder{}
		message.WriteString(esc("Введите Ваши контактные данные, например номер телефона и адрес электронной почты, либо напишите, как в Вами можно связаться\n"))
		message.WriteString("_" + esc("например: +7-000-000-00-00, mail@example.com") + "_")
		if err := bot.msg(update, message.String()); err != nil {
			return false, err
		}
		state.Values["School"] = school
		state.DialogStep = RegistrationStepContacts
		return false, nil
	},

	RegistrationStepContacts: func(bot *Bot, update *tgbotapi.Update, state *storage.DialogState) (bool, error) {
		contacts := trim(update.Message.Text, 100)
		if len(contacts) == 0 {
			if err := bot.msg(update, esc("Попробуйте ввести контакты еще раз")); err != nil {
				return false, err
			}
			return false, nil
		}
		message := strings.Builder{}
		message.WriteString(esc("И последний вопрос, какие предпочитаете языки и среды программирования\n"))
		message.WriteString("_" + esc("например: C++, Visual Studio") + "_")
		if err := bot.msg(update, message.String()); err != nil {
			return false, err
		}
		state.Values["Contacts"] = contacts
		state.DialogStep = RegistrationStepLanguages
		return false, nil
	},

	RegistrationStepLanguages: func(bot *Bot, update *tgbotapi.Update, state *storage.DialogState) (bool, error) {
		languages := trim(update.Message.Text, 200)
		participant := &storage.ContestParticipant{
			ParticipantId: state.ParticipantId,
			ContestId:     state.Values["ContestId"].(uint64),
			Name:          state.Values["Name"].(string),
			School:        state.Values["School"].(string),
			Contacts:      state.Values["Contacts"].(string),
			Languages:     languages,
		}
		if err := storage.SaveContestParticipant(participant); err != nil {
			log.Errorf("registration: unable to save contest participant: %s", err)
			if err := bot.msg(update, esc("Не удалось зарегистрироваться на контест. Попробуйте еще раз")); err != nil {
				return true, err
			}
			return true, nil
		}
		message := strings.Builder{}
		message.WriteString(esc("Спасибо за ответы. Регистрация завершена :)\n"))
		message.WriteString("*Логин:* `" + esc(participant.Login) + "`\n")
		message.WriteString("*Пароль:* `" + esc(participant.Password) + "`\n\n")
		message.WriteString("Посмотреть сведения о контесте и проверить регистрационные данные можно через команду /contests")
		if err := bot.msg(update, message.String()); err != nil {
			return true, err
		}
		return true, nil
	},
}
