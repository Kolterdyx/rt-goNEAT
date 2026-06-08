package genetics

// InnovationsObserver manages records of structural innovations shared across the population.
type InnovationsObserver interface {
	// StoreInnovation records a new innovation.
	StoreInnovation(innovation Innovation)
	// Innovations returns all recorded innovations as a slice.
	Innovations() []Innovation
	// NextInnovationNumber returns the next unique global innovation number.
	NextInnovationNumber() int64
	// FindLinkInnovation returns the recorded innovation for adding a link between
	// inNodeId and outNodeId (recurrent or not), or nil if none exists.
	FindLinkInnovation(inNodeId, outNodeId int, isRecurrent bool) *Innovation
	// FindNodeInnovation returns the recorded innovation for splitting the gene with
	// oldInnovNum between inNodeId and outNodeId, or nil if none exists.
	FindNodeInnovation(inNodeId, outNodeId int, oldInnovNum int64) *Innovation
}

// Innovation records a structural change in a genome so that the same change in different
// genomes within the same simulation window can be assigned the same innovation number,
// maintaining alignment for crossover.
type Innovation struct {
	// The two nodes between which the innovation occurred
	InNodeId  int
	OutNodeId int
	// InnovationNum is the primary innovation number assigned to this change
	InnovationNum int64
	// InnovationNum2 holds the second innovation number for node-split innovations (two new links)
	InnovationNum2 int64

	// NewWeight is the weight assigned to a new link innovation
	NewWeight float64
	// NewTraitNum is the trait index for a new link innovation
	NewTraitNum int
	// NewNodeId is the ID of the new node in a node-split innovation
	NewNodeId int

	// OldInnovNum is the innovation number of the gene that was split (node innovations only)
	OldInnovNum int64

	// IsRecurrent flags a recurrent link innovation
	IsRecurrent bool

	// innovationType distinguishes node-split from link-add innovations
	innovationType innovationType
}

// NewInnovationForNode constructs an innovation record for inserting a new node.
func NewInnovationForNode(nodeInId, nodeOutId int, innovationNum1, innovationNum2 int64, newNodeId int, oldInnovNum int64) *Innovation {
	return &Innovation{
		innovationType: newNodeInnType,
		InNodeId:       nodeInId,
		OutNodeId:      nodeOutId,
		InnovationNum:  innovationNum1,
		InnovationNum2: innovationNum2,
		NewNodeId:      newNodeId,
		OldInnovNum:    oldInnovNum,
	}
}

// NewInnovationForLink constructs an innovation record for adding a new (non-recurrent) link.
func NewInnovationForLink(nodeInId, nodeOutId int, innovationNum int64, weight float64, traitId int) *Innovation {
	return &Innovation{
		innovationType: newLinkInnType,
		InNodeId:       nodeInId,
		OutNodeId:      nodeOutId,
		InnovationNum:  innovationNum,
		NewWeight:      weight,
		NewTraitNum:    traitId,
	}
}

// NewInnovationForRecurrentLink constructs an innovation record for adding a (possibly recurrent) link.
func NewInnovationForRecurrentLink(nodeInId, nodeOutId int, innovationNum int64, weight float64, traitId int, recur bool) *Innovation {
	return &Innovation{
		innovationType: newLinkInnType,
		InNodeId:       nodeInId,
		OutNodeId:      nodeOutId,
		InnovationNum:  innovationNum,
		NewWeight:      weight,
		NewTraitNum:    traitId,
		IsRecurrent:    recur,
	}
}
