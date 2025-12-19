package models

import (
	"encoding/json"
	"testing"
)

func TestStackNode_IsBlocked(t *testing.T) {
	tests := []struct {
		name string
		node *StackNode
		want bool
	}{
		{
			name: "no parent - not blocked",
			node: &StackNode{
				PR:     &PR{Number: 1, State: PRStateOpen},
				Parent: nil,
			},
			want: false,
		},
		{
			name: "parent open - blocked",
			node: &StackNode{
				PR: &PR{Number: 2, State: PRStateOpen},
				Parent: &StackNode{
					PR: &PR{Number: 1, State: PRStateOpen},
				},
			},
			want: true,
		},
		{
			name: "parent merged - not blocked",
			node: &StackNode{
				PR: &PR{Number: 2, State: PRStateOpen},
				Parent: &StackNode{
					PR: &PR{Number: 1, State: PRStateMerged},
				},
			},
			want: false,
		},
		{
			name: "parent draft - blocked",
			node: &StackNode{
				PR: &PR{Number: 2, State: PRStateOpen},
				Parent: &StackNode{
					PR: &PR{Number: 1, State: PRStateDraft},
				},
			},
			want: true,
		},
		{
			name: "parent closed - blocked",
			node: &StackNode{
				PR: &PR{Number: 2, State: PRStateOpen},
				Parent: &StackNode{
					PR: &PR{Number: 1, State: PRStateClosed},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.IsBlocked(); got != tt.want {
				t.Errorf("IsBlocked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStackNode_GetRoot(t *testing.T) {
	// Build a 3-level stack
	root := &StackNode{
		PR:    &PR{Number: 1},
		Depth: 0,
	}
	middle := &StackNode{
		PR:     &PR{Number: 2},
		Parent: root,
		Depth:  1,
	}
	leaf := &StackNode{
		PR:     &PR{Number: 3},
		Parent: middle,
		Depth:  2,
	}
	root.Children = []*StackNode{middle}
	middle.Children = []*StackNode{leaf}

	tests := []struct {
		name     string
		node     *StackNode
		wantRoot *StackNode
	}{
		{
			name:     "root returns itself",
			node:     root,
			wantRoot: root,
		},
		{
			name:     "middle returns root",
			node:     middle,
			wantRoot: root,
		},
		{
			name:     "leaf returns root",
			node:     leaf,
			wantRoot: root,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.GetRoot(); got != tt.wantRoot {
				t.Errorf("GetRoot() = %v, want %v", got.PR.Number, tt.wantRoot.PR.Number)
			}
		})
	}
}

func TestStackNode_HasChildren(t *testing.T) {
	nodeWithChildren := &StackNode{
		Children: []*StackNode{{PR: &PR{Number: 1}}},
	}
	nodeWithoutChildren := &StackNode{
		Children: []*StackNode{},
	}
	nodeNilChildren := &StackNode{}

	if !nodeWithChildren.HasChildren() {
		t.Error("HasChildren() should return true when children exist")
	}
	if nodeWithoutChildren.HasChildren() {
		t.Error("HasChildren() should return false for empty children")
	}
	if nodeNilChildren.HasChildren() {
		t.Error("HasChildren() should return false for nil children")
	}
}

func TestStackNode_IsRoot(t *testing.T) {
	root := &StackNode{PR: &PR{Number: 1}}
	child := &StackNode{PR: &PR{Number: 2}, Parent: root}

	if !root.IsRoot() {
		t.Error("IsRoot() should return true for node without parent")
	}
	if child.IsRoot() {
		t.Error("IsRoot() should return false for node with parent")
	}
}

func TestStack_Size(t *testing.T) {
	emptyStack := &Stack{AllNodes: []*StackNode{}}
	stackWithNodes := &Stack{
		AllNodes: []*StackNode{
			{PR: &PR{Number: 1}},
			{PR: &PR{Number: 2}},
			{PR: &PR{Number: 3}},
		},
	}

	if emptyStack.Size() != 0 {
		t.Errorf("Size() = %d, want 0", emptyStack.Size())
	}
	if stackWithNodes.Size() != 3 {
		t.Errorf("Size() = %d, want 3", stackWithNodes.Size())
	}
}

func TestStack_IsEmpty(t *testing.T) {
	emptyStack := &Stack{AllNodes: []*StackNode{}}
	nonEmptyStack := &Stack{AllNodes: []*StackNode{{PR: &PR{Number: 1}}}}

	if !emptyStack.IsEmpty() {
		t.Error("IsEmpty() should return true for empty stack")
	}
	if nonEmptyStack.IsEmpty() {
		t.Error("IsEmpty() should return false for non-empty stack")
	}
}

func TestStackNode_JSONSerialization(t *testing.T) {
	// Build a simple parent-child relationship
	parent := &StackNode{
		PR:    &PR{Number: 1, Title: "Parent PR"},
		Depth: 0,
	}
	child := &StackNode{
		PR:     &PR{Number: 2, Title: "Child PR"},
		Parent: parent,
		Depth:  1,
	}
	parent.Children = []*StackNode{child}

	// Serialize the child
	data, err := json.Marshal(child)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Verify Parent is NOT in the JSON (would cause infinite loop)
	jsonStr := string(data)
	if containsStr(jsonStr, "Parent PR") {
		t.Error("Parent should not be serialized to JSON")
	}

	// Verify child's own PR data is present
	if !containsStr(jsonStr, "Child PR") {
		t.Error("PR should be serialized")
	}

	// Verify we can unmarshal without errors
	var decoded StackNode
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Depth != child.Depth {
		t.Errorf("Depth = %d, want %d", decoded.Depth, child.Depth)
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
