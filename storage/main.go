package storage

import (
	"errors"
	"github.com/timshannon/bolthold"
	"math/rand"
	"sort"
	"strings"
)

var (
	store *bolthold.Store
)

func Open(fileName string) error {
	var err error

	store, err = bolthold.Open(fileName, 0666, nil)
	if err != nil {
		return err
	}

	return nil
}

func Close() error {
	err := store.Close()
	store = nil
	return err
}

///////////////////////////////////////////////////////////////////////////////

// GetContests List of all contests, ordered by id
func GetContests() ([]Contest, error) {
	var contests []Contest
	if err := store.Find(&contests, nil); err != nil {
		return nil, err
	}
	return contestsSorted(contests), nil
}

// GetContest One contest by id
func GetContest(id uint64) (*Contest, error) {
	var contest Contest
	if err := store.FindOne(&contest, bolthold.Where(bolthold.Key).Eq(id)); err != nil {
		return nil, err
	}
	return &contest, nil
}

// GetContestByName Find contest by its name
func GetContestByName(name string) (*Contest, error) {
	var contest Contest
	if err := store.FindOne(&contest, bolthold.Where("Name").Eq(name)); err != nil {
		return nil, err
	}
	return &contest, nil
}

// SaveContest Create new or update contest
func SaveContest(contest *Contest) error {
	if contest.Id != 0 {
		return store.Update(contest.Id, contest)
	} else {
		return store.Insert(bolthold.NextSequence(), contest)
	}
}

func contestsSorted(contests []Contest) []Contest {
	sort.Slice(contests, func(i, j int) bool {
		return contests[i].Id < contests[j].Id
	})
	return contests
}

///////////////////////////////////////////////////////////////////////////////

// GetContestParticipants List all participants registered to given contest
func GetContestParticipants(contestId uint64) ([]ContestParticipant, error) {
	var participants []ContestParticipant
	if err := store.Find(&participants, bolthold.Where("ContestId").Eq(contestId)); err != nil {
		return nil, err
	}

	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Id < participants[j].Id
	})

	return participants, nil
}

// GetContestParticipantParticipation Participant registrations
func GetContestParticipantParticipation(participantId int64) ([]ContestParticipant, error) {
	var participants []ContestParticipant
	if err := store.Find(&participants, bolthold.Where("ParticipantId").Eq(participantId)); err != nil {
		return nil, err
	}
	return participants, nil
}

// GetContestParticipant Get one contest registration
func GetContestParticipant(id uint64) (*ContestParticipant, error) {
	var participant ContestParticipant
	if err := store.FindOne(&participant, bolthold.Where(bolthold.Key).Eq(id)); err != nil {
		return nil, err
	}
	return &participant, nil
}

// SaveContestParticipant Create new or update contest registration
func SaveContestParticipant(participant *ContestParticipant) error {
	if len(participant.Login) == 0 {
		participant.Login = "p_" + generateRandomString(5)
	}
	if len(participant.Password) == 0 {
		participant.Password = generateRandomString(10)
	}

	if participant.Id != 0 {
		return store.Update(participant.Id, participant)
	} else {
		return store.Insert(bolthold.NextSequence(), participant)
	}
}

// DeleteContestParticipant Delete contest registration
func DeleteContestParticipant(id uint64) error {
	return store.Delete(id, &ContestParticipant{})
}

func generateRandomString(length int) string {
	vowels := []rune{'e', 'u', 'i', 'o', 'a'}
	consonants := []rune{'q', 'r', 't', 'p', 's', 'd', 'g', 'h', 'k', 'z', 'x', 'v', 'b', 'n', 'm'}

	str := strings.Builder{}

	for i := 0; i < length; i += 2 {
		str.WriteRune(consonants[rand.Intn(len(consonants))])
		if i != length-1 {
			str.WriteRune(vowels[rand.Intn(len(vowels))])
		}
	}

	return str.String()
}

///////////////////////////////////////////////////////////////////////////////

// GetDialogState Get current dialog state
func GetDialogState(participantId int64) (*DialogState, error) {
	var state DialogState
	if err := store.FindOne(&state, bolthold.Where(bolthold.Key).Eq(participantId)); err != nil {
		if err == bolthold.ErrNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &state, nil
}

// SaveDialogState Save given participant dialog state
func SaveDialogState(state *DialogState) error {
	if state.ParticipantId == 0 {
		return errors.New("saving dialog state with empty ParticipantId")
	}
	if state.DialogType == "" {
		return errors.New("saving dialog state with empty DialogType")
	}
	if state.DialogStep == "" {
		return errors.New("saving dialog state with empty DialogStep")
	}
	return store.Upsert(state.ParticipantId, state)
}

// DeleteDialogState Remove given dialog state
func DeleteDialogState(participantId int64) error {
	return store.Delete(participantId, &DialogState{})
}

///////////////////////////////////////////////////////////////////////////////

// GetContestNotifications List all contest notifications
func GetContestNotifications(contestId uint64) ([]ContestNotification, error) {
	var notifications []ContestNotification
	if err := store.Find(&notifications, bolthold.Where("ContestId").Eq(contestId)); err != nil {
		return nil, err
	}

	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].Id < notifications[j].Id
	})

	return notifications, nil
}

// GetContestNotification Find given contest notification
func GetContestNotification(notificationId uint64) (*ContestNotification, error) {
	var notification ContestNotification
	if err := store.FindOne(&notification, bolthold.Where(bolthold.Key).Eq(notificationId)); err != nil {
		return nil, err
	}
	return &notification, nil
}

// SaveContestNotification Create new or update contest notification
func SaveContestNotification(notification *ContestNotification) error {
	if notification.Id != 0 {
		return store.Update(notification.Id, notification)
	} else {
		return store.Insert(bolthold.NextSequence(), notification)
	}
}

// DeleteContestNotification Remove given contest notification
func DeleteContestNotification(notificationId uint64) error {
	return store.Delete(notificationId, &ContestNotification{})
}
