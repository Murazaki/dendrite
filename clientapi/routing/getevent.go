// Copyright 2019 Alex Chen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routing

import (
	"net/http"

	"github.com/matrix-org/dendrite/clientapi/auth/authtypes"
	"github.com/matrix-org/dendrite/clientapi/jsonerror"
	"github.com/matrix-org/dendrite/internal/config"
	"github.com/matrix-org/dendrite/roomserver/api"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/matrix-org/util"
)

type getEventRequest struct {
	req            *http.Request
	device         *authtypes.Device
	roomID         string
	eventID        string
	cfg            *config.Dendrite
	federation     *gomatrixserverlib.FederationClient
	keyRing        gomatrixserverlib.KeyRing
	requestedEvent gomatrixserverlib.Event
}

// GetEvent implements GET /_matrix/client/r0/rooms/{roomId}/event/{eventId}
// https://matrix.org/docs/spec/client_server/r0.4.0.html#get-matrix-client-r0-rooms-roomid-event-eventid
func GetEvent(
	req *http.Request,
	device *authtypes.Device,
	roomID string,
	eventID string,
	cfg *config.Dendrite,
	rsAPI api.RoomserverInternalAPI,
	federation *gomatrixserverlib.FederationClient,
	keyRing gomatrixserverlib.KeyRing,
) util.JSONResponse {
	eventsReq := api.QueryEventsByIDRequest{
		EventIDs: []string{eventID},
	}
	var eventsResp api.QueryEventsByIDResponse
	err := rsAPI.QueryEventsByID(req.Context(), &eventsReq, &eventsResp)
	if err != nil {
		util.GetLogger(req.Context()).WithError(err).Error("queryAPI.QueryEventsByID failed")
		return jsonerror.InternalServerError()
	}

	if len(eventsResp.Events) == 0 {
		// Event not found locally
		return util.JSONResponse{
			Code: http.StatusNotFound,
			JSON: jsonerror.NotFound("The event was not found or you do not have permission to read this event"),
		}
	}

	requestedEvent := eventsResp.Events[0].Event

	r := getEventRequest{
		req:            req,
		device:         device,
		roomID:         roomID,
		eventID:        eventID,
		cfg:            cfg,
		federation:     federation,
		keyRing:        keyRing,
		requestedEvent: requestedEvent,
	}

	stateReq := api.QueryStateAfterEventsRequest{
		RoomID:       r.requestedEvent.RoomID(),
		PrevEventIDs: r.requestedEvent.PrevEventIDs(),
		StateToFetch: []gomatrixserverlib.StateKeyTuple{{
			EventType: gomatrixserverlib.MRoomMember,
			StateKey:  device.UserID,
		}},
	}
	var stateResp api.QueryStateAfterEventsResponse
	if err := rsAPI.QueryStateAfterEvents(req.Context(), &stateReq, &stateResp); err != nil {
		util.GetLogger(req.Context()).WithError(err).Error("queryAPI.QueryStateAfterEvents failed")
		return jsonerror.InternalServerError()
	}

	if !stateResp.RoomExists {
		util.GetLogger(req.Context()).Errorf("Expected to find room for event %s but failed", r.requestedEvent.EventID())
		return jsonerror.InternalServerError()
	}

	if !stateResp.PrevEventsExist {
		// Missing some events locally; stateResp.StateEvents unavailable.
		return util.JSONResponse{
			Code: http.StatusNotFound,
			JSON: jsonerror.NotFound("The event was not found or you do not have permission to read this event"),
		}
	}

	for _, stateEvent := range stateResp.StateEvents {
		if stateEvent.StateKeyEquals(r.device.UserID) {
			membership, err := stateEvent.Membership()
			if err != nil {
				util.GetLogger(req.Context()).WithError(err).Error("stateEvent.Membership failed")
				return jsonerror.InternalServerError()
			}
			if membership == gomatrixserverlib.Join {
				return util.JSONResponse{
					Code: http.StatusOK,
					JSON: gomatrixserverlib.ToClientEvent(r.requestedEvent, gomatrixserverlib.FormatAll),
				}
			}
		}
	}

	return util.JSONResponse{
		Code: http.StatusNotFound,
		JSON: jsonerror.NotFound("The event was not found or you do not have permission to read this event"),
	}
}
