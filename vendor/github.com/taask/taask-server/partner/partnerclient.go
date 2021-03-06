package partner

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cohix/simplcrypto"
	log "github.com/cohix/simplog"
	"github.com/pkg/errors"
	"github.com/taask/taask-server/auth"
	"github.com/taask/taask-server/timeout"
	"github.com/taask/taask-server/update"
	"google.golang.org/grpc"
)

// StartOutgoingManager continually tries to connect as the outgoing partner
func (m *Manager) StartOutgoingManager() {
	retry := 5

	for {
		// this loop ensures that we always reach out as the outgoing partner whenever needed
		err := m.Run()

		if err != nil {
			log.LogWarn(err.Error())
		} else {
			log.LogInfo("startOutgoingPartnerManager retrying...")
		}

		<-time.After(time.Duration(time.Second * time.Duration(retry)))
		retry *= 2
	}
}

// Run starts the partner manager
func (m *Manager) Run() error {
	log.LogInfo("PartnerManager Run")
	// defer m.lockUnlockPartner()

	if m.partner == nil {
		partner := &Partner{
			Update:     update.NewPartnerUpdate(),
			host:       m.config.Service.Host,
			port:       m.config.Service.Port,
			updateLock: &sync.Mutex{},
		}

		m.partner = partner
	}

	if m.partner.HealthChecker != nil {
		if m.partner.HealthChecker.IsHealthy {
			log.LogInfo("partner healthChecker is healthy")
			return nil
		}

		log.LogInfo("partner healthChecker is unhealthy, attempting to connect to partner")
	} else {
		log.LogInfo("partner healthChecker doesn't exist, attempting to connect to partner")
	}

	if err := m.initPartnerClient(m.partner); err != nil {
		return errors.Wrap(err, "PartnerManager failed to initClient, will retry")
	}

	if err := m.authenticatePartner(m.partner); err != nil {
		return errors.Wrap(err, "PartnerManager failed to authenticatePartner, will retry")
	}

	log.LogInfo("authenticated with partner, starting update stream")

	var client PartnerService_StreamUpdatesClient

	if m.partner.Client != nil {
		var err error
		client, err = m.partner.Client.StreamUpdates(context.Background())
		if err != nil {
			return errors.Wrap(err, "PartnerManager failed to StreamUpdates, will retry in 5s")
		}
	} else {
		log.LogInfo("m.partner.Client doesn't exist, aborting...")
		return nil
	}

	log.LogInfo("sending session to partner")
	if err := m.sendSessionToPartner(client); err != nil {
		return errors.Wrap(err, "PartnerManager failed to sendSessionTopartner")
	}

	m.partner.HealthChecker = newHealthChecker()
	go m.partner.HealthChecker.startHealthCheckingWithClient(client)

	log.LogInfo("waiting for partner data key")
	if err := m.receiveDataKeyFromPartner(m.partner, client); err != nil {
		return errors.Wrap(err, "PartnerManager failed to receiveDataKeyFromPartner, will retry in 5s")
	}

	log.LogInfo("received data key from partner")
	sendChan, recvChan := m.clientSendRecvChans(m.partner, client)

	log.LogInfo("update stream starting")
	if err := m.streamUpdates(sendChan, recvChan, m.partner.HealthChecker.UnhealthyChan); err != nil {
		return errors.Wrap(err, "PartnerManager encountered streamUpdates error, will retry")
	}

	return nil
}

// clientSendRecvChans creates channels to send and receive from the partner, and continuously reads and writes
func (m *Manager) clientSendRecvChans(partner *Partner, client PartnerService_StreamUpdatesClient) (chan update.PartnerUpdate, chan update.PartnerUpdate) {
	sendChan := make(chan update.PartnerUpdate)
	recvChan := make(chan update.PartnerUpdate)

	go func(client PartnerService_StreamUpdatesClient, sendChan chan update.PartnerUpdate, recvChan chan update.PartnerUpdate) {
		for {
			select {
			case update := <-sendChan:
				updateReq, err := m.encryptAndSignUpdateForPartner(partner, &update)
				if err != nil {
					log.LogWarn(errors.Wrap(err, "clientSendRecvChans failed to encryptAndSignUpdateForPartner").Error())
					continue
				}

				if err := client.Send(updateReq); err != nil {
					log.LogWarn(errors.Wrap(err, "streamClientRecvChan failed to Recv, terminating partner stream...").Error())
					break
				}
			default:
				updateReq, err := client.Recv()
				if err != nil {
					log.LogWarn(errors.Wrap(err, "streamClientRecvChan failed to Recv, terminating partner stream...").Error())
					break
				}

				update, err := m.decryptAndVerifyUpdateFromPartner(partner, updateReq)
				if err != nil {
					log.LogWarn(errors.Wrap(err, "failed to decryptAndVerifyUpdateFromPartner").Error())
					continue
				} else if update == nil {
					log.LogInfo("received health check, discarding...")
					continue
				}

				recvChan <- *update
			}
		}
	}(client, sendChan, recvChan)

	return sendChan, recvChan
}

func (m *Manager) initPartnerClient(partner *Partner) error {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", partner.host, partner.port), grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to Dial")
	}

	client := NewPartnerServiceClient(conn)

	partner.Client = client

	return nil
}

func (m *Manager) authenticatePartner(partner *Partner) error {
	keypair, err := simplcrypto.GenerateNewKeyPair()
	if err != nil {
		return errors.Wrap(err, "failed to GenerateNewKeyPair")
	}

	timestamp := time.Now().Unix()

	nonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonce, uint64(timestamp))
	hashWithNonce := append(m.config.MemberGroup.AuthHash, nonce...)

	authHashSig, err := keypair.Sign(hashWithNonce)
	if err != nil {
		return errors.Wrap(err, "failed to Sign")
	}

	attempt := &auth.Attempt{
		MemberUUID:        m.UUID,
		GroupUUID:         m.config.MemberGroup.UUID,
		PubKey:            keypair.SerializablePubKey(),
		AuthHashSignature: authHashSig,
		Timestamp:         timestamp,
	}

	log.LogInfo("sending partner auth attempt")
	authResp, err := partner.Client.AuthPartner(timeout.AuthContext(), attempt)
	if err != nil {
		return errors.Wrap(err, "failed to AuthPartner")
	}

	log.LogInfo("partner auth attempt succeeded")
	challengeBytes, err := keypair.Decrypt(authResp.EncChallenge)
	if err != nil {
		return errors.Wrap(err, "failed to Decrypt challenge")
	}

	masterRunnerPubKey, err := simplcrypto.KeyPairFromSerializedPubKey(authResp.MasterPubKey)
	if err != nil {
		return errors.Wrap(err, "failed to KeyPairFromSerializablePubKey")
	}

	challengeSig, err := keypair.Sign(challengeBytes)
	if err != nil {
		return errors.Wrap(err, "failed to Sign challenge")
	}

	session := &activeSession{
		Session: &auth.Session{
			MemberUUID:          m.UUID,
			GroupUUID:           m.config.MemberGroup.UUID,
			SessionChallengeSig: challengeSig,
		},
		Keypair:      keypair,
		MasterPubKey: masterRunnerPubKey,
	}

	partner.ActiveSession = session

	return nil
}

func (m *Manager) sendSessionToPartner(client PartnerService_StreamUpdatesClient) error {
	updateReq := &UpdateRequest{
		Session: m.partner.ActiveSession.Session,
	}

	return client.Send(updateReq)
}

func (m *Manager) receiveDataKeyFromPartner(partner *Partner, client PartnerService_StreamUpdatesClient) error {
	updateReq, err := client.Recv()
	if err != nil {
		return errors.Wrap(err, "failed to Recv data key update")
	}

	if updateReq.UpdateSignature == nil {
		return errors.New("data key update signature missing")
	}

	if updateReq.EncDataKey == nil {
		return errors.New("data key update encDataKey is empty")
	}

	dataKeyJSON, err := partner.ActiveSession.Keypair.Decrypt(updateReq.EncDataKey)
	if err != nil {
		return errors.Wrap(err, "failed to Decrypt data key update")
	}

	if err := partner.ActiveSession.MasterPubKey.Verify(dataKeyJSON, updateReq.UpdateSignature); err != nil {
		return errors.Wrap(err, "failed to Verify data key update")
	}

	dataKey := simplcrypto.SymKey{}
	if err := json.Unmarshal(dataKeyJSON, &dataKey); err != nil {
		return errors.Wrap(err, "failed to Unmarshal data key")
	}

	partner.UUID = updateReq.PartnerUUID
	partner.DataKey = &dataKey

	return nil
}
