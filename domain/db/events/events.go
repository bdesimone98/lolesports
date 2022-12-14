package events

import (
	"context"
	"fmt"
	"github.com/apex/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	api "lolesports/domain/service"
	"lolesports/utils"
	"strings"
	"time"
)

const (
	mongoQueryTimeout          = 10 * time.Second
	eventCollectionName string = "events"
)

func (repository *scenesRepository) Add(ctx context.Context, event AllEventData) (*AllEventData, error) {
	var cancel context.CancelFunc
	if ctx == nil {
		ctx, cancel = context.WithTimeout(context.Background(), mongoQueryTimeout)
		defer cancel()
	}

	logCTX := utils.BaseLogCtx()

	options := options.FindOneAndUpdate().SetUpsert(true)

	response := repository.db.Collection(eventCollectionName).FindOneAndUpdate(
		ctx,
		bson.M{"event.id": event.Event.Id},
		bson.M{"$set": event},
		options)

	if response.Err() != nil && response.Err().Error() != "mongo: no documents in result" {
		utils.WithFields(logCTX, log.Fields{
			"context": "event_repository_add",
			"error":   strings.ReplaceAll(response.Err().Error(), " ", "_"),
		})
		logCTX.Level = log.ErrorLevel
		repository.logger.HandleLog(logCTX)
		return nil, response.Err()
	}

	var savedEvent AllEventData
	response.Decode(&savedEvent)

	utils.WithFields(logCTX, log.Fields{
		"context": "event_repository_add",
		"detail":  fmt.Sprintf("Saved_event:%s", savedEvent.Event.Id),
	})
	repository.logger.HandleLog(logCTX)

	return &savedEvent, nil
}

func (repository *scenesRepository) UpdateMany(filter bson.D, set bson.D) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongoQueryTimeout)
	defer cancel()

	logCTX := utils.BaseLogCtx()
	_, err := repository.db.Collection(eventCollectionName).UpdateMany(
		ctx,
		filter,
		set,
	)

	if err != nil {
		utils.WithFields(logCTX, log.Fields{
			"context": "event_update_many",
			"error":   strings.ReplaceAll(err.Error(), " ", "_"),
		})
		logCTX.Level = log.ErrorLevel
		repository.logger.HandleLog(logCTX)
		return err
	}
	return nil
}

func (repository *scenesRepository) Find(filter bson.M, findOption *options.FindOptions) ([]*AllEventData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoQueryTimeout)
	defer cancel()

	logCTX := utils.BaseLogCtx()
	response, err := repository.db.Collection(eventCollectionName).Find(
		ctx,
		filter,
		findOption,
	)

	if err != nil {
		utils.WithFields(logCTX, log.Fields{
			"context": "events_find",
			"error":   strings.ReplaceAll(err.Error(), " ", "_"),
		})
		logCTX.Level = log.ErrorLevel
		repository.logger.HandleLog(logCTX)
		return nil, err
	}
	defer response.Close(ctx)

	var events []*AllEventData
	err = response.All(ctx, &events)
	if err != nil {
		utils.WithFields(logCTX, log.Fields{
			"context": "events_find",
			"error":   strings.ReplaceAll(err.Error(), " ", "_"),
		})
		logCTX.Level = log.ErrorLevel
		repository.logger.HandleLog(logCTX)
		return nil, err
	}

	return events, nil
}

func (repository *scenesRepository) FindOne(filter bson.M) (*AllEventData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoQueryTimeout)
	defer cancel()

	logCTX := utils.BaseLogCtx()

	response := repository.db.Collection(eventCollectionName).FindOne(
		ctx,
		filter,
	)

	if response.Err() != nil {
		return nil, response.Err()
	}

	var event *AllEventData
	err := response.Decode(&event)
	if err != nil {
		utils.WithFields(logCTX, log.Fields{
			"context": "event_parse",
			"error":   strings.ReplaceAll(response.Err().Error(), " ", "_"),
		})
		logCTX.Level = log.ErrorLevel
		repository.logger.HandleLog(logCTX)
		return nil, response.Err()
	}

	return event, nil
}

func (repository *scenesRepository) AddWindowFrames2Event(ctx context.Context, eventID string, lastFrame int, frames []api.WindowFrame) error {
	var cancel context.CancelFunc
	if ctx == nil {
		ctx, cancel = context.WithTimeout(context.Background(), mongoQueryTimeout)
		defer cancel()
	}

	logCTX := utils.BaseLogCtx()

	_, err := repository.db.Collection(eventCollectionName).UpdateOne(
		ctx,
		bson.D{{"event.id", eventID}},
		bson.D{
			{"$set", bson.D{
				{"last_window_frame", lastFrame},
			}},
			{"$push", bson.D{
				{"window_frames", bson.D{
					{"$each", frames},
					{"$sort", bson.D{{"second", 1}}},
				}},
			}},
		},
	)

	if err != nil {
		utils.WithFields(logCTX, log.Fields{
			"context": "add_frame_to_event",
			"error":   strings.ReplaceAll(err.Error(), " ", "_"),
		})
		logCTX.Level = log.ErrorLevel
		repository.logger.HandleLog(logCTX)
		return err
	}

	utils.WithFields(logCTX, log.Fields{
		"detail": "add_frame_to_event",
		"event":  strings.ReplaceAll(eventID, " ", "_"),
	})
	logCTX.Level = log.DebugLevel
	repository.logger.HandleLog(logCTX)

	return nil
}

func (repository *scenesRepository) AddFrames2Event(ctx context.Context, eventID string, lastFrame int, frames []api.Frame) error {
	var cancel context.CancelFunc
	if ctx == nil {
		ctx, cancel = context.WithTimeout(context.Background(), mongoQueryTimeout)
		defer cancel()
	}

	logCTX := utils.BaseLogCtx()

	_, err := repository.db.Collection(eventCollectionName).UpdateOne(
		ctx,
		bson.D{{"event.id", eventID}},
		bson.D{
			{"$set", bson.D{
				{"last_frame", lastFrame},
			}},
			{"$push", bson.D{
				{"frames", bson.D{
					{"$each", frames},
					{"$sort", bson.D{{"second", 1}}},
				}},
			}},
		},
	)

	if err != nil {
		utils.WithFields(logCTX, log.Fields{
			"context": "add_frame_to_event",
			"error":   strings.ReplaceAll(err.Error(), " ", "_"),
		})
		logCTX.Level = log.ErrorLevel
		repository.logger.HandleLog(logCTX)
		return err
	}

	utils.WithFields(logCTX, log.Fields{
		"detail": "add_frame_to_event",
		"event":  strings.ReplaceAll(eventID, " ", "_"),
	})
	logCTX.Level = log.DebugLevel
	repository.logger.HandleLog(logCTX)

	return nil
}

func (repository *scenesRepository) GetNotFetched() ([]*AllEventData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongoQueryTimeout)
	defer cancel()

	projection := bson.A{
		bson.M{"$project": bson.M{
			"event": 1,
			"games": bson.M{"$objectToArray": "$games"},
		}},

		bson.M{"$unwind": bson.M{"path": "$games"}},
		bson.M{"$match": bson.M{"games.v.fetched": false}},
		bson.M{"$group": bson.M{
			"_id": "$event.id",
			"games": bson.M{
				"$push": "$games",
			},
			"event": bson.M{
				"$first": "$event",
			},
		}},
		bson.M{"$project": bson.M{
			"event": 1,
			"games": bson.M{"$arrayToObject": "$games"},
		}},
	}

	logCTX := utils.BaseLogCtx()
	response, err := repository.db.Collection(eventCollectionName).Aggregate(
		ctx,
		projection,
		nil,
	)

	if err != nil {
		utils.WithFields(logCTX, log.Fields{
			"context": "events_find",
			"error":   strings.ReplaceAll(err.Error(), " ", "_"),
		})
		logCTX.Level = log.ErrorLevel
		repository.logger.HandleLog(logCTX)
		return nil, err
	}
	defer response.Close(ctx)

	var events []*AllEventData
	err = response.All(ctx, &events)
	if err != nil {
		utils.WithFields(logCTX, log.Fields{
			"context": "events_find",
			"error":   strings.ReplaceAll(err.Error(), " ", "_"),
		})
		logCTX.Level = log.ErrorLevel
		repository.logger.HandleLog(logCTX)
		return nil, err
	}

	return events, nil

}
