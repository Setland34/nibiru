package testutil

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/gogoproto/proto"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func RequireNotHasTypedEvent(t require.TestingT, ctx sdk.Context, event proto.Message) {
	name := proto.MessageName(event)
	for _, ev := range ctx.EventManager().Events() {
		if ev.Type == name {
			t.Errorf("unexpected event found: %s", name)
		}
	}
}

func RequireHasTypedEvent(t require.TestingT, ctx sdk.Context, event proto.Message) {
	for _, abciEvent := range ctx.EventManager().Events() {
		if abciEvent.Type != proto.MessageName(event) {
			continue
		}
		typedEvent, err := sdk.ParseTypedEvent(abci.Event{
			Type:       abciEvent.Type,
			Attributes: abciEvent.Attributes,
		})
		require.NoError(t, err)

		require.EqualValues(t, event, typedEvent, "events do not match")
		return
	}

	t.Errorf("event not found")
}

func RequireContainsTypedEvent(t require.TestingT, ctx sdk.Context, event proto.Message) {
	foundEvents := []proto.Message{}
	for _, abciEvent := range ctx.EventManager().Events() {
		eventName := proto.MessageName(event)
		if abciEvent.Type != eventName {
			continue
		}
		typedEvent, err := sdk.ParseTypedEvent(abci.Event{
			Type:       abciEvent.Type,
			Attributes: abciEvent.Attributes,
		})
		require.NoError(t, err)

		if reflect.DeepEqual(typedEvent, event) {
			return
		} else {
			foundEvents = append(foundEvents, typedEvent)
		}
	}

	t.Errorf("event not found, event: %+v, found events: %+v", event, foundEvents)
}

// EventHasAttributeValue parses the given ABCI event at a key to see if it
// matches (contains) the wanted value.
//
// Args:
//   - abciEvent: The event under test
//   - key: The key for which we'll check the value
//   - want: The desired value
func EventHasAttributeValue(abciEvent sdk.Event, key string, want string) error {
	attr, ok := abciEvent.GetAttribute(key)
	if !ok {
		return fmt.Errorf("abci event does not contain key: %s", key)
	}
	got := attr.Value

	if !strings.Contains(got, want) {
		return fmt.Errorf("expected %s %s, got %s", key, want, got)
	}

	return nil
}
