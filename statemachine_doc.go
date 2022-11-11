package stateswitch

import (
	"encoding/json"
	"fmt"
	"sort"
)

type StateMachineDocumentation interface {
	// DescribeState lets you optionally add documentation for a particular
	// [State]. This description will be included in the JSON generated by
	// AsJSON
	DescribeState(state State, stateDocumentation StateDoc)

	// DescribeTransitionType lets you optionally add documentation for a
	// particular [TransitionType]. This description will be included in the
	// JSON generated by AsJSON
	DescribeTransitionType(transitionType TransitionType, transitionTypeDocumentation TransitionTypeDoc)

	// Generates a machine-readable JSON representation of the state machine
	// states and transitions. See StateMachineJSON for the format. Such JSON
	// can be used to generate documentation or to generate a state machine
	// diagram
	AsJSON() ([]byte, error)

	Export() StateMachineJSON
}

type stateMachineDocumentation struct {
	stateDocs          map[State]StateDoc
	transitionTypeDocs map[TransitionType]TransitionTypeDoc
}

func initStateMachineDocumentation(sm *stateMachine) {
	sm.stateDocs = make(map[State]StateDoc)
	sm.transitionTypeDocs = make(map[TransitionType]TransitionTypeDoc)

	sm.DescribeState(State("initial"), StateDoc{
		Name:        "Initial",
		Description: "The initial state of the state machine. This is a synthetic state that is not actually part of the state machine. It appears in documentation when transition rules hold a single source state that is an empty string",
	})
}

type StateDoc struct {
	// A human readable name for the state
	Name string

	// A more verbose description of the state
	Description string
}

type TransitionTypeDoc struct {
	// A human readable name for the transition type
	Name string

	// A more verbose description of the transition type
	Description string
}

type TransitionRuleDoc struct {
	// A short name for the transition rule
	Name string

	// A more verbose description of the transition rule
	Description string
}

type StateMachineJSON struct {
	TransitionRuleNodes []TransitionRuleNode          `json:"transition_rules_nodes"`
	TransitionRuleEdges []TransitionRuleEdge          `json:"transition_rules_edges"`
	TransitionRules     []TransitionRuleJSON          `json:"transition_rules"`
	States              map[string]StateJSON          `json:"states"`
	TransitionTypes     map[string]TransitionTypeJSON `json:"transition_types"`
}

type StateJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TransitionTypeJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TransitionRuleEdge struct {
	From        string `json:"from,omitempty"` // from is set if this is an edge
	To          string `json:"to,omitempty"`   // to is set if this is an edge
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"` // assign a label to the edge
}

type TransitionRuleNode struct {
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
}

type TransitionRuleJSON struct {
	TransitionType   TransitionType `json:"transition_type"`
	SourceStates     []string       `json:"source_states"`
	DestinationState string         `json:"destination_state"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
}

type StateDocJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (sm *stateMachine) DescribeState(state State, stateDocumentation StateDoc) {
	sm.stateDocs[state] = stateDocumentation
}

func (sm *stateMachine) DescribeTransitionType(transitionType TransitionType, transitionTypeDocumentation TransitionTypeDoc) {
	sm.transitionTypeDocs[transitionType] = transitionTypeDocumentation
}

func (sm *stateMachine) Export() StateMachineJSON {
	// Generate a sorted list of all states to avoid non-deterministic JSON
	// output
	keys := make([]TransitionType, 0, len(sm.transitionRules))
	for tt := range sm.transitionRules {
		keys = append(keys, tt)
	}
	sort.Slice(keys, func(i, j int) bool {
		return string(keys[i]) < string(keys[j])
	})

	stateMachineJSON := StateMachineJSON{}
	for _, transition := range keys {
		for _, rule := range sm.transitionRules[transition] {
			var sourceStates []string
			if len(rule.SourceStates) == 1 && rule.SourceStates[0] == State("") {
				sourceStates = []string{"initial"}
			} else {
				sourceStates = make([]string, len(rule.SourceStates))
				for i, state := range rule.SourceStates {
					sourceStates[i] = string(state)
				}
			}
			destState := string(rule.DestinationState)

			stateMachineJSON.TransitionRules = append(stateMachineJSON.TransitionRules, TransitionRuleJSON{
				TransitionType:   transition,
				SourceStates:     sourceStates,
				DestinationState: destState,
				Name:             rule.Documentation.Name,
				Description:      rule.Documentation.Description,
			})

			// populate nodes
			for _, sourceState := range sourceStates {
				stateMachineJSON.TransitionRuleEdges = append(stateMachineJSON.TransitionRuleEdges, TransitionRuleEdge{
					From:        string(sourceState),
					To:          string(destState),
					Description: rule.Documentation.Description,
					Name:        string(transition),
				})
			}
		}
	}

	stateMachineJSON.States = make(map[string]StateJSON)
	for stateID, stateDoc := range sm.stateDocs {
		stateMachineJSON.States[string(stateID)] = StateJSON{
			Name:        stateDoc.Name,
			Description: stateDoc.Description,
		}

		// populate edges
		stateMachineJSON.TransitionRuleNodes = append(stateMachineJSON.TransitionRuleNodes, TransitionRuleNode{
			ID:          stateDoc.Name,
			Description: stateDoc.Description,
		})

	}

	stateMachineJSON.TransitionTypes = make(map[string]TransitionTypeJSON)
	for transitionTypeID, transitionTypeDoc := range sm.transitionTypeDocs {
		stateMachineJSON.TransitionTypes[string(transitionTypeID)] = TransitionTypeJSON{
			Name:        transitionTypeDoc.Name,
			Description: transitionTypeDoc.Description,
		}
	}

	return stateMachineJSON
}

func (sm *stateMachine) AsJSON() ([]byte, error) {
	marshaled, err := json.MarshalIndent(sm.Export(), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state machine to JSON: %w", err)
	}

	return marshaled, nil
}

//func (sm *stateMachine) AsJSON() ([]byte, error) {
//	// Generate a sorted list of all states to avoid non-deterministic JSON
//	// output
//	keys := make([]TransitionType, 0, len(sm.transitionRules))
//	for tt := range sm.transitionRules {
//		keys = append(keys, tt)
//	}
//	sort.Slice(keys, func(i, j int) bool {
//		return string(keys[i]) < string(keys[j])
//	})
//
//	stateMachineJSON := StateMachineJSON{}
//	for _, transition := range keys {
//		for _, rule := range sm.transitionRules[transition] {
//			var sourceStates []string
//			if len(rule.SourceStates) == 1 && rule.SourceStates[0] == State("") {
//				sourceStates = []string{"initial"}
//			} else {
//				sourceStates = make([]string, len(rule.SourceStates))
//				for i, state := range rule.SourceStates {
//					sourceStates[i] = string(state)
//				}
//			}
//			destState := string(rule.DestinationState)
//
//			stateMachineJSON.TransitionRules = append(stateMachineJSON.TransitionRules, TransitionRuleJSON{
//				TransitionType:   transition,
//				SourceStates:     sourceStates,
//				DestinationState: destState,
//				Name:             rule.Documentation.Name,
//				Description:      rule.Documentation.Description,
//			})
//		}
//	}
//
//	stateMachineJSON.States = make(map[string]StateJSON)
//	for stateID, stateDoc := range sm.stateDocs {
//		stateMachineJSON.States[string(stateID)] = StateJSON{
//			Name:        stateDoc.Name,
//			Description: stateDoc.Description,
//		}
//	}
//
//	stateMachineJSON.TransitionTypes = make(map[string]TransitionTypeJSON)
//	for transitionTypeID, transitionTypeDoc := range sm.transitionTypeDocs {
//		stateMachineJSON.TransitionTypes[string(transitionTypeID)] = TransitionTypeJSON{
//			Name:        transitionTypeDoc.Name,
//			Description: transitionTypeDoc.Description,
//		}
//	}
//
//	marshaled, err := json.MarshalIndent(stateMachineJSON, "", "  ")
//	if err != nil {
//		return nil, fmt.Errorf("failed to marshal state machine to JSON: %w", err)
//	}
//
//	return marshaled, nil
//}
