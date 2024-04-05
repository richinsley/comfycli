package dagger

import (
	"fmt"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/google/uuid"
)

const MaxTopologyCount = 2

type IDaggerBase interface {
	GetInstanceID() string
	GetParent() IDaggerBase
	SetParent(IDaggerBase)
	EmitError(string)
	PurgeAll()
}

type IDaggerBasePin interface {
	IDaggerBase
	GetDirection() PinDirection
	GetPinName() string
	SetPinName(string)
	GetParentNode() IDaggerNode
	SetParentNode(IDaggerNode)
	IsInputPin() bool
	GetAutoCloneRefCount() int
	GetAutoCloneCount() int
	GetMaxAutoClone() int
	GetAutoCloneNameTemplate() string
	GetIndex() int
	GetAutoCloneMaster() IDaggerBasePin
	SetAutoCloneMaster(IDaggerBasePin)
	IsAutoCloned() bool
	CanRename() bool
	SetCanRename(bool)
	GetTopologySystem() int
	IsConnected() bool
	GetOriginalName() string
	CanConnectToPin(IDaggerBasePin) bool
	GetAutoClone() bool
	SetAutoClone(int, string) bool
	IncAutoCloneCount()
	DecAutoCloneCount()
	GenClonedNameFromTemplate()
	Cloned(IDaggerBasePin)
	Clone() IDaggerBasePin
	OnRemoved()
	OnCloned()
}

type IDaggerOutputPin interface {
	IDaggerBasePin
	GetConnectedTo() *[]IDaggerBasePin
	DisconnectAll(bool) bool
	AllowMultiConnect() bool
	SetAllowMultiConnect(bool)
	IsConnected() bool
	GetConnectedToUUIDs() []string
	ConnectToInput(IDaggerInputPin) bool
	DisconnectPin(IDaggerInputPin, bool) bool
}

type IDaggerInputPin interface {
	IDaggerBasePin
	GetConnectedTo() IDaggerOutputPin
	SetConnectedTo(IDaggerOutputPin)
	DisconnectPin(bool) bool
	SetMaxAutoClone(int)
}

// DaggerBase is the base struct for all Dagger objects.
type DaggerBase struct {
	instanceID string
	parent     IDaggerBase
}

// NewDaggerBase creates a new instance of DaggerBase.
func NewDaggerBase() *DaggerBase {
	return &DaggerBase{
		instanceID: uuid.New().String(),
	}
}

// GetInstanceID returns the unique ID for this instance.
func (db *DaggerBase) GetInstanceID() string {
	return db.instanceID
}

// GetParent returns the parent DaggerBase object this object belongs to.
func (db *DaggerBase) GetParent() IDaggerBase {
	return db.parent
}

// SetParent sets the parent DaggerBase object this object belongs to.
func (db *DaggerBase) SetParent(parent IDaggerBase) {
	db.parent = parent
}

// EmitError logs an error.
func (db *DaggerBase) EmitError(err string) {
	fmt.Println("Error:", err)
}

// PurgeAll cleans up when the object hierarchy is unraveled.
func (db *DaggerBase) PurgeAll() {}

// DaggerSignal emulates the Qt signal/slot pattern.
type DaggerSignal struct {
	callbacks []func(args ...interface{})
}

// NewDaggerSignal creates a new instance of DaggerSignal.
func NewDaggerSignal() *DaggerSignal {
	return &DaggerSignal{}
}

// Connect adds a connection callback.
func (ds *DaggerSignal) Connect(slot func(args ...interface{})) {
	ds.callbacks = append(ds.callbacks, slot)
}

// Disconnect removes a connection callback.
func (ds *DaggerSignal) Disconnect(slot func(args ...interface{})) {
	for i, cb := range ds.callbacks {
		if fmt.Sprintf("%p", cb) == fmt.Sprintf("%p", slot) {
			ds.callbacks = append(ds.callbacks[:i], ds.callbacks[i+1:]...)
			break
		}
	}
}

// DisconnectAll removes all connection callbacks.
func (ds *DaggerSignal) DisconnectAll() {
	ds.callbacks = nil
}

// Emit signals all connected callbacks.
func (ds *DaggerSignal) Emit(args ...interface{}) {
	for _, cb := range ds.callbacks {
		cb(args...)
	}
}

// PinDirection represents the direction of a pin.
type PinDirection string

const (
	PinDirectionUnknown PinDirection = "Unknown"
	PinDirectionInput   PinDirection = "Input"
	PinDirectionOutput  PinDirection = "Output"
)

// DaggerBasePin is the base struct for all Dagger pins.
type DaggerBasePin struct {
	DaggerBase
	pinName               string
	parentNode            IDaggerNode
	direction             PinDirection
	nameSet               bool
	canRename             bool
	maxAutoClone          int
	autoCloneCount        int
	autoCloneRef          int
	autoCloneMaster       IDaggerBasePin
	autoCloneNameTemplate string
	originalName          string
	parentNodeChanged     *DaggerSignal
	parentGraphChanged    *DaggerSignal
	pinNameChanged        *DaggerSignal
	canRenameChanged      *DaggerSignal
	pinConnected          *DaggerSignal
	pinDisconnected       *DaggerSignal
}

// NewDaggerBasePin creates a new instance of DaggerBasePin.
func NewDaggerBasePin(direction PinDirection) *DaggerBasePin {
	return &DaggerBasePin{
		DaggerBase:         *NewDaggerBase(),
		maxAutoClone:       0,
		parentNodeChanged:  NewDaggerSignal(),
		parentGraphChanged: NewDaggerSignal(),
		pinNameChanged:     NewDaggerSignal(),
		canRenameChanged:   NewDaggerSignal(),
		pinConnected:       NewDaggerSignal(),
		pinDisconnected:    NewDaggerSignal(),
		direction:          direction,
	}
}

type IDaggerNode interface {
	IDaggerBase
	GetParentGraph() *DaggerGraph
	GetOrdinal(int) int
	SetOrdinal(int, int)
	GetSubgraphAffiliation(int) int
	SetSubgraphAffiliation(int, int)
	GetName() string
	SetName(string)
	GetInputPins(int) *DaggerPinCollection
	GetOutputPins(int) *DaggerPinCollection
	GetDescendents(int) []IDaggerNode
	SetDescendents(int, []IDaggerNode)
	GetAscendents(int) []IDaggerNode
	IsTopLevel(int) bool
	IsBottomLevel(int) bool
	DisconnectAllPins() bool
	GetDaggerOutputPin(string, int) IDaggerOutputPin
	GetDaggerInputPin(string, int) IDaggerInputPin
	IsTrueSource(int) bool
	IsTrueDest(int) bool
	GetCurrentTSystemEval() int
	SetCurrentTSystemEval(int)
	ShouldClonePin(IDaggerBasePin) bool
	ForceCloneWithName(IDaggerBasePin, string) IDaggerBasePin
	IsDescendentOf(IDaggerNode, int) bool
	CanRemovePin(IDaggerBasePin) bool
	ClonePin(IDaggerBasePin, IDaggerBasePin) IDaggerBasePin
	ShouldRemoveClonePin(IDaggerBasePin) bool
	RemoveClonePin(IDaggerBasePin) bool
	SetParentGraph(*DaggerGraph)
	GetFirstUnconnectedInputPin(int) IDaggerInputPin
	GetFirstUnconnectedOutputPin(int) IDaggerOutputPin
	// signals
	BeforeAddedToGraph() *DaggerSignal
	AfterAddedToGraph() *DaggerSignal
	AddedToGraph() *DaggerSignal
}

// GetDirection returns the direction of the pin.
func (dbp *DaggerBasePin) GetDirection() PinDirection {
	return dbp.direction
}

// GetPinName returns the pin's name.
func (dbp *DaggerBasePin) GetPinName() string {
	return dbp.pinName
}

// SetPinName sets the pin's name.
func (dbp *DaggerBasePin) SetPinName(name string) {
	dbp.pinName = name
	if !dbp.nameSet {
		dbp.nameSet = true
		dbp.originalName = name
	}

	if dbp.parentNode != nil {
		dbp.pinNameChanged.Emit()
	}
}

// GetParentNode returns the node this pin belongs to.
func (dbp *DaggerBasePin) GetParentNode() IDaggerNode {
	return dbp.parentNode
}

// SetParentNode sets the node this pin belongs to.
func (dbp *DaggerBasePin) SetParentNode(node IDaggerNode) {
	dbp.parentNode = node
	dbp.parentNodeChanged.Emit()
}

// IsInputPin returns true if this is an input pin.
func (dbp *DaggerBasePin) IsInputPin() bool {
	direction := dbp.GetDirection()
	return direction == PinDirectionInput
}

// GetAutoCloneRefCount returns the total number of times the pin was auto-cloned.
func (dbp *DaggerBasePin) GetAutoCloneRefCount() int {
	return dbp.autoCloneRef
}

// SetAutoCloneMaster sets the pin that this pin will clone from when connected.
func (dbp *DaggerBasePin) SetAutoCloneMaster(master IDaggerBasePin) {
	dbp.autoCloneMaster = master
}

// GetAutoCloneCount returns the number of times this pin was auto-cloned.
func (dbp *DaggerBasePin) GetAutoCloneCount() int {
	return dbp.autoCloneCount
}

// GetMaxAutoClone returns the maximum number of times this pin can be auto-cloned.
func (dbp *DaggerBasePin) GetMaxAutoClone() int {
	return dbp.maxAutoClone
}

// GetAutoCloneNameTemplate returns the template used for creating a unique name for cloned pins.
func (dbp *DaggerBasePin) GetAutoCloneNameTemplate() string {
	return dbp.autoCloneNameTemplate
}

// GetIndex returns the index of this pin within its pin collection.
func (dbp *DaggerBasePin) GetIndex() int {
	parentCollection := dbp.GetParent().(*DaggerPinCollection)
	return parentCollection.GetIndex(dbp)
}

// GetAutoCloneMaster returns the pin that this pin will clone from when connected.
func (dbp *DaggerBasePin) GetAutoCloneMaster() IDaggerBasePin {
	return dbp.autoCloneMaster
}

// IsAutoCloned returns true if this pin is an auto-cloned pin.
func (dbp *DaggerBasePin) IsAutoCloned() bool {
	return dbp.autoCloneMaster != nil && dbp.autoCloneMaster != dbp
}

// CanRename returns true if this pin can be renamed.
func (dbp *DaggerBasePin) CanRename() bool {
	return dbp.canRename
}

// SetCanRename sets whether the pin can be renamed.
func (dbp *DaggerBasePin) SetCanRename(val bool) {
	dbp.canRename = val
	if dbp.parentNode != nil {
		dbp.canRenameChanged.Emit()
	}
}

// GetTopologySystem returns the topology system this pin belongs to.
func (dbp *DaggerBasePin) GetTopologySystem() int {
	return dbp.GetParent().(*DaggerPinCollection).GetTopologySystem()
}

// IsConnected returns true if this pin has a connection to another pin.
func (dbp *DaggerBasePin) IsConnected() bool {
	return false
}

// GetOriginalName returns the original name of the pin if it was renamed.
func (dbp *DaggerBasePin) GetOriginalName() string {
	return dbp.originalName
}

// CanConnectToPin returns true if this pin can connect to the given pin.
func (dbp *DaggerBasePin) CanConnectToPin(pin IDaggerBasePin) bool {
	return dbp.GetDirection() != pin.GetDirection()
}

// GetAutoClone returns true if this pin auto-clones when connected.
func (dbp *DaggerBasePin) GetAutoClone() bool {
	var idp IDaggerBasePin = dbp
	return dbp.autoCloneMaster == idp
}

// SetAutoClone sets whether this pin auto-clones when connected.
func (dbp *DaggerBasePin) SetAutoClone(maxAutoCloneCount int, autoCloneNameTemplate string) bool {
	dbp.maxAutoClone = maxAutoCloneCount
	dbp.autoCloneNameTemplate = autoCloneNameTemplate
	dbp.autoCloneMaster = dbp
	return true
}

// IncAutoCloneCount increments the auto-clone count.
func (dbp *DaggerBasePin) IncAutoCloneCount() {
	dbp.autoCloneCount++
	dbp.autoCloneRef++
}

// DecAutoCloneCount decrements the auto-clone count.
func (dbp *DaggerBasePin) DecAutoCloneCount() {
	dbp.autoCloneCount--
}

// GenClonedNameFromTemplate generates a cloned name from the template.
func (dbp *DaggerBasePin) GenClonedNameFromTemplate() {
	rcount := fmt.Sprintf("%d", dbp.autoCloneMaster.GetAutoCloneRefCount())
	newName := strings.ReplaceAll(dbp.autoCloneMaster.GetAutoCloneNameTemplate(), "%", rcount)
	dbp.SetPinName(newName)
}

// Cloned is called after a pin was cloned.
func (dbp *DaggerBasePin) Cloned(fromMaster IDaggerBasePin) {
	dbp.autoCloneMaster = fromMaster
	dbp.autoCloneMaster.IncAutoCloneCount()
	dbp.canRename = fromMaster.CanRename()

	dbp.GenClonedNameFromTemplate()
}

// Clone creates a new instance of the pin's subclass.
func (dbp *DaggerBasePin) Clone() IDaggerBasePin {
	if dbp.autoCloneMaster == nil {
		return nil
	}

	if dbp.autoCloneMaster.GetDirection() == PinDirectionInput {
		npin := NewDaggerInputPin()
		// copier.Copy(npin, dbp)
		return npin
	}

	return nil
}

// OnRemoved is called by the pin collection when the pin is removed.
func (dbp *DaggerBasePin) OnRemoved() {}

// OnCloned is called by a parent node after this pin was cloned from a master.
func (dbp *DaggerBasePin) OnCloned() {}

// PurgeAll cleans up when the object hierarchy is unraveled.
func (dbp *DaggerBasePin) PurgeAll() {
	dbp.DaggerBase.PurgeAll()
}

// DaggerInputPin represents a directed input into a DaggerNode.
type DaggerInputPin struct {
	DaggerBasePin
	connectedTo *DaggerOutputPin
}

// NewDaggerInputPin creates a new instance of DaggerInputPin.
func NewDaggerInputPin() *DaggerInputPin {
	return &DaggerInputPin{
		DaggerBasePin: *NewDaggerBasePin(PinDirectionInput),
	}
}

// GetDirection returns the pin's direction.
func (dip *DaggerInputPin) GetDirection() PinDirection {
	return PinDirectionInput
}

// GetConnectedTo returns the DaggerOutputPin this pin is connected to.
func (dip *DaggerInputPin) GetConnectedTo() IDaggerOutputPin {
	return dip.connectedTo
}

func (dip *DaggerInputPin) SetConnectedTo(pin IDaggerOutputPin) {
	if pin == nil {
		dip.connectedTo = nil
		return
	}
	opin := pin.(*DaggerOutputPin)
	dip.connectedTo = opin
}

// GetConnectedToUUID returns the instance ID of the DaggerOutputPin this pin is connected to.
func (dip *DaggerInputPin) GetConnectedToUUID() string {
	if dip.IsConnected() {
		return dip.connectedTo.GetInstanceID()
	}
	return "00000000-0000-0000-0000-000000000000"
}

// IsConnected returns true if this pin is connected.
func (dip *DaggerInputPin) IsConnected() bool {
	return dip.connectedTo != nil
}

// CanConnectToPin returns true if this pin can connect to the given DaggerOutputPin.
func (dip *DaggerInputPin) CanConnectToPin(pin IDaggerBasePin) bool {
	tsystem := dip.GetTopologySystem()
	if tsystem != pin.GetTopologySystem() {
		dip.EmitError("pins must belong to the same topology system")
		return false
	}

	if dip.parentNode == nil {
		return dip.DaggerBasePin.CanConnectToPin(pin)
	}

	if dip.parentNode == pin.GetParentNode() {
		dip.EmitError("pins belong to the same parent node")
		return false
	}

	retv := false
	if dip.parentNode.GetParentGraph().GetEnableTopology() {
		if !dip.parentNode.IsDescendentOf(pin.GetParentNode(), tsystem) {
			retv = dip.DaggerBasePin.CanConnectToPin(pin)
		} else if pin.GetParentNode().GetOrdinal(tsystem) <= dip.parentNode.GetOrdinal(tsystem) {
			retv = dip.DaggerBasePin.CanConnectToPin(pin)
		}
	} else {
		retv = true
	}
	return retv
}

// DisconnectPin disconnects this pin.
func (dip *DaggerInputPin) DisconnectPin(forceDisconnect bool) bool {
	if dip.parentNode == nil {
		return false
	}

	if dip.IsConnected() {
		return dip.connectedTo.DisconnectPin(dip, forceDisconnect)
	}

	return true
}

// SetMaxAutoClone sets the maximum number of times this pin can be auto-cloned.
func (dip *DaggerInputPin) SetMaxAutoClone(maxAutoCloneCount int) {
	dip.maxAutoClone = maxAutoCloneCount
}

// PurgeAll cleans up when the object hierarchy is unraveled.
func (dip *DaggerInputPin) PurgeAll() {
	dip.DaggerBasePin.PurgeAll()
	dip.connectedTo = nil
}

// DaggerOutputPin represents a directed output out of a DaggerNode.
type DaggerOutputPin struct {
	DaggerBasePin
	connectedTo       []IDaggerBasePin
	allowMultiConnect bool
}

// NewDaggerOutputPin creates a new instance of DaggerOutputPin.
func NewDaggerOutputPin() *DaggerOutputPin {
	return &DaggerOutputPin{
		DaggerBasePin:     *NewDaggerBasePin(PinDirectionOutput),
		allowMultiConnect: true,
	}
}

// GetDirection returns the pin's direction.
func (dop *DaggerOutputPin) GetDirection() PinDirection {
	return PinDirectionOutput
}

// GetConnectedTo returns a list of all DaggerInputPins this pin is connected to.
func (dop *DaggerOutputPin) GetConnectedTo() *[]IDaggerBasePin {
	return &dop.connectedTo
}

// AllowMultiConnect returns true if this output pin allows for multiple connections.
func (dop *DaggerOutputPin) AllowMultiConnect() bool {
	return dop.allowMultiConnect
}

// SetAllowMultiConnect sets whether this output pin allows for multiple connections.
func (dop *DaggerOutputPin) SetAllowMultiConnect(val bool) {
	dop.allowMultiConnect = val
}

// IsConnected returns true if this pin has any connections.
func (dop *DaggerOutputPin) IsConnected() bool {
	return len(dop.connectedTo) != 0
}

// GetConnectedToUUIDs returns a list of instance IDs for each DaggerInputPin this pin is connected to.
func (dop *DaggerOutputPin) GetConnectedToUUIDs() []string {
	var uuids []string
	for _, pin := range dop.connectedTo {
		uuids = append(uuids, pin.GetInstanceID())
	}
	return uuids
}

// ConnectToInput connects this pin to the given DaggerInputPin.
func (dop *DaggerOutputPin) ConnectToInput(input IDaggerInputPin) bool {
	if input == nil {
		dop.EmitError("Input pin was null in ConnectToInput")
		return false
	}

	outputPinNode := dop.GetParentNode()
	if outputPinNode == nil {
		dop.EmitError("Output pin is not associated with a DaggerNode")
		return false
	}
	outputPinContainer := outputPinNode.GetParentGraph()

	inputPinNode := input.GetParentNode()
	if inputPinNode == nil {
		dop.EmitError("Input pin is not associated with a DaggerNode")
		return false
	}
	inputPinContainer := input.GetParentNode().GetParentGraph()

	if outputPinContainer == nil {
		dop.EmitError("Output pin is not associated with a DaggerNode or DaggerGraph")
		return false
	}

	if inputPinContainer == nil {
		dop.EmitError("Input pin is not associated with a DaggerNode or DaggerGraph")
		return false
	}

	if inputPinContainer != outputPinContainer {
		dop.EmitError("Input pin and Output pin are not associated with the same DaggerGraph")
		return false
	}

	if outputPinContainer.GetEnableTopology() {
		if !input.CanConnectToPin(dop) {
			dop.EmitError("Input pin indicates it cannot connect to this output pin")
			return false
		}

		if !dop.CanConnectToPin(input) {
			dop.EmitError("Parent node indicates Input pin cannot connect to this output pin")
			return false
		}
	}

	if input.IsConnected() {
		if input.GetAutoCloneMaster() != nil {
			dop.EmitError("cannot swap connections on cloned pins")
			return false
		}

		if !input.DisconnectPin(false) {
			dop.EmitError("Input pin is already connected and was not allowed to disconnect")
			return false
		}
	}

	if outputPinContainer.BeforePinsConnected(dop, input) {
		dop.connectedTo = append(dop.connectedTo, input)
		input.SetConnectedTo(dop)

		outputPinContainer.OnPinsConnected(dop, input)

		// dop.pinConnected.Emit(input)
		// input.pinConnected.Emit(dop)

		outputPinContainer.AfterPinsConnected(dop, input)

		return true
	}

	return false
}

// CanConnectToPin returns true if this pin can connect to the given DaggerInputPin.
func (dop *DaggerOutputPin) CanConnectToPin(pin IDaggerBasePin) bool {
	if pin == nil || pin.GetDirection() != PinDirectionInput {
		return false
	}

	mtop := dop.GetTopologySystem()
	ttop := pin.GetTopologySystem()
	if mtop != ttop {
		return false
	}

	if pin.IsConnected() {
		return false
	}

	if !dop.allowMultiConnect && dop.IsConnected() {
		return false
	}

	if !pin.GetParentNode().IsDescendentOf(dop.parentNode, mtop) {
		return dop.DaggerBasePin.CanConnectToPin(pin)
	} else if pin.GetParentNode().GetOrdinal(mtop) >= dop.parentNode.GetOrdinal(mtop) {
		return dop.DaggerBasePin.CanConnectToPin(pin)
	}

	return false
}

// DisconnectPin disconnects this pin from the given DaggerInputPin.
func (dop *DaggerOutputPin) DisconnectPin(input IDaggerInputPin, forceDisconnect bool) bool {
	parentGraph := dop.parentNode.GetParentGraph()
	if parentGraph == nil {
		return false
	}

	if containsPin(dop.connectedTo, input) {
		if forceDisconnect || parentGraph.BeforePinsDisconnected(dop, input) {
			removePin(&dop.connectedTo, input)
			input.SetConnectedTo(nil)

			parentGraph.OnPinsDisconnected(dop, input)

			dop.pinDisconnected.Emit(input)
			// input.pinDisconnected.Emit(dop)

			parentGraph.AfterPinsDisconnected(dop, input)
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}

// DisconnectAll disconnects this pin from all DaggerInputPins.
func (dop *DaggerOutputPin) DisconnectAll(forceDisconnect bool) bool {
	ccount := len(dop.connectedTo)
	for i := ccount - 1; i >= 0; i-- {
		pin := dop.connectedTo[i]
		ipin, _ := pin.(*DaggerInputPin)
		if !ipin.DisconnectPin(forceDisconnect) {
			return false
		}
	}
	return true
}

// PurgeAll cleans up when the object hierarchy is unraveled.
func (dop *DaggerOutputPin) PurgeAll() {
	dop.DaggerBasePin.PurgeAll()
	dop.connectedTo = nil
}

// DaggerPinCollection acts as a container for pins of one particular direction.
type DaggerPinCollection struct {
	DaggerBase
	direction      PinDirection
	parentNode     IDaggerNode
	topologySystem int
	pinCollection  map[string]IDaggerBasePin
	orderedPins    []IDaggerBasePin
	pinRemoved     *DaggerSignal
	pinAdded       *DaggerSignal
}

// NewDaggerPinCollection creates a new instance of DaggerPinCollection.
func NewDaggerPinCollection(parentNode IDaggerNode, direction PinDirection, topologySystem int) *DaggerPinCollection {
	return &DaggerPinCollection{
		DaggerBase:     *NewDaggerBase(),
		direction:      direction,
		parentNode:     parentNode,
		topologySystem: topologySystem,
		pinCollection:  make(map[string]IDaggerBasePin),
		pinRemoved:     NewDaggerSignal(),
		pinAdded:       NewDaggerSignal(),
		orderedPins:    make([]IDaggerBasePin, 0),
	}
}

// GetTopologySystem returns the topology system this pin collection belongs to.
func (dpc *DaggerPinCollection) GetTopologySystem() int {
	return dpc.topologySystem
}

// GetPin finds and returns a pin with the given name.
func (dpc *DaggerPinCollection) GetPin(withName string) IDaggerBasePin {
	return dpc.pinCollection[withName]
}

// AddPin adds a given pin to this collection.
func (dpc *DaggerPinCollection) AddPin(pin IDaggerBasePin, name string) bool {
	if pin == nil {
		return false
	}

	pin.SetParent(dpc)

	if name != "" {
		pin.SetPinName(name)
	} else if pin.GetPinName() != "" {
		pin.SetPinName(pin.GetInstanceID())
	}

	if _, exists := dpc.pinCollection[pin.GetPinName()]; exists {
		nn := strings.ReplaceAll(pin.GetPinName(), "0123456789", "")
		cc := 0
		for {
			an := fmt.Sprintf("%s%d", nn, cc)
			if _, exists := dpc.pinCollection[an]; !exists {
				break
			}
			cc++
		}
		pin.SetPinName(fmt.Sprintf("%s%d", nn, cc))
	}

	pin.SetParentNode(dpc.parentNode)

	dpc.pinCollection[pin.GetPinName()] = pin
	dpc.orderedPins = append(dpc.orderedPins, pin)

	dpc.pinAdded.Emit(pin)

	return true
}

// SetPinName sets the name for a given pin.
func (dpc *DaggerPinCollection) SetPinName(pin IDaggerBasePin, name string) bool {
	if pin.GetPinName() == name {
		return true
	}

	if dpc.GetPin(name) != nil {
		return false
	}

	delete(dpc.pinCollection, pin.GetPinName())
	dpc.pinCollection[name] = pin

	pin.SetPinName(name)

	return true
}

// RemovePin removes a given pin from this collection.
func (dpc *DaggerPinCollection) RemovePin(pin IDaggerBasePin) bool {
	if containsPin(dpc.orderedPins, pin) && !pin.IsConnected() {
		if dpc.parentNode != nil && !dpc.parentNode.CanRemovePin(pin) {
			return false
		}

		delete(dpc.pinCollection, pin.GetPinName())
		removePin(&dpc.orderedPins, pin)
	}
	return false
}

// GetIndex returns the index of the given pin in the collection.
func (dpc *DaggerPinCollection) GetIndex(pin IDaggerBasePin) int {
	for i, p := range dpc.orderedPins {
		if p == pin {
			return i
		}
	}
	return -1
}

// GetParentNode returns the parent node (same result as DaggerBase.GetParent).
func (dpc *DaggerPinCollection) GetParentNode() IDaggerNode {
	return dpc.parentNode
}

// GetPinDirection returns the direction of pins in this collection.
func (dpc *DaggerPinCollection) GetPinDirection() PinDirection {
	return dpc.direction
}

// GetAllPins returns a list of all the pins in the collection.
func (dpc *DaggerPinCollection) GetAllPins() []IDaggerBasePin {
	return dpc.orderedPins
}

// GetAllNonConnectedPins returns a list of all pins in the collection that are not connected.
func (dpc *DaggerPinCollection) GetAllNonConnectedPins() []IDaggerBasePin {
	var nonConnectedPins []IDaggerBasePin
	for _, pin := range dpc.orderedPins {
		if !pin.IsConnected() {
			nonConnectedPins = append(nonConnectedPins, pin)
		}
	}
	return nonConnectedPins
}

// GetAllConnectedPins returns a list of all pins in the collection that are connected.
func (dpc *DaggerPinCollection) GetAllConnectedPins() []IDaggerBasePin {
	var connectedPins []IDaggerBasePin
	for _, pin := range dpc.orderedPins {
		if pin.IsConnected() {
			connectedPins = append(connectedPins, pin)
		}
	}
	return connectedPins
}

// PurgeAll cleans up when the object hierarchy is unraveled.
func (dpc *DaggerPinCollection) PurgeAll() {
	for _, pin := range dpc.orderedPins {
		pin.PurgeAll()
	}
	dpc.pinCollection = nil
	dpc.orderedPins = nil

	dpc.DaggerBase.PurgeAll()
}

// GetFirstUnconnectedPin returns the first unconnected pin or nil if no pins are unconnected.
func (dpc *DaggerPinCollection) GetFirstUnconnectedPin() IDaggerBasePin {
	for _, p := range dpc.orderedPins {
		if !p.IsConnected() {
			return p
		}
	}
	return nil
}

// Helper functions

// containsPin checks if a pin is present in a slice of pins.
func containsPin(pins []IDaggerBasePin, pin IDaggerBasePin) bool {
	for _, p := range pins {
		if p == pin {
			return true
		}
	}
	return false
}

// removePin removes a pin from a slice of pins.
func removePin(pins *[]IDaggerBasePin, pin IDaggerBasePin) {
	for i, p := range *pins {
		if p == pin {
			// We found the pin to remove.
			copy((*pins)[i:], (*pins)[i+1:]) // Shift elements to the left.
			*pins = (*pins)[:len(*pins)-1]   // Slice off the last element.
			break                            // Assuming only one instance should be removed.
		}
	}
}

// DaggerNode represents a node (vertex) in the graph.
type DaggerNode struct {
	DaggerBase
	currentTSystemEval  int
	name                string
	parentGraph         *DaggerGraph
	descendents         [][]IDaggerNode
	subgraphAffiliation []int
	ordinal             []int
	outputPins          []*DaggerPinCollection
	inputPins           []*DaggerPinCollection
	afterAddedToGraph   *DaggerSignal
	beforeAddedToGraph  *DaggerSignal
	addedToGraph        *DaggerSignal
	pinCloned           *DaggerSignal
	nameChanged         *DaggerSignal
}

// NewDaggerNode creates a new instance of DaggerNode.
func NewDaggerNode() *DaggerNode {
	node := &DaggerNode{
		DaggerBase:          *NewDaggerBase(),
		currentTSystemEval:  -1,
		name:                "DaggerNode",
		descendents:         make([][]IDaggerNode, MaxTopologyCount),
		subgraphAffiliation: make([]int, MaxTopologyCount),
		ordinal:             make([]int, MaxTopologyCount),
		outputPins:          make([]*DaggerPinCollection, MaxTopologyCount),
		inputPins:           make([]*DaggerPinCollection, MaxTopologyCount),
		afterAddedToGraph:   NewDaggerSignal(),
		beforeAddedToGraph:  NewDaggerSignal(),
		addedToGraph:        NewDaggerSignal(),
		pinCloned:           NewDaggerSignal(),
		nameChanged:         NewDaggerSignal(),
	}

	for i := 0; i < MaxTopologyCount; i++ {
		node.subgraphAffiliation[i] = -1
		node.ordinal[i] = -1
		node.inputPins[i] = NewDaggerPinCollection(node, PinDirectionInput, i)
		node.outputPins[i] = NewDaggerPinCollection(node, PinDirectionOutput, i)
	}

	return node
}

func (dn *DaggerNode) BeforeAddedToGraph() *DaggerSignal {
	return dn.beforeAddedToGraph
}

func (dn *DaggerNode) AfterAddedToGraph() *DaggerSignal {
	return dn.afterAddedToGraph
}

func (dn *DaggerNode) AddedToGraph() *DaggerSignal {
	return dn.addedToGraph
}

// GetParentGraph returns the graph this node belongs to.
func (dn *DaggerNode) GetParentGraph() *DaggerGraph {
	return dn.parentGraph
}

// SetParentGraph sets the graph this node belongs to.
func (dn *DaggerNode) SetParentGraph(graph *DaggerGraph) {
	dn.parentGraph = graph
}

// GetFirstUnconnectedInputPin returns the first unconnected input pin or nil if no pins are unconnected.
func (dn *DaggerNode) GetFirstUnconnectedInputPin(topologySystem int) IDaggerInputPin {
	return dn.inputPins[topologySystem].GetFirstUnconnectedPin().(IDaggerInputPin)
}

// GetFirstUnconnectedOutputPin returns the first unconnected input pin or nil if no pins are unconnected.
func (dn *DaggerNode) GetFirstUnconnectedOutputPin(topologySystem int) IDaggerOutputPin {
	return dn.outputPins[topologySystem].GetFirstUnconnectedPin().(IDaggerOutputPin)
}

// GetOrdinal returns the order of causality for the given topology that this node represents.
func (dn *DaggerNode) GetOrdinal(topologySystem int) int {
	return dn.ordinal[topologySystem]
}

// SetOrdinal sets the order of causality for the given topology that this node represents.
func (dn *DaggerNode) SetOrdinal(topologySystem int, ord int) {
	dn.ordinal[topologySystem] = ord
}

// GetSubgraphAffiliation returns the index of the subgraph this node belongs to.
func (dn *DaggerNode) GetSubgraphAffiliation(topologySystem int) int {
	return dn.subgraphAffiliation[topologySystem]
}

// SetSubgraphAffiliation sets the index of the subgraph this node belongs to.
func (dn *DaggerNode) SetSubgraphAffiliation(topologySystem int, index int) {
	dn.subgraphAffiliation[topologySystem] = index
}

// GetName returns the name for this node.
func (dn *DaggerNode) GetName() string {
	return dn.name
}

// SetName sets the name for this node.
func (dn *DaggerNode) SetName(newName string) {
	dn.name = newName
	dn.nameChanged.Emit(newName)
}

// GetInputPins returns the DaggerPinCollection for the input pins of the given topology.
func (dn *DaggerNode) GetInputPins(topologySystem int) *DaggerPinCollection {
	return dn.inputPins[topologySystem]
}

// GetOutputPins returns the DaggerPinCollection for the output pins of the given topology.
func (dn *DaggerNode) GetOutputPins(topologySystem int) *DaggerPinCollection {
	return dn.outputPins[topologySystem]
}

// GetDescendents returns the list of nodes that the 'cause' of this node will 'effect'.
func (dn *DaggerNode) GetDescendents(topologySystem int) []IDaggerNode {
	return dn.descendents[topologySystem]
}

// SetDescendents sets the list of nodes that the 'cause' of this node will 'effect'.
func (dn *DaggerNode) SetDescendents(topologySystem int, desc []IDaggerNode) {
	dn.descendents[topologySystem] = desc
}

// GetAscendents returns the list of nodes that 'cause' this node.
func (dn *DaggerNode) GetAscendents(topologySystem int) []IDaggerNode {
	var retv []IDaggerNode
	if dn.parentGraph != nil {
		all := dn.parentGraph.GetNodes()
		for _, node := range all {
			if node != dn && containsNode(node.GetDescendents(topologySystem), dn) {
				retv = append(retv, node)
			}
		}
	}
	return retv
}

// IsTopLevel returns true if this node has no connected input pins for the given topology.
func (dn *DaggerNode) IsTopLevel(topologySystem int) bool {
	allPins := dn.inputPins[topologySystem].GetAllPins()
	for _, pin := range allPins {
		if pin.IsConnected() {
			ipin, _ := pin.(*DaggerInputPin)
			if ipin.GetConnectedTo().GetParentNode() != nil {
				return false
			}
		}
	}
	return true
}

// IsBottomLevel returns true if this node has no connected output pins for the given topology.
func (dn *DaggerNode) IsBottomLevel(topologySystem int) bool {
	allPins := dn.outputPins[topologySystem].GetAllPins()
	for _, pin := range allPins {
		if pin.IsConnected() {
			return false
		}
	}
	return true
}

// DisconnectAllPins disconnects all this node's pins on all of its topologies.
func (dn *DaggerNode) DisconnectAllPins() bool {
	for i := 0; i < dn.parentGraph.GetTopologyCount(); i++ {
		allOutput := dn.outputPins[i].GetAllPins()
		for j := len(allOutput) - 1; j >= 0; j-- {
			opin, _ := allOutput[j].(IDaggerOutputPin)
			if !opin.DisconnectAll(false) {
				return false
			}
		}

		allInput := dn.inputPins[i].GetAllPins()
		for j := len(allInput) - 1; j >= 0; j-- {
			ipin, _ := allInput[j].(*DaggerInputPin)
			if !ipin.DisconnectPin(false) {
				return false
			}
		}
	}

	return true
}

// GetDaggerOutputPin finds and returns the DaggerOutputPin with the given name.
func (dn *DaggerNode) GetDaggerOutputPin(withName string, topologySystem int) IDaggerOutputPin {
	opin := dn.outputPins[topologySystem].GetPin(withName)
	if opin != nil {
		return opin.(IDaggerOutputPin)
	}
	return nil
}

// GetDaggerInputPin finds and returns the DaggerInputPin with the given name.
func (dn *DaggerNode) GetDaggerInputPin(withName string, topologySystem int) IDaggerInputPin {
	return dn.inputPins[topologySystem].GetPin(withName).(IDaggerInputPin)
}

// IsTrueSource returns true if this node has no input pins.
func (dn *DaggerNode) IsTrueSource(topologySystem int) bool {
	return len(dn.inputPins[topologySystem].GetAllPins()) == 0
}

// IsTrueDest returns true if this node has no output pins.
func (dn *DaggerNode) IsTrueDest(topologySystem int) bool {
	return len(dn.outputPins[topologySystem].GetAllPins()) == 0
}

// GetCurrentTSystemEval returns the current topology system that is being evaluated in CalculateTopology.
func (dn *DaggerNode) GetCurrentTSystemEval() int {
	return dn.currentTSystemEval
}

// SetCurrentTSystemEval sets the current topology system that is being evaluated in CalculateTopology.
func (dn *DaggerNode) SetCurrentTSystemEval(system int) {
	dn.currentTSystemEval = system
}

// ShouldClonePin is called by DaggerGraph after pins are connected to see if a pin should be cloned.
func (dn *DaggerNode) ShouldClonePin(pin IDaggerBasePin) bool {
	if pin.GetAutoCloneMaster() != nil {
		if pin.IsInputPin() {
			toMax := pin.GetAutoCloneMaster().GetMaxAutoClone()
			if toMax != 0 {
				if toMax == -1 || pin.GetAutoCloneMaster().GetAutoCloneCount() < toMax {
					return true
				}
			}
		} else {
			opin, _ := pin.(*DaggerOutputPin)
			if len(*opin.GetConnectedTo()) == 1 {
				toMax := pin.GetAutoCloneMaster().GetMaxAutoClone()
				if toMax != 0 {
					if toMax == -1 || pin.GetAutoCloneMaster().GetAutoCloneCount() < toMax {
						return true
					}
				}
			}
		}
	}
	return false
}

// ForceCloneWithName should only be used by deserialization - calling directly can result in duplicate pin names.
func (dn *DaggerNode) ForceCloneWithName(pin IDaggerBasePin, pinName string) IDaggerBasePin {
	retv := dn.ClonePin(pin, nil)
	if retv != nil {
		parentCollection := pin.GetParent().(*DaggerPinCollection)
		if !parentCollection.SetPinName(retv, pinName) {
			dn.RemoveClonePin(pin)
			retv = nil
		}
	}
	return retv
}

// RenamePin renames a given pin. Fails if the pin is not flagged with CanRename. Pins should rarely be renamed.
func (dn *DaggerNode) RenamePin(pin IDaggerBasePin, pinName string) bool {
	if !pin.CanRename() {
		return false
	}

	parentCollection := pin.GetParent().(*DaggerPinCollection)
	return parentCollection.SetPinName(pin, pinName)
}

// ClonePin is called by DaggerGraph after pins are connected to clone a pin.
// If forceAutoCloneMaster is not nil, the pin will be cloned from forceAutoCloneMaster instead of its own autoCloneMaster property.
func (dn *DaggerNode) ClonePin(pin IDaggerBasePin, forceAutoCloneMaster IDaggerBasePin) IDaggerBasePin {
	if pin.IsInputPin() {
		input := forceAutoCloneMaster
		if input == nil {
			input = pin.GetAutoCloneMaster()
		}
		if input == nil {
			return nil
		}

		clonedInput := input.Clone()
		if clonedInput != nil {
			clonedInput.Cloned(input)
			parentCollection := pin.GetParent().(*DaggerPinCollection)
			if parentCollection.AddPin(clonedInput, "") {
				dn.pinCloned.Emit(clonedInput)
				clonedInput.OnCloned()
				return clonedInput
			}
		}
	} else {
		output := forceAutoCloneMaster
		if output == nil {
			output = pin.GetAutoCloneMaster()
		}
		if output == nil {
			return nil
		}

		clonedOutput := output.Clone()
		if clonedOutput != nil {
			clonedOutput.Cloned(output)
			parentCollection := pin.GetParent().(*DaggerPinCollection)
			if parentCollection.AddPin(clonedOutput, "") {
				dn.pinCloned.Emit(clonedOutput)
				clonedOutput.OnCloned()
				return clonedOutput
			}
		}
	}

	return nil
}

// ShouldRemoveClonePin is called by DaggerGraph after pins are disconnected to determine if an autocloned pin should be removed.
func (dn *DaggerNode) ShouldRemoveClonePin(pin IDaggerBasePin) bool {
	if pin.GetAutoCloneMaster() != nil {
		return !pin.IsConnected()
	}
	return false
}

// RemoveClonePin is called by DaggerGraph to remove a cloned pin (the pin that is removed might not be the one that is requested).
func (dn *DaggerNode) RemoveClonePin(pin IDaggerBasePin) bool {
	parentCollection := pin.GetParent().(*DaggerPinCollection)
	if pin.GetAutoCloneMaster() != pin {
		return parentCollection.RemovePin(pin)
	} else {
		all := parentCollection.GetAllNonConnectedPins()
		for _, tpin := range all {
			if tpin != pin && tpin.GetAutoCloneMaster() == pin.GetAutoCloneMaster() {
				return parentCollection.RemovePin(tpin)
			}
		}
	}

	return false
}

// CanRemovePin is overridden to allow a node to decide if a pin can actually be removed. Also useful to detect that a pin is about to be removed.
func (dn *DaggerNode) CanRemovePin(pin IDaggerBasePin) bool {
	return true
}

// IsDescendentOf returns true if the given node is a descendent of this node in the specified topology system.
func (dn *DaggerNode) IsDescendentOf(node IDaggerNode, topologySystem int) bool {
	for _, desc := range dn.descendents[topologySystem] {
		if desc == node {
			return true
		}
	}
	return false
}

// PurgeAll cleans up when the object hierarchy is unraveled. Subclasses should always call super.PurgeAll().
func (dn *DaggerNode) PurgeAll() {
	dn.DaggerBase.PurgeAll()
	for i := 0; i < MaxTopologyCount; i++ {
		if dn.inputPins[i] != nil {
			dn.inputPins[i].PurgeAll()
		}

		if dn.outputPins[i] != nil {
			dn.outputPins[i].PurgeAll()
		}

		dn.descendents[i] = nil
	}

	dn.outputPins = nil
	dn.inputPins = nil
	dn.descendents = nil
}

// DaggerGraph represents a collection of DaggerNodes interconnected via DaggerBasePins.
type DaggerGraph struct {
	DaggerBase
	nodes            []IDaggerNode
	subGraphCount    []int
	maxOrdinal       []int
	topologyCount    int
	pinsDisconnected *DaggerSignal
	pinsConnected    *DaggerSignal
	nodeRemoved      *DaggerSignal
	nodeAdded        *DaggerSignal
	topologyChanged  *DaggerSignal
	topologyEnabled  bool
}

// NewDaggerGraph creates a new instance of DaggerGraph.
func NewDaggerGraph(topologyCount int) *DaggerGraph {
	if topologyCount == 0 {
		topologyCount = 1
	}

	graph := &DaggerGraph{
		DaggerBase:       *NewDaggerBase(),
		subGraphCount:    make([]int, MaxTopologyCount),
		maxOrdinal:       make([]int, MaxTopologyCount),
		topologyCount:    topologyCount,
		pinsDisconnected: NewDaggerSignal(),
		pinsConnected:    NewDaggerSignal(),
		nodeRemoved:      NewDaggerSignal(),
		nodeAdded:        NewDaggerSignal(),
		topologyChanged:  NewDaggerSignal(),
		topologyEnabled:  true,
	}

	graph.CalculateTopology()

	return graph
}

func (dg *DaggerGraph) CalculateTopology() {
	dg.CalculateTopologyDepthFirstSearch()
}

// GetTopLevelNodes returns a list of all nodes that have no connected input pins with the given topology.
func (dg *DaggerGraph) GetTopLevelNodes(topologySystem int) []IDaggerNode {
	var retv []IDaggerNode
	for _, node := range dg.nodes {
		if node.IsTopLevel(topologySystem) {
			retv = append(retv, node)
		}
	}
	return retv
}

// GetEnableTopology returns the value of enableTopology.
func (dg *DaggerGraph) GetEnableTopology() bool {
	return dg.topologyEnabled
}

// SetEnableTopology sets the value of enableTopology.
func (dg *DaggerGraph) SetEnableTopology(enabled bool) {
	if enabled == dg.topologyEnabled {
		return
	}

	// if we are re-enabling the topology, calculate it now
	dg.topologyEnabled = enabled
	dg.CalculateTopology()
}

// GetNodes returns all the nodes in the graph.
func (dg *DaggerGraph) GetNodes() []IDaggerNode {
	return dg.nodes
}

// GetMaxOrdinal returns the highest ordinal for the given topology system.
func (dg *DaggerGraph) GetMaxOrdinal(topologySystem int) int {
	return dg.maxOrdinal[topologySystem]
}

// GetSubGraphCount returns the number of subgraphs in the DaggerGraph.
func (dg *DaggerGraph) GetSubGraphCount(topologySystem int) int {
	return dg.subGraphCount[topologySystem]
}

// GetTopologyCount returns the number of topology systems in the DaggerGraph.
func (dg *DaggerGraph) GetTopologyCount() int {
	return dg.topologyCount
}

// BeforePinsConnected is called before two pins are connected to test if they are currently allowed to connect.
// Override to provide logic for cases when pins are not allowed to connect (e.g., when data is being processed).
// If the method returns false, the pins will fail to connect.
func (dg *DaggerGraph) BeforePinsConnected(connectFrom IDaggerOutputPin, connectTo IDaggerInputPin) bool {
	return true
}

// AfterPinsConnected is called after two pins are connected to test if either needs to be cloned.
func (dg *DaggerGraph) AfterPinsConnected(connectFrom IDaggerOutputPin, connectTo IDaggerInputPin) {
	if connectFrom.GetParentNode().ShouldClonePin(connectFrom) {
		if connectFrom.GetParentNode().ClonePin(connectFrom, nil) == nil {
			dg.EmitError("failed to autoclone pin")
		}
	}

	if connectTo.GetParentNode().ShouldClonePin(connectTo) {
		if connectTo.GetParentNode().ClonePin(connectTo, nil) == nil {
			dg.EmitError("failed to autoclone pin")
		}
	}
}

// BeforePinsDisconnected is called before two pins are disconnected to test if they are currently allowed to disconnect.
// Override to provide logic for cases when pins are not allowed to disconnect (e.g., when data is being processed).
// If the method returns false, the pins will fail to disconnect.
func (dg *DaggerGraph) BeforePinsDisconnected(connectFrom IDaggerBasePin, connectTo IDaggerBasePin) bool {
	return true
}

// AfterPinsDisconnected is called after two pins are disconnected to test if either needs to be removed.
func (dg *DaggerGraph) AfterPinsDisconnected(connectFrom IDaggerBasePin, connectTo IDaggerBasePin) {
	if connectFrom.GetParentNode().ShouldRemoveClonePin(connectFrom) {
		if !connectFrom.GetParentNode().RemoveClonePin(connectFrom) {
			dg.EmitError("failed to remove autocloned pin")
		}
	}

	if connectTo.GetParentNode().ShouldRemoveClonePin(connectTo) {
		if !connectTo.GetParentNode().RemoveClonePin(connectTo) {
			dg.EmitError("failed to remove autocloned pin")
		}
	}
}

// OnPinsDisconnected is called when pins are disconnected.
func (dg *DaggerGraph) OnPinsDisconnected(disconnectOutput IDaggerOutputPin, disconnectInput IDaggerInputPin) {
	dg.CalculateTopology()
	dg.pinsDisconnected.Emit(disconnectOutput.GetInstanceID(), disconnectInput.GetInstanceID())
}

// OnPinsConnected is called when pins are connected.
func (dg *DaggerGraph) OnPinsConnected(connectFrom IDaggerOutputPin, connectTo IDaggerInputPin) {
	dg.CalculateTopology()
	dg.pinsConnected.Emit(connectFrom, connectTo)
}

// GetBottomLevelNodes returns a list of all nodes that have no connected output pins.
func (dg *DaggerGraph) GetBottomLevelNodes(topologySystem int) []IDaggerNode {
	var retv []IDaggerNode
	for _, node := range dg.nodes {
		if node.IsBottomLevel(topologySystem) {
			retv = append(retv, node)
		}
	}
	return retv
}

// GetSubGraphNodes returns a list of nodes in a certain subgraph with the given topology system.
func (dg *DaggerGraph) GetSubGraphNodes(topologySystem, index int) []IDaggerNode {
	var retv []IDaggerNode
	if index > dg.subGraphCount[topologySystem]-1 {
		// emitError("Subgraph index out of range")
		return retv
	}

	for _, node := range dg.nodes {
		if index == node.GetSubgraphAffiliation(topologySystem) {
			retv = append(retv, node)
		}
	}
	return retv
}

// GetSubGraphs returns an array containing arrays of nodes for each subgraph in the graph.
func (dg *DaggerGraph) GetSubGraphs(topologySystem int) [][]IDaggerNode {
	var retv [][]IDaggerNode
	for i := 0; i < dg.subGraphCount[topologySystem]; i++ {
		retv = append(retv, dg.GetSubGraphNodes(topologySystem, i))
	}
	return retv
}

// GetNodesWithOrdinal returns a list of all nodes with the given ordinal in a topology.
func (dg *DaggerGraph) GetNodesWithOrdinal(topologySystem int, ordinal int) []IDaggerNode {
	var retv []IDaggerNode
	for _, node := range dg.nodes {
		if node.GetOrdinal(topologySystem) == ordinal {
			retv = append(retv, node)
		}
	}
	return retv
}

// GetNodesWithName returns a list of all nodes with the given name.
func (dg *DaggerGraph) GetNodesWithName(name string) []IDaggerNode {
	var retv []IDaggerNode
	for _, node := range dg.nodes {
		if node.GetName() == name {
			retv = append(retv, node)
		}
	}
	return retv
}

// GetPinWithInstanceID finds and returns a pin that has the given instanceID.
func (dg *DaggerGraph) GetPinWithInstanceID(pinInstanceID string) IDaggerBasePin {
	for _, node := range dg.nodes {
		for i := 0; i < dg.topologyCount; i++ {
			retv := node.GetInputPins(i).GetPin(pinInstanceID)
			if retv != nil {
				return retv
			}

			retv = node.GetOutputPins(i).GetPin(pinInstanceID)
			if retv != nil {
				return retv
			}
		}
	}
	return nil
}

// GetNodeWithInstanceID finds and returns a node with the given instanceID.
func (dg *DaggerGraph) GetNodeWithInstanceID(nodeInstanceID string) IDaggerNode {
	for _, node := range dg.nodes {
		if node.GetInstanceID() == nodeInstanceID {
			return node
		}
	}
	return nil
}

// AllConnections returns an array of all DaggerInputPins that are connected.
func (dg *DaggerGraph) AllConnections(topologySystem int) []*DaggerInputPin {
	var retv []*DaggerInputPin
	for _, node := range dg.nodes {
		pins := node.GetInputPins(topologySystem).GetAllPins()
		for _, pin := range pins {
			if pin.IsConnected() {
				retv = append(retv, pin.(*DaggerInputPin))
			}
		}
	}
	return retv
}

// RemoveNode removes a node from the graph. If the node has any connections, they are disconnected first.
func (dg *DaggerGraph) RemoveNode(node *DaggerNode) bool {
	if node == nil {
		return false
	}

	if dg.BeforeNodeRemoved(node) {
		if !node.DisconnectAllPins() {
			// emitError("failed to remove node")
			return false
		}

		for i, n := range dg.nodes {
			if n == node {
				dg.nodes = append(dg.nodes[:i], dg.nodes[i+1:]...)
				break
			}
		}

		node.PurgeAll()

		dg.nodeRemoved.Emit(node.instanceID)

		node.parentGraph = nil

		dg.CalculateTopology()

		return true
	}
	return false
}

// AddNode adds a node to the graph.
func (dg *DaggerGraph) AddNode(node IDaggerNode, calculate bool) IDaggerNode {
	if node.GetParentGraph() != nil {
		// emitError("DaggerNode is already associated with a DaggerGraph")
		return nil
	}

	node.BeforeAddedToGraph().Emit()
	node.SetParentGraph(dg)
	dg.nodes = append(dg.nodes, node)

	if calculate {
		dg.CalculateTopology()
	} else {
		for t := 0; t < dg.topologyCount; t++ {
			dg.subGraphCount[t]++
			dg.maxOrdinal[t] = max(1, dg.maxOrdinal[t])
			node.SetSubgraphAffiliation(t, dg.subGraphCount[t]+1)
			node.SetOrdinal(t, 0)
			node.SetDescendents(t, make([]IDaggerNode, 0))
		}
	}

	dg.nodeAdded.Emit(node)
	node.AddedToGraph().Emit()
	node.AfterAddedToGraph().Emit()
	return node
}

// AddNodes adds multiple nodes to the graph.
func (dg *DaggerGraph) AddNodes(nodes []*DaggerNode) []*DaggerNode {
	for _, node := range nodes {
		if node.parentGraph != nil {
			// emitError("DaggerNode is already associated with a DaggerGraph")
			return nil
		}
	}

	for _, node := range nodes {
		node.beforeAddedToGraph.Emit()
		node.parentGraph = dg
		dg.nodes = append(dg.nodes, node)
		dg.nodeAdded.Emit(node)
		node.AddedToGraph().Emit()
		node.afterAddedToGraph.Emit()
	}
	dg.CalculateTopology()

	return nodes
}

// BeforeNodeRemoved is called before a node is removed from the graph.
func (dg *DaggerGraph) BeforeNodeRemoved(node *DaggerNode) bool {
	return true
}

// GraphTopologyChanged is called when the topology has changed.
func (dg *DaggerGraph) graphTopologyChanged() {}

// Helper functions

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// containsNode checks if a node is present in a slice of nodes.
func containsNode(nodes []IDaggerNode, node IDaggerNode) bool {
	for _, n := range nodes {
		if n == node {
			return true
		}
	}
	return false
}

// ######################## Depth First Search Topology Algorithm ########################

// CalculateTopologyDepthFirstSearch is called internally every time an action alters the topology of the graph.
// Uses brute force depth first search
func (dg *DaggerGraph) CalculateTopologyDepthFirstSearch() {
	if !dg.topologyEnabled {
		return
	}

	for t := 0; t < dg.topologyCount; t++ {
		dg.maxOrdinal[t] = 0

		for _, node := range dg.nodes {
			node.SetOrdinal(t, -1)
			node.SetSubgraphAffiliation(t, -1)
			node.SetDescendents(t, make([]IDaggerNode, 0))
		}

		tnodes := dg.GetTopLevelNodes(t)

		// Create a List of Sets. Each Set will hold a List of all Nodes that a top level Node touches on it's way to bottom level Nodes.
		// We'll in turn use these Sets to calculate subgraph affiliations.
		var touchedSetList []mapset.Set

		for i, node := range tnodes {
			// top level nodes always have an Ordinal of 0
			node.SetOrdinal(t, 0)

			// create a "touched" Set
			touchedSet := mapset.NewSet()

			allOutPins := node.GetOutputPins(t).GetAllPins()
			for _, output := range allOutPins {
				opin, _ := output.(IDaggerOutputPin)
				connectedToPins := opin.GetConnectedTo()

				for _, inpin := range *connectedToPins {
					newset := dg.recurseCalculateTopologyDepthFirstSearch(1, inpin.GetParentNode(), touchedSet, t)

					for _, setnode := range newset.ToSlice() {
						if !containsNode(node.GetDescendents(t), setnode.(IDaggerNode)) {
							node.SetDescendents(t, append(node.GetDescendents(t), setnode.(IDaggerNode)))
						}
					}

					sort.Slice(node.GetDescendents(t), func(i, j int) bool {
						return node.GetDescendents(t)[i].GetOrdinal(t) < node.GetDescendents(t)[j].GetOrdinal(t)
					})
				}
			}

			touchedSet.Add(node)

			if i == 0 {
				touchedSetList = append(touchedSetList, touchedSet)
			} else {
				merged := false
				for u, set := range touchedSetList {
					intersection := touchedSet.Intersect(set)

					if intersection.Cardinality() > 0 {
						touchedSetList[u] = touchedSetList[u].Union(touchedSet)
						merged = true
						break
					}
				}

				if !merged {
					touchedSetList = append(touchedSetList, touchedSet)
				}
			}
		}

		for i, set := range touchedSetList {
			for _, node := range set.ToSlice() {
				node.(IDaggerNode).SetSubgraphAffiliation(t, i)
			}
		}

		dg.subGraphCount[t] = len(touchedSetList)
	}

	dg.graphTopologyChanged()
	dg.topologyChanged.Emit()
}

// recurseCalculateTopologyDepthFirstSearch is a recursive helper function used in CalculateTopologyDepthFirstSearch.
func (dg *DaggerGraph) recurseCalculateTopologyDepthFirstSearch(level int, node IDaggerNode, touchedSet mapset.Set, topologySystem int) mapset.Set {
	retv := mapset.NewSet()
	if node == nil {
		return retv
	}

	// set our precedence if level is larger than current value
	node.SetOrdinal(topologySystem, max(level, node.GetOrdinal(topologySystem)))
	dg.maxOrdinal[topologySystem] = max(dg.maxOrdinal[topologySystem], node.GetOrdinal(topologySystem))

	allOutPins := node.GetOutputPins(topologySystem).GetAllPins()
	for _, output := range allOutPins {
		opin, _ := output.(IDaggerOutputPin)
		ipins := opin.GetConnectedTo()
		for _, p2 := range *ipins {
			// recurse through all it's connected pins
			newset := dg.recurseCalculateTopologyDepthFirstSearch(level+1, p2.GetParentNode(), touchedSet, topologySystem)
			retv = retv.Union(newset)
		}
	}

	// add this node to the touched Set (we don't want to be included in our own _descendents Set)
	touchedSet.Add(node)

	// recreate our _descendents List from the new set
	rslice := retv.ToSlice()
	for _, snode := range rslice {
		if !containsNode(node.GetDescendents(topologySystem), snode.(IDaggerNode)) {
			descendents := node.GetDescendents(topologySystem)
			node.SetDescendents(topologySystem, append(descendents, snode.(IDaggerNode)))
		}
	}

	sort.Slice(node.GetDescendents(topologySystem), func(i, j int) bool {
		return node.GetDescendents(topologySystem)[i].GetOrdinal(topologySystem) < node.GetDescendents(topologySystem)[j].GetOrdinal(topologySystem)
	})

	retv.Add(node)

	return retv
}
