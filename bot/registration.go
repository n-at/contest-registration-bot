package bot

import (
	"contest-registration-bot/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"strings"
)

func (bot *Bot) processRegistration(update *tgbotapi.Update, registrationState *storage.RegistrationState) error {
	switch registrationState.Step {
	case storage.RegistrationStepZero:
		registrationState.Step = storage.RegistrationStepName
		if err := storage.SaveRegistrationState(registrationState); err != nil {
			log.Errorf("registration: unable to save registration state: %s", err)
			return bot.msg(update, esc("Не удалось сохранить данные :("))
		}
		return bot.msg(update, esc("Начинаем регистрацию на контест. Введите Ваше имя:"))

	case storage.RegistrationStepName:
		name := strings.TrimSpace(update.Message.Text)
		name = trimString(name, 100)
		if len(name) == 0 {
			return bot.msg(update, esc("Попробуйте ввести имя еще раз"))
		}
		registrationState.Name = name
		registrationState.Step = storage.RegistrationStepSchool
		if err := storage.SaveRegistrationState(registrationState); err != nil {
			log.Errorf("registration: unable to save registration state: %s", err)
			return bot.msg(update, esc("Не удалось сохранить данные :(\nПопробуйте ввести имя еще раз"))
		}
		return bot.msg(update, esc("Введите название Вашей школы или ВУЗа, а также класс (или курс и группу):"))

	case storage.RegistrationStepSchool:
		school := strings.TrimSpace(update.Message.Text)
		school = trimString(school, 200)
		if len(school) == 0 {
			return bot.msg(update, esc("Попробуйте ввести название образовательной организации еще раз"))
		}
		registrationState.School = school
		registrationState.Step = storage.RegistrationStepContacts
		if err := storage.SaveRegistrationState(registrationState); err != nil {
			log.Errorf("registration: unable to save registration state: %s", err)
			return bot.msg(update, esc("Не удалось сохранить данные :(\nПопробуйте ввести название образовательной организации еще раз"))
		}
		return bot.msg(update, esc("Введите Ваши контактные данные (номер телефона и адрес электронной почты):"))

	case storage.RegistrationStepContacts:
		contacts := strings.TrimSpace(update.Message.Text)
		contacts = trimString(contacts, 100)
		if len(contacts) == 0 {
			return bot.msg(update, esc("Попробуйте ввести контакты еще раз"))
		}
		registrationState.Contacts = contacts
		registrationState.Step = storage.RegistrationStepLanguages
		if err := storage.SaveRegistrationState(registrationState); err != nil {
			log.Errorf("registration: unable to save registration state: %s", err)
			return bot.msg(update, esc("Не удалось сохранить данные :(\nПопробуйте ввести контакты еще раз"))
		}
		return bot.msg(update, esc("И последний вопрос, какие предпочитаете языки и среды программирования:"))

	case storage.RegistrationStepLanguages:
		languages := strings.TrimSpace(update.Message.Text)
		languages = trimString(languages, 200)
		participant := &storage.ContestParticipant{
			ParticipantId: registrationState.ParticipantId,
			ContestId:     registrationState.ContestId,
			Name:          registrationState.Name,
			School:        registrationState.School,
			Contacts:      registrationState.Contacts,
			Languages:     languages,
		}
		if err := storage.DeleteRegistrationState(registrationState.ParticipantId); err != nil {
			log.Errorf("registration: unable to delete registration state: %s", err)
			return bot.msg(update, esc("Не удалось сохранить данные :(\nПопробуйте ввести языки программирования еще раз"))
		}
		if err := storage.SaveContestParticipant(participant); err != nil {
			log.Errorf("registration: unable to save contest participant: %s", err)
			return bot.msg(update, esc("Не удалось зарегистрироваться на контест. Попробуйте еще раз"))
		}
		message := strings.Builder{}
		message.WriteString(esc("Регистрация завершена :)\n"))
		message.WriteString("**Логин:** `" + esc(participant.Login) + "`\n")
		message.WriteString("**Пароль:** `" + esc(participant.Password) + "`")
		return bot.msg(update, message.String())

	default:
		if err := storage.DeleteRegistrationState(registrationState.ParticipantId); err != nil {
			log.Errorf("registration: unable to delete registration state: %s", err)
		}
		return bot.msg(update, esc("Ошибка регистрации :("))
	}
}
