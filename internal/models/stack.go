package models

// StackNode represents a node in a PR dependency tree (stacked PRs).
// A stack forms when PRs build on each other's branches.
type StackNode struct {
	// The pull request at this node
	PR *PR `json:"pr"`

	// Parent node (excluded from JSON to avoid circular references)
	Parent *StackNode `json:"-"`

	// Child nodes (PRs that depend on this one)
	Children []*StackNode `json:"children,omitempty"`

	// Stack metadata
	Depth    int  `json:"depth"`     // 0 = root
	IsOrphan bool `json:"is_orphan"` // Parent was merged but this PR still targets that branch
}

// Stack represents a collection of related PRs that form a dependency tree.
// A repository may have multiple independent stacks.
type Stack struct {
	// Root nodes (PRs with no parent in the set)
	Roots []*StackNode `json:"roots"`

	// All nodes flattened for easy iteration (excluded from JSON to avoid cycles)
	AllNodes []*StackNode `json:"-"`
}

// IsBlocked returns true if this PR has an unmerged parent PR.
// A blocked PR cannot be merged until its parent is merged first.
func (n *StackNode) IsBlocked() bool {
	if n.Parent == nil {
		return false
	}
	// If parent exists and is not merged, this PR is blocked
	if n.Parent.PR != nil && n.Parent.PR.State != PRStateMerged {
		return true
	}
	return false
}

// GetRoot walks up the tree to find the root StackNode.
// Returns the node itself if it has no parent.
func (n *StackNode) GetRoot() *StackNode {
	current := n
	for current.Parent != nil {
		current = current.Parent
	}
	return current
}

// HasChildren returns true if this node has any child PRs.
func (n *StackNode) HasChildren() bool {
	return len(n.Children) > 0
}

// IsRoot returns true if this node is a root (has no parent).
func (n *StackNode) IsRoot() bool {
	return n.Parent == nil
}

// Size returns the total number of nodes in the stack.
func (s *Stack) Size() int {
	return len(s.AllNodes)
}

// IsEmpty returns true if the stack has no nodes.
func (s *Stack) IsEmpty() bool {
	return len(s.AllNodes) == 0
}
