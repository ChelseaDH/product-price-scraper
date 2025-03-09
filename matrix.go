package main

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
	"sync"
)

type MatrixClient struct {
	client     *mautrix.Client
	roomId     id.RoomID
	cancelSync context.CancelFunc
}

func connectToMatrix(ctx context.Context, logger *slog.Logger, config Matrix, dbPath string) (*MatrixClient, error) {
	client, err := mautrix.NewClient(config.HomeServer, id.UserID(config.UserName), config.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("Error creating client: %v\n", err)
	}

	whoAmI, err := client.Whoami(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error getting whoami: %v\n", err)
	}

	client.UserID = whoAmI.UserID
	client.DeviceID = whoAmI.DeviceID
	client.AccessToken = config.AccessToken

	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	// Capture the room ID of a received message
	syncer.OnEventType(event.EventMessage, func(ctx context.Context, evt *event.Event) {
		if evt.Sender != client.UserID {
			logger.Info(fmt.Sprintf("Received message in room %s: %s\n", evt.RoomID, evt.Content.AsMessage().Body))
		}
	})
	// Join or leave a room after receiving an invitation
	syncer.OnEventType(event.StateMember, func(ctx context.Context, evt *event.Event) {
		if evt.GetStateKey() != client.UserID.String() {
			return
		}

		if evt.Content.AsMember().Membership == event.MembershipInvite {
			if evt.Sender.Homeserver() == client.UserID.Homeserver() {
				logger.Info(fmt.Sprintf("Joining room %s after invite from %s", evt.RoomID, evt.Sender))
				_, err = client.JoinRoomByID(ctx, evt.RoomID)
			} else {
				logger.Info(fmt.Sprintf("Declining invite from %s to join room %s", evt.Sender, evt.RoomID))
				_, err = client.LeaveRoom(ctx, evt.RoomID)
			}

			if err != nil {
				LogError(logger, "Failed to join room after invite", err)
			}
		} else {
			logger.Info(fmt.Sprintf("Room membership changed to %s in room %s", evt.Content.AsMember().Membership, evt.RoomID))
		}
	})

	crypto, err := cryptohelper.NewCryptoHelper(client, []byte("prices"), dbPath)
	if err != nil {
		return nil, fmt.Errorf("Error creating cryptohelper: %v\n", err)
	}

	err = crypto.Init(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error initializing cryptohelper: %v\n", err)
	}
	// Set the client crypto helper to automatically encrypt outgoing messages
	client.Crypto = crypto

	logger.Info(fmt.Sprintf("Connected to %s", config.HomeServer))

	syncCtx, cancelSync := context.WithCancel(ctx)
	var syncStopWait sync.WaitGroup
	syncStopWait.Add(1)

	cancelSyncAndWait := func() {
		cancelSync()
		syncStopWait.Wait()
	}

	go func() {
		defer syncStopWait.Done()
		err = client.SyncWithContext(syncCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			LogError(logger, "Failed to sync", err)
		}
	}()

	return &MatrixClient{client: client, roomId: id.RoomID(config.RoomID), cancelSync: cancelSyncAndWait}, nil
}

func (m *MatrixClient) Stop() error {
	m.cancelSync()
	if ch, ok := m.client.Crypto.(*cryptohelper.CryptoHelper); ok {
		err := ch.Close()
		if err != nil {
			return fmt.Errorf("Error closing cryptohelper: %v\n", err)
		}
	}

	return nil
}

func (m *MatrixClient) SendMessage(markdown string) error {
	content := format.RenderMarkdown(markdown, true, true)

	_, err := m.client.SendMessageEvent(context.Background(), m.roomId, event.EventMessage, content)
	return err
}
