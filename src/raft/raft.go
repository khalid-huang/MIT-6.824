package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)
import "labrpc"

// import "bytes"
// import "labgob"



//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

const (
	FOLLOWER     = "follower"
	CANDIDATE    = "candidate"
	LEADER       = "leader"

	HEARBEAT_INTERVAL      = 100
	MIN_ELECTION_INTERVAL  = 400
	MAX_ELECTION_INTERVAL  = 500
)

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm int            //该server当前的term
	votedFor int               //在当前term下该server的投票对象，如果没有投，设置为-1
	voteAcquired int           //获得的投票数

	state string                  //server的状态， follower、candidate、leader
	electionTimer *time.Timer
	voteCh chan struct{}       //成功投票的信号
	appendCh chan struct{}     //成功更新log的信息（包含heartbreat的响应）

}



// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	term = rf.currentTerm
	isleader = rf.state == LEADER
	return term, isleader
}


//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}


//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}


//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term int          //candidate's term
	CandidateId int   //cnadidate's id
}

type AppendEntriesArgs struct {
	Term int           //leader's term
	LeaderId int       //leader's id
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term int                //currentTerm
	VoteGrandted bool       //received vote or not
}

type AppendEntriesReply struct {
	Term int
	Success bool
}

//
// example RequestVote RPC handler.
// 远程方法调用 ，其他的peer(state是candidate的)通过远程方法调用这个函数，这个函数是server用于判断是否进行投票
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGrandted = false
	} else if args.Term > rf.currentTerm {
		//fmt.Printf("server %d 状态为 %s, term从%d经由server %d变成%d\n", rf.me, rf.state, rf.currentTerm, args.CandidateId, args.Term)
		rf.currentTerm = args.Term
		rf.updateStateTo(FOLLOWER)
		//rf.voteAcquired = 0 //已获得的投票数重置为0
		rf.votedFor = args.CandidateId
		//fmt.Printf("--- server %d vote for server %d in term %d--- \n ", rf.me, args.CandidateId, rf.currentTerm)
		//reply.term =
		reply.Term = args.Term
		reply.VoteGrandted = true
	} else {
		if rf.votedFor == -1 {
			rf.votedFor = args.CandidateId
			reply.VoteGrandted = true
		} else {
			reply.VoteGrandted = false
		}
	}
	if reply.VoteGrandted == true {
		go func() {
			rf.voteCh <- struct{}{}
		}()
	}
}

//用于处理接收到心跳时的处理函数，由leader远程方法调用follower的AppendEntries
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	rf.mu.Unlock()

	if args.Term < rf.currentTerm {
		reply.Success = false
		reply.Term = rf.currentTerm
	} else if args.Term > rf.currentTerm {
		rf.updateStateTo(FOLLOWER)
		//rf.votedFor = -1
		reply.Success = true
		reply.Term = args.Term
	} else {
		reply.Term = args.Term
		reply.Success = true
	}
	go func() {
		rf.appendCh <- struct{}{}
	}()
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}


//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).


	return index, term, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
}

func randomElectionDuration() time.Duration {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return time.Millisecond * time.Duration(
		r.Int63n(MAX_ELECTION_INTERVAL - MIN_ELECTION_INTERVAL) + MIN_ELECTION_INTERVAL)
}

//重置一些Raft peer的参数
func (rf *Raft) resetRaftState() {
	rf.votedFor = -1
	rf.voteAcquired = 0
}

func (rf *Raft) updateStateTo(state string) {
	if rf.state == state {
		return
	}

	preState := rf.state

	switch state {
	case FOLLOWER:
		rf.state = FOLLOWER
		rf.resetRaftState()
	case CANDIDATE:
		rf.state = CANDIDATE
		rf.resetRaftState()
	case LEADER:
		rf.state = LEADER
		rf.resetRaftState()
	default:
		fmt.Printf("Warning: invalid state %s, do nothing.\n", state)
	}
	fmt.Printf("In term %d: Server %d transfer from %s to %s \n",
		rf.currentTerm, rf.me, preState, rf.state)
}

func (rf *Raft) startElection() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	rf.currentTerm += 1
	rf.votedFor = rf.me
	rf.voteAcquired = 1
	rf.electionTimer.Reset(randomElectionDuration())

	rf.broadcastRequestVote()
}

//广播选举; 遍历全部的peers，发送RequestVote请求
func (rf *Raft) broadcastRequestVote() {
	//封装参数
	args := RequestVoteArgs{
		Term: rf.currentTerm,
		CandidateId: rf.me,
	}

	for i,_ := range rf.peers {
		if i == rf.me {
			continue
		}

		go func(server int) {
			var reply RequestVoteReply
			ok := rf.sendRequestVote(server, &args, &reply)
			if ok == true {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				if reply.VoteGrandted == true {
					rf.voteAcquired += 1
				} else {
					if reply.Term > rf.currentTerm {
						rf.currentTerm = reply.Term
						rf.updateStateTo(FOLLOWER)
						//rf.voteAcquired = 0
						//rf.votedFor = -1
					}
				}
			} else {
				fmt.Printf("Server %d send vote req failed \n", rf.me)
			}
		}(i)
	}
}

//广播心跳
func (rf *Raft) broadcastAppendEntries() {
	args := AppendEntriesArgs{
		Term: rf.currentTerm,
		LeaderId: rf.me,
	}

	for i, _ := range rf.peers {
		if i == rf.me {
			continue
		}

		go func(server int) {
			var reply AppendEntriesReply

			ok := rf.sendAppendEntries(server, &args, &reply)
			if ok == true {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				if reply.Success == true {

				} else {
					if reply.Term > rf.currentTerm {
						rf.currentTerm = reply.Term
						rf.updateStateTo(FOLLOWER)
					}
				}
			} else {
				fmt.Printf("server %d send AppendEntries to server %d failed \n", rf.me, i)
			}
		}(i)
	}
}

func (rf *Raft) startLoop() {
	rf.electionTimer = time.NewTimer(randomElectionDuration())
	for {
		switch  rf.state {
		case FOLLOWER:
			select {
			//投票之后要重置定时器
			case <-rf.voteCh:
				rf.electionTimer.Reset(randomElectionDuration())
			//收到心跳时，要重置定时器
			case <-rf.appendCh:
				rf.electionTimer.Reset(randomElectionDuration())
			//超过限定时间没有收到心跳，发起竞选
			case <-rf.electionTimer.C:
				rf.updateStateTo(CANDIDATE)
				rf.startElection()
			}
		case CANDIDATE:
			select {
			//收到来自leader的心跳
			case <-rf.appendCh:
				rf.updateStateTo(FOLLOWER)
				rf.votedFor = -1
			//超时
			case <-rf.electionTimer.C:
				rf.electionTimer.Reset(randomElectionDuration())
				rf.startElection()
			}
			//检验是否投票数达到要求
			fmt.Printf("server %d 在term为%d获得的票数为 %d： \n", rf.me, rf.currentTerm, rf.voteAcquired)
			if rf.voteAcquired > len(rf.peers) / 2 {
				fmt.Printf("半数以上是大于：%d\n", len(rf.peers) / 2)
				fmt.Printf("server %d 在term为%d获得的票数为 %d：成为LAEDER \n", rf.me, rf.currentTerm, rf.voteAcquired)
				rf.updateStateTo(LEADER)
				rf.broadcastAppendEntries()
			}

		case LEADER:
			//发送广播心跳
			rf.broadcastAppendEntries()
			time.Sleep(HEARBEAT_INTERVAL * time.Millisecond)
		}
	}
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.currentTerm = 1
	rf.votedFor = -1
	rf.state = FOLLOWER
	rf.appendCh = make(chan struct{})
	rf.voteCh = make(chan struct{})

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go rf.startLoop()

	return rf
}
