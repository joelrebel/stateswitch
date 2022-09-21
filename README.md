[![Actions Status](https://github.com/filanov/stateswitch/workflows/make_all/badge.svg)](https://github.com/filanov/stateswitch/actions)

# stateswitch

## Overview

A simple and clear way to create and represent state machine.

```go
sm := stateswitch.NewStateMachine()

// Define the state machine rules (and optionally document each rule)
sm.AddTransition(stateswitch.TransitionRule{
    TransitionType:   TransitionTypeSetHwInfo,
    SourceStates:     stateswitch.States{StateDiscovering, StateKnown, StateInsufficient},
    DestinationState: StateKnown,
    Condition:        th.IsSufficient,
    Transition:       th.SetHwInfo,
    PostTransition:   th.PostSetHwInfo,
    Documentation: stateswitch.TransitionRuleDoc{
        Name:        "Move to known when receiving good hardware information",
        Description: "Once we receive hardware information from a server, we can consider it known if the hardware information is sufficient",
    },
})
sm.AddTransition(stateswitch.TransitionRule{
    TransitionType:   TransitionTypeSetHwInfo,
    SourceStates:     stateswitch.States{StateDiscovering, StateKnown, StateInsufficient},
    DestinationState: StateInsufficient,
    Condition:        th.IsInsufficient,
    Transition:       th.SetHwInfo,
    PostTransition:   th.PostSetHwInfo,
    Documentation: stateswitch.TransitionRuleDoc{
        Name:        "Move to insufficient when receiving bad hardware information",
        Description: "Once we receive hardware infomration from a server, we consider the server to be insufficient if the hardware is insufficient",
    },
})
sm.AddTransition(stateswitch.TransitionRule{
    TransitionType:   TransitionTypeRegister,
    SourceStates:     stateswitch.States{""},
    DestinationState: StateDiscovering,
    Condition:        nil,
    Transition:       nil,
    PostTransition:   th.RegisterNew,
    Documentation: stateswitch.TransitionRuleDoc{
        Name:        "Initial registration",
        Description: "A new server which registers enters our initial discovering state",
    },
})
sm.AddTransition(stateswitch.TransitionRule{
    TransitionType:   TransitionTypeRegister,
    SourceStates:     stateswitch.States{StateDiscovering, StateKnown, StateInsufficient},
    DestinationState: StateDiscovering,
    Condition:        nil,
    Transition:       nil,
    PostTransition:   th.RegisterAgain,
    Documentation: stateswitch.TransitionRuleDoc{
        Name:        "Re-registration",
        Description: "We should ignore repeated registrations from servers that are already registered",
    },
})

// Document transition types (optional)
sm.DescribeTransitionType(TransitionTypeSetHwInfo, stateswitch.TransitionTypeDoc{
    Name:        "Set hardware info",
    Description: "Triggered for every hardware information change",
})
sm.DescribeTransitionType(TransitionTypeRegister, stateswitch.TransitionTypeDoc{
    Name:        "Register",
    Description: "Triggered when a server registers",
})

// Document possible states (optional)
sm.DescribeState(StateDiscovering, stateswitch.StateDoc{
    Name:        "Discovering",
    Description: "Indicates that the server has registered but we still don't know anything about its hardware",
})
sm.DescribeState(StateKnown, stateswitch.StateDoc{
    Name:        "Discovering",
    Description: "Indicates that the server has registered but we still don't know anything about its hardware",
})
sm.DescribeState(StateInsufficient, stateswitch.StateDoc{
    Name:        "Insufficient",
    Description: "Indicates that the server has sufficient hardware",
})
```

## Usage

First your state object need to implement the state interface:

```go
type StateSwitch interface {
	// State return current state
	State() State
	// SetState set a new state
	SetState(state State) error
}
```

Then you need to create state machine

```go
sm := stateswitch.NewStateMachine()
```

Add transitions with the expected behavior 
```go
sm.AddTransition(stateswitch.TransitionRule{
    TransitionType:   TransitionTypeSetHwInfo,
    SourceStates:     stateswitch.States{StateDiscovering, StateKnown, StateInsufficient},
    DestinationState: StateInsufficient,
    Condition:        th.IsInsufficient,
    Transition:       th.SetHwInfo,
    PostTransition:   th.PostSetHwInfo,
})
```

`TransitionRule` define the behavior that will be selected for your object by transition type,
source state and conditions that you define.
The first transition that will satisfy those requirements will be activated. 
`Condtion`, `Transition`, `PostTranstion` and `Documentation` are all optional, the transition may only change the state.

Since `Condtion` represent boolean entity, stateswitch provides means to create a combination of these entities from basic 
boolean operations: `Not`,`And`, `Or`.  For example, rule with complex condition:

```go
sm.AddTransition(stateswitch.TransitionRule{
    TransitionType:   TransitionTypeSetHwInfo,
    SourceStates:     stateswitch.States{StateDiscovering, StateKnown, StateInsufficient},
    DestinationState: StatePending,
    Condition:        And(th.IsConnected, th.HasInventory, Not(th.RoleDefined)),
    Transition:       th.SetHwInfo,
    PostTransition:   th.PostSetHwInfo,
})
```

Run transition by type, state machine will select the right one for you.

```go
h.sm.Run(TransitionTypeSetHwInfo, &stateHost{host: host}, &TransitionArgsSetHwInfo{hwInfo: hw})
```

for more details and full examples take a look at the examples section.

### State machine representation

Once a state-machine has been initialized, you can generate a JSON file that describes it by using `AsJSON`:

```go
machineJSON, err := sm.AsJSON()
if err != nil {
    panic(err)
}

fmt.Println(string(machineJSON))
```

This file can be used, for example, for generating documentation for your state machine.

In the above example, this results in the following JSON:

```json
{
  "transition_rules": [
    {
      "transition_type": "Register",
      "source_states": [
        "initial"
      ],
      "destination_state": "discovering",
      "name": "Initial registration",
      "description": "A new server which registers enters our initial discovering state"
    },
    {
      "transition_type": "Register",
      "source_states": [
        "discovering",
        "known",
        "insufficient"
      ],
      "destination_state": "discovering",
      "name": "Re-registration",
      "description": "We should ignore repeated registrations from servers that are already registered"
    },
    {
      "transition_type": "SetHwInfo",
      "source_states": [
        "discovering",
        "known",
        "insufficient"
      ],
      "destination_state": "known",
      "name": "Move to known when receiving good hardware information",
      "description": "Once we receive hardware information from a server, we can consider it known if the hardware information is sufficient"
    },
    {
      "transition_type": "SetHwInfo",
      "source_states": [
        "discovering",
        "known",
        "insufficient"
      ],
      "destination_state": "insufficient",
      "name": "Move to insufficient when receiving bad hardware information",
      "description": "Once we receive hardware infomration from a server, we consider the server to be insufficient if the hardware is insufficient"
    }
  ],
  "states": {
    "discovering": {
      "name": "Discovering",
      "description": "Indicates that the server has registered but we still don't know anything about its hardware"
    },
    "initial": {
      "name": "Initial",
      "description": "The initial state of the state machine. This is a synthetic state that is not actually part of the state machine. It appears in documentation when transition rules hold a single source state that is an empty string"
    },
    "insufficient": {
      "name": "Insufficient",
      "description": "Indicates that the server has sufficient hardware"
    },
    "known": {
      "name": "Discovering",
      "description": "Indicates that the server has registered but we still don't know anything about its hardware"
    }
  },
  "transition_types": {
    "Register": {
      "name": "Register",
      "description": "Triggered when a server registers"
    },
    "SetHwInfo": {
      "name": "Set hardware info",
      "description": "Triggered for every hardware information change"
    }
  }
}
```

## Examples

Example can be found [here](https://github.com/filanov/stateswitch/tree/master/examples)
