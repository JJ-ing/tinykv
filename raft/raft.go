// Copyright 2015 The etcd Authors
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

package raft

import (
	"errors"
	pb "github.com/pingcap-incubator/tinykv/proto/pkg/eraftpb"
	"math/rand"
)

// None is a placeholder node ID used when there is no leader.
const None uint64 = 0

// StateType represents the role of a node in a cluster.
type StateType uint64

const (
	StateFollower StateType = iota
	StateCandidate
	StateLeader
)

var stmap = [...]string{
	"StateFollower",
	"StateCandidate",
	"StateLeader",
}

func (st StateType) String() string {
	return stmap[uint64(st)]
}

// ErrProposalDropped is returned when the proposal is ignored by some cases,
// so that the proposer can be notified and fail fast.
var ErrProposalDropped = errors.New("raft proposal dropped")

// Config contains the parameters to start a raft.
type Config struct {
	// ID is the identity of the local raft. ID cannot be 0.
	ID uint64

	// peers contains the IDs of all nodes (including self) in the raft cluster. It
	// should only be set when starting a new raft cluster. Restarting raft from
	// previous configuration will panic if peers is set. peer is private and only
	// used for testing right now.
	peers []uint64

	// ElectionTick is the number of Node.Tick invocations that must pass between
	// elections. That is, if a follower does not receive any message from the
	// leader of current term before ElectionTick has elapsed, it will become
	// candidate and start an election. ElectionTick must be greater than
	// HeartbeatTick. We suggest ElectionTick = 10 * HeartbeatTick to avoid
	// unnecessary leader switching.
	ElectionTick int
	// HeartbeatTick is the number of Node.Tick invocations that must pass between
	// heartbeats. That is, a leader sends heartbeat messages to maintain its
	// leadership every HeartbeatTick ticks.
	HeartbeatTick int

	// Storage is the storage for raft. raft generates entries and states to be
	// stored in storage. raft reads the persisted entries and states out of
	// Storage when it needs. raft reads out the previous state and configuration
	// out of storage when restarting.
	Storage Storage
	// Applied is the last applied index. It should only be set when restarting
	// raft. raft will not return entries to the application smaller or equal to
	// Applied. If Applied is unset when restarting, raft might return previous
	// applied entries. This is a very application dependent configuration.
	Applied uint64
}

func (c *Config) validate() error {
	if c.ID == None {
		return errors.New("cannot use none as id")
	}

	if c.HeartbeatTick <= 0 {
		return errors.New("heartbeat tick must be greater than 0")
	}

	if c.ElectionTick <= c.HeartbeatTick {
		return errors.New("election tick must be greater than heartbeat tick")
	}

	if c.Storage == nil {
		return errors.New("storage cannot be nil")
	}

	return nil
}

// Progress represents a follower’s progress in the view of the leader. Leader maintains
// progresses of all followers, and sends entries to the follower based on its progress.
type Progress struct {
	Match, Next uint64
}

type Raft struct {
	id uint64

	Term uint64
	Vote uint64

	// the log
	RaftLog *RaftLog

	// log replication progress of each peers
	Prs map[uint64]*Progress

	// this peer's role
	State StateType

	// votes records
	votes map[uint64]bool

	// msgs need to send
	msgs []pb.Message

	// the leader id
	Lead uint64

	// heartbeat interval, should send
	heartbeatTimeout int
	// baseline of election interval
	electionTimeout int

	// random election interval
	randomElectionTimeout int

	// number of ticks since it reached last heartbeatTimeout.
	// only leader keeps heartbeatElapsed.
	heartbeatElapsed int
	// number of ticks since it reached last electionTimeout
	electionElapsed int

	// leadTransferee is id of the leader transfer target when its value is not zero.
	// Follow the procedure defined in section 3.10 of Raft phd thesis.
	// (https://web.stanford.edu/~ouster/cgi-bin/papers/OngaroPhD.pdf)
	// (Used in 3A leader transfer)
	leadTransferee uint64

	// Only one conf change may be pending (in the log, but not yet
	// applied) at a time. This is enforced via PendingConfIndex, which
	// is set to a value >= the log index of the latest pending
	// configuration change (if any). Config changes are only allowed to
	// be proposed if the leader's applied index is greater than this
	// value.
	// (Used in 3A conf change)
	PendingConfIndex uint64
}

// newRaft return a raft peer with the given config
func newRaft(c *Config) *Raft {
	if err := c.validate(); err != nil {
		panic(err.Error())
	}
	// Your Code Here (2A).
	id := c.ID
	RaftLog := newLog(c.Storage)

	Prs := make(map[uint64]*Progress)
	for i:=0; i<len(c.peers); i++{
		Prs[c.peers[i]] = &Progress{0, 0}
	}

	votes := make(map[uint64]bool)

	msgs := make([]pb.Message, 0)

	raft := &Raft{
		id: id, Term: 0, RaftLog: RaftLog, Prs: Prs, votes: votes, msgs: msgs,
		heartbeatTimeout: c.HeartbeatTick,
		electionTimeout: c.ElectionTick,
		randomElectionTimeout: c.ElectionTick + rand.Intn(c.ElectionTick),
	}
	//...
	//fmt.Printf("raft.log:  %d --------------------\n", raft.RaftLog.LastIndex())
	return raft
}

// sendAppend sends an append RPC with new entries (if any) and the
// current commit index to the given peer. Returns true if a message was sent.
func (r *Raft) sendAppend(to uint64) bool {
	// Your Code Here (2A).
	preLogIndex := r.Prs[to].Next - 1
	preLogTerm, _ := r.RaftLog.Term(preLogIndex)
	tempEntries := make([]*pb.Entry, 0)
	//for i:=r.Prs[to].Next; i<= r.RaftLog.LastIndex(); i++{
	//	tempEntries = append(tempEntries, &pb.Entry{RaftLog.entries[i]})
	//}
	msg := pb.Message{
		MsgType: pb.MessageType_MsgAppend,
		From: r.id,
		To: to,
		Term: r.Term,
		LogTerm: preLogTerm,
		Index: preLogIndex,
		Entries: tempEntries,
		Commit: r.RaftLog.committed,
		}
	r.msgs = append(r.msgs, msg)
	return true
}

//// broadcastSendAppend send MessageType_MsgAppend to every others
//func (r *Raft) broadcastSendAppend(){
//	for i:= range r.Prs{
//		if i!=r.id{
//
//		}
//	}
//}

// sendHeartbeat sends a heartbeat RPC to the given peer.
func (r *Raft) sendHeartbeat(to uint64) {
	// Your Code Here (2A).
	msg := pb.Message{
		MsgType: pb.MessageType_MsgHeartbeat,
		To: to,
		From: r.id,
		Term: r.Term,
		//Commit: min(r.RaftLog.committed, r.Prs[to].Match),
	}
	r.msgs = append(r.msgs, msg)
}

// broadcastHeartBeat send heart beat to every others
func (r *Raft) broadcastHeartBeat(){
	for i:= range r.Prs{
		if i!=r.id{
			r.sendHeartbeat(i)
		}
	}
}


//sendRequestVote sends a RequestVote RPC to the given peer
func (r *Raft) sendRequestVote(to uint64) {
	//fmt.Printf("in send requestvote")
	lastLogIndex := r.RaftLog.LastIndex()
	lastLogTerm, _ := r.RaftLog.Term(lastLogIndex)

	msg := pb.Message{
		MsgType: pb.MessageType_MsgRequestVote,
		To: to,
		From: r.id,
		Term: r.Term,	// term+1 before call the function
		Index: lastLogIndex,
		LogTerm: lastLogTerm,
	}
	r.msgs = append(r.msgs, msg)
}

//broadcastRequestVote start a new election
func (r *Raft) broadcastRequestVote(){
	if len(r.Prs)==1{	// only one node
		r.becomeLeader()
		return
	}

	for i:= range r.Prs{	//broadcast vote request
		if i!=r.id {
			//fmt.Printf("vote request --------> %d\n", i)
			r.sendRequestVote(i)
		}
	}
}
// tick advances the internal logical clock by a single tick.
func (r *Raft) tick() {
	// Your Code Here (2A).
	switch r.State {
	case StateLeader:
		r.heartbeatElapsed = r.heartbeatElapsed+1
		if r.heartbeatElapsed>=r.heartbeatTimeout{
			//reset
			r.heartbeatElapsed = 0
			//signal to send heartbeat, local message
			_ = r.Step(pb.Message{MsgType: pb.MessageType_MsgBeat})
		}
	case StateCandidate:
		r.electionElapsed = r.electionElapsed+1
		if r.electionElapsed>=r.randomElectionTimeout{
			r.electionElapsed = 0
			//signal to start a new election, local message
			_ = r.Step(pb.Message{MsgType: pb.MessageType_MsgHup})
		}
	case StateFollower:
		r.electionElapsed = r.electionElapsed+1
		if r.electionElapsed>=r.randomElectionTimeout{
			r.electionElapsed = 0
			//signal to start a new election, local message
			_ = r.Step(pb.Message{MsgType: pb.MessageType_MsgHup})
		}
	}
}

// becomeFollower transform this peer's state to Follower
func (r *Raft) becomeFollower(term uint64, lead uint64) {
	// Your Code Here (2A).
	r.State = StateFollower
	//r.electionTimeout = 150 + rand.Intn(150)	// random election timeout (150-300ms)
	r.randomElectionTimeout = r.electionTimeout + rand.Intn(r.electionTimeout)
	r.electionElapsed = 0
	r.Term = term
	r.Lead = lead
	r.Vote = None

}

// becomeCandidate transform this peer's state to candidate
func (r *Raft) becomeCandidate() {
	// Your Code Here (2A).
	r.State = StateCandidate
	//r.electionTimeout = 150 + rand.Intn(150)	// random election timeout (150-300ms)
	r.randomElectionTimeout = r.electionTimeout + rand.Intn(r.electionTimeout)
	r.electionElapsed = 0
	r.Term = r.Term + 1
	r.Lead = None
	r.Vote = r.id
	r.votes = make(map[uint64]bool)
	r.votes[r.id] = true
}

// becomeLeader transform this peer's state to leader
func (r *Raft) becomeLeader() {
	// Your Code Here (2A).
	// NOTE: Leader should propose a noop entry on its term
	r.State = StateLeader
	r.heartbeatElapsed = 0
	r.Lead = r.id

	lastIndex := r.RaftLog.LastIndex()
	for i := range r.Prs {
		if i==r.id{
			r.Prs[i].Match = lastIndex +1
			r.Prs[i].Next = lastIndex +1 +1
		}else{
			r.Prs[i].Match = 0
			r.Prs[i].Next = lastIndex+1
		}
	}

	// --- Comment for pass 2aa
	//noop := pb.Entry{EntryType: pb.EntryType_EntryNormal,
	//				Term: r.Term,
	//				Index: lastIndex+1,
	//				Data: nil}
	//r.RaftLog.entries = append(r.RaftLog.entries, noop)


	r.broadcastHeartBeat()
	//if(len(r.Prs)==1){
	//	r.UpdateCommit()
	//} else{
	//	//Boardcast msgs to other Followers
	//	for j:=1;j<=len(r.Prs);j++{
	//		if uint64(j)!=r.id {
	//			r.sendAppend(uint64(j))
	//		}
	//	}
	//}

}

// Step the entrance of handle message, see `MessageType`
// on `eraftpb.proto` for what msgs should be handled
func (r *Raft) Step(m pb.Message) error {
	// Your Code Here (2A).
	//fmt.Printf("Step-> r.state:%d, r.id:%d, r.term:%d\n", r.State, r.id, r.Term)
	//fmt.Println(m)
	var err error
	switch r.State {
	case StateFollower:
		err = r.FollowerStep(m)
	case StateCandidate:
		err = r.CandidateStep(m)
	case StateLeader:
		err = r.LeaderStep(m)
	}
	return err
}

func (r *Raft) FollowerStep(m pb.Message) error{
	switch m.MsgType {
	case pb.MessageType_MsgHup:
		r.becomeCandidate()
		//fmt.Printf("FollowerStep-> r.state:%d, r.id:%d, r.term:%d\n", r.State, r.id, r.Term)
		//fmt.Printf("candidate to request vote\n")
		r.broadcastRequestVote()
	case pb.MessageType_MsgBeat:
	case pb.MessageType_MsgPropose:
	case pb.MessageType_MsgAppend:
		r.handleAppendEntries(m)	// to deal with entries
	case pb.MessageType_MsgAppendResponse:
	case pb.MessageType_MsgRequestVote:
		r.handleRequestVote(m)
	case pb.MessageType_MsgRequestVoteResponse:
	case pb.MessageType_MsgSnapshot:
	case pb.MessageType_MsgHeartbeat:
		r.handleHeartbeat(m)
	case pb.MessageType_MsgHeartbeatResponse:
	case pb.MessageType_MsgTransferLeader:	//redirection

	case pb.MessageType_MsgTimeoutNow:	//#?
	}
	return nil
}

func (r *Raft) CandidateStep(m pb.Message) error{
	switch m.MsgType {
	case pb.MessageType_MsgHup:
		r.becomeCandidate()
		r.broadcastRequestVote()
	case pb.MessageType_MsgBeat:
	case pb.MessageType_MsgPropose:
	case pb.MessageType_MsgAppend:
		r.handleAppendEntries(m)
	case pb.MessageType_MsgAppendResponse:
	case pb.MessageType_MsgRequestVote:
		r.handleRequestVote(m)
	case pb.MessageType_MsgRequestVoteResponse:
		r.handleRequestVoteResponse(m)
	case pb.MessageType_MsgSnapshot:
	case pb.MessageType_MsgHeartbeat:
		r.handleHeartbeat(m)
	case pb.MessageType_MsgHeartbeatResponse:
	case pb.MessageType_MsgTransferLeader:
	case pb.MessageType_MsgTimeoutNow:
	}
	return nil
}

func (r *Raft) LeaderStep(m pb.Message) error{
	switch m.MsgType {
	case pb.MessageType_MsgHup:
	case pb.MessageType_MsgBeat:
		r.broadcastHeartBeat()
	case pb.MessageType_MsgPropose:
		//...add entries
	case pb.MessageType_MsgAppend:
		r.handleAppendEntries(m)		//#?
	case pb.MessageType_MsgAppendResponse:
		r.handleAppendEntriesResponse(m)
	case pb.MessageType_MsgRequestVote:
		r.handleRequestVote(m)	 // a new election has started
	case pb.MessageType_MsgRequestVoteResponse:
	case pb.MessageType_MsgSnapshot:
	case pb.MessageType_MsgHeartbeat:
		//r.handleHeartbeat(m)	//#?
	case pb.MessageType_MsgHeartbeatResponse:
		// deal with response of heartbeat RPC request
		if m.Reject==true{	// a peer's term is higher than myself
			r.becomeFollower(m.Term, m.From)
		}else{
			r.sendAppend(m.From)	// the peer is alive, then sendAppend
		}
	case pb.MessageType_MsgTransferLeader:
	case pb.MessageType_MsgTimeoutNow:
	}
	return nil
}


// handleAppendEntries handle AppendEntries RPC request
func (r *Raft) handleAppendEntries(m pb.Message) {
	// Your Code Here (2A).
	msg := pb.Message{
		MsgType: pb.MessageType_MsgAppendResponse,
		From: r.id,
		To: m.From,
		Term: r.Term,
		Reject: false,
		Index: m.Index+uint64(len(m.Entries)),	//record the last log index, lead needs it when deal with the response
	}
	if m.Term < r.Term {
		msg.Reject = true
	}else{	// reset
		r.becomeFollower(m.Term, m.From)
	}
	if t, _ :=r.RaftLog.Term(m.Index); t!=m.LogTerm{
		msg.Reject = true
	}

	//deal with entries
	//...
	if m.Commit > r.RaftLog.committed{
		r.RaftLog.committed = min(m.Commit, m.Index+uint64(len(m.Entries)))
	}
	return
}

// handleAppendEntriesResponse handle response of AppendEntries RPC request
func (r *Raft) handleAppendEntriesResponse(m pb.Message) {
	if m.Reject==true{	// failed
		if m.Term > r.Term{		//a higher term
			r.becomeFollower(m.Term, m.From)
		}else{					//conflict in log
			r.Prs[m.From].Next --
			r.sendAppend(m.From)
		}
	}else{				// ok
		r.Prs[m.From].Match = m.Index
		r.Prs[m.From].Next = m.Index +1
	}
}

// handleHeartbeat handle Heartbeat RPC request
func (r *Raft) handleHeartbeat(m pb.Message) {
	// Your Code Here (2A).
	msg := pb.Message{
		MsgType: pb.MessageType_MsgHeartbeatResponse,
		From: r.id,
		To: m.From,
		Term: r.Term,
		Reject: false,
	}
	if m.Term < r.Term {
		msg.Reject = true
	}else{	// know the leader
		r.becomeFollower(m.Term, m.From)
	}
	return
}

//handleRequestVote handle RequestVote RPC request.
func (r *Raft) handleRequestVote(m pb.Message) {
	msg := pb.Message{
		MsgType: pb.MessageType_MsgRequestVoteResponse,
		To: m.From,
		From: r.id,
		Term: r.Term,
		Reject: true,	//default reject
	}
	isNewer := false
	cur,_ :=r.RaftLog.Term(r.RaftLog.LastIndex())
	if m.LogTerm>cur{
		isNewer = true
	}else if cur==m.LogTerm && m.Index>=r.RaftLog.LastIndex(){
			isNewer = true
	}
	//fmt.Printf("handle-> m.from:%d, m.Term:%d, r.id:%d, r.Term:%d, isNewer:%v\n", m.From, m.Term, r.id, r.Term, isNewer)
	//if !isNewer{
		//fmt.Printf("isNewer-> m.LogTerm:%d, m.Index:%d, r.lastTerm:%d, r.lastIndex:%d\n",
		//	m.LogTerm, m.Index, cur, r.RaftLog.LastIndex())
	//}
	if (m.Term>r.Term ||(m.Term==r.Term && (r.Vote==None||r.Vote==m.From))) && isNewer {	//term is higher and log is newer
		msg.Reject = false
		r.becomeFollower(m.Term, m.From)	// follower --> follower (reset)
											// candidate --> follower
											// leader --> follower
		r.Vote = m.From				// first become follower, then set the vote
		r.Term = m.Term				// press: to compare the term with other vote request's
	}
	r.msgs = append(r.msgs, msg)
}

//handleRequestVoteResponse handle the response of vote request
func (r *Raft) handleRequestVoteResponse(m pb.Message){
	if m.Reject==true{
		if m.Term>r.Term{
			r.becomeFollower(m.Term, m.From)
		}
		return
	}else{
		support := 0
		r.votes[m.From] = true
		for i := range r.votes{
			if r.votes[i]==true{
				support ++
			}
		}
		if support>len(r.Prs)/2{	// get the support of most peer
			r.becomeLeader()		// note: len(r.Prs)
		}
	}
}

// handleSnapshot handle Snapshot RPC request
func (r *Raft) handleSnapshot(m pb.Message) {
	// Your Code Here (2C).
}

// addNode add a new node to raft group
func (r *Raft) addNode(id uint64) {
	// Your Code Here (3A).
}

// removeNode remove a node from raft group
func (r *Raft) removeNode(id uint64) {
	// Your Code Here (3A).
}
