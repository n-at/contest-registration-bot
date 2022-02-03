package storage

import (
	"errors"
	"github.com/timshannon/bolthold"
	"sort"
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
	err := store.Find(&contests, nil)
	if err != nil {
		return nil, err
	}
	return contestsSorted(contests), nil
}

// GetContest One contest by id
func GetContest(id uint64) (*Contest, error) {
	var contest Contest
	err := store.FindOne(&contest, bolthold.Where(bolthold.Key).Eq(id))
	return &contest, err
}

// SaveContest Create new or update contest
func SaveContest(contest *Contest) error {
	if contest.Id != 0 {
		return store.Update(contest.Id, contest)
	} else {
		return store.Insert(bolthold.NextSequence(), contest)
	}
}

// GetParticipantContests List all contests where given participant registered
func GetParticipantContests(participantId string) ([]Contest, error) {
	participants, err := GetContestParticipantParticipation(participantId)
	if err != nil {
		return nil, err
	}

	var contestIds []uint64
	for _, participant := range participants {
		contestIds = append(contestIds, participant.ContestId)
	}

	var contests []Contest
	err = store.Find(&contests, bolthold.Where(bolthold.Key).In(contestIds))
	if err != nil {
		return nil, err
	}
	return contestsSorted(contests), nil
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
	err := store.Find(&participants, bolthold.Where("ContestId").Eq(contestId))
	if err != nil {
		return nil, err
	}

	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Id < participants[j].Id
	})

	return participants, nil
}

// GetContestParticipantParticipation Participant registrations
func GetContestParticipantParticipation(participantId string) ([]ContestParticipant, error) {
	var participants []ContestParticipant
	err := store.Find(&participants, bolthold.Where("ParticipantId").Eq(participantId))
	return participants, err
}

// GetContestParticipant Get one contest registration
func GetContestParticipant(id uint64) (*ContestParticipant, error) {
	var participant ContestParticipant
	err := store.FindOne(&participant, bolthold.Where(bolthold.Key).Eq(id))
	return &participant, err
}

// SaveContestParticipant Create new or update contest registration
func SaveContestParticipant(participant *ContestParticipant) error {
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

///////////////////////////////////////////////////////////////////////////////

// GetRegistrationState Get current registration state
func GetRegistrationState(participantId string) (*RegistrationState, error) {
	var state RegistrationState
	err := store.FindOne(&state, bolthold.Where(bolthold.Key).Eq(participantId))
	return &state, err
}

// SaveRegistrationState Save given participant registration state
func SaveRegistrationState(state *RegistrationState) error {
	if state.ParticipantId == "" {
		return errors.New("saving registration state with empty ParticipantId")
	}
	return store.Upsert(state.ParticipantId, state)
}

///////////////////////////////////////////////////////////////////////////////

// GetContestNotifications List all contest notifications
func GetContestNotifications(contestId uint64) ([]ContestNotification, error) {
	var notifications []ContestNotification
	err := store.Find(&notifications, bolthold.Where(bolthold.Key).Eq(contestId))
	if err != nil {
		return nil, err
	}

	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].Id < notifications[j].Id
	})

	return notifications, err
}

// SaveContestNotification Create new or update contest notification
func SaveContestNotification(notification *ContestNotification) error {
	if notification.Id != 0 {
		return store.Update(notification.Id, notification)
	} else {
		return store.Insert(bolthold.NextSequence(), notification)
	}
}
