// Copyright 2017 Vector Creations Ltd
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

package federationapi

import (
	appserviceAPI "github.com/matrix-org/dendrite/appservice/api"
	"github.com/matrix-org/dendrite/clientapi/auth/storage/accounts"
	"github.com/matrix-org/dendrite/clientapi/auth/storage/devices"
	federationSenderAPI "github.com/matrix-org/dendrite/federationsender/api"
	"github.com/matrix-org/dendrite/internal/basecomponent"
	roomserverAPI "github.com/matrix-org/dendrite/roomserver/api"

	// TODO: Are we really wanting to pull in the producer from clientapi
	"github.com/matrix-org/dendrite/clientapi/producers"
	"github.com/matrix-org/dendrite/federationapi/routing"
	"github.com/matrix-org/gomatrixserverlib"
)

// SetupFederationAPIComponent sets up and registers HTTP handlers for the
// FederationAPI component.
func SetupFederationAPIComponent(
	base *basecomponent.BaseDendrite,
	accountsDB accounts.Database,
	deviceDB devices.Database,
	federation *gomatrixserverlib.FederationClient,
	keyRing *gomatrixserverlib.KeyRing,
	rsAPI roomserverAPI.RoomserverInternalAPI,
	asAPI appserviceAPI.AppServiceQueryAPI,
	federationSenderAPI federationSenderAPI.FederationSenderInternalAPI,
	eduProducer *producers.EDUServerProducer,
) {
	roomserverProducer := producers.NewRoomserverProducer(rsAPI)

	routing.Setup(
		base.PublicAPIMux, base.Cfg, rsAPI, asAPI, roomserverProducer,
		eduProducer, federationSenderAPI, *keyRing,
		federation, accountsDB, deviceDB,
	)
}
