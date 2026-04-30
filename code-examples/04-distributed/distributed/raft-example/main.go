// Raft 一致性算法 — 简化模拟实现
// 演示：Leader 选举（Term/投票/随机超时）、日志复制（AppendEntries）、多节点通信
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：
//   go run ./raft-example/
//
// 本示例为纯 Go 实现，使用 channel 模拟节点间 RPC 通信

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================================
// Raft 核心数据结构
// ============================================================

// NodeRole 节点角色
type NodeRole int

const (
	Follower  NodeRole = iota // 跟随者：被动响应请求
	Candidate                 // 候选者：发起选举
	Leader                    // 领导者：处理写请求，发送心跳
)

func (r NodeRole) String() string {
	switch r {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

// LogEntry Raft 日志条目
type LogEntry struct {
	Term    int    // 日志所属的任期
	Index   int    // 日志索引
	Command string // 命令内容（如 "SET x=1"）
}

// VoteRequest 投票请求（RequestVote RPC）
type VoteRequest struct {
	Term         int // 候选者的任期
	CandidateID  int // 候选者 ID
	LastLogIndex int // 候选者最后一条日志的索引
	LastLogTerm  int // 候选者最后一条日志的任期
}

// VoteResponse 投票响应
type VoteResponse struct {
	Term        int  // 响应者的当前任期
	VoteGranted bool // 是否投票
	VoterID     int  // 投票者 ID
}

// AppendEntriesRequest 日志追加请求（AppendEntries RPC / 心跳）
type AppendEntriesRequest struct {
	Term         int        // Leader 的任期
	LeaderID     int        // Leader ID
	PrevLogIndex int        // 前一条日志的索引
	PrevLogTerm  int        // 前一条日志的任期
	Entries      []LogEntry // 要追加的日志条目（空则为心跳）
	LeaderCommit int        // Leader 的 commitIndex
}

// AppendEntriesResponse 日志追加响应
type AppendEntriesResponse struct {
	Term    int  // 响应者的当前任期
	Success bool // 是否成功
	NodeID  int  // 响应者 ID
}

// ============================================================
// Raft 节点实现
// ============================================================

// RaftNode Raft 节点
type RaftNode struct {
	mu sync.Mutex

	// 节点标识
	id    int
	peers []int // 其他节点 ID 列表

	// 持久化状态（每次变更需持久化，此处简化为内存）
	currentTerm int        // 当前任期
	votedFor    int        // 当前任期投票给了谁（-1 表示未投票）
	log         []LogEntry // 日志条目

	// 易失性状态
	role        NodeRole
	commitIndex int // 已提交的最高日志索引
	lastApplied int // 已应用到状态机的最高日志索引

	// Leader 专用状态
	nextIndex  map[int]int // 每个 Follower 的下一条日志索引
	matchIndex map[int]int // 每个 Follower 已复制的最高日志索引

	// 通信通道（模拟 RPC）
	voteReqCh     chan VoteRequest
	voteRespCh    chan VoteResponse
	appendReqCh   chan AppendEntriesRequest
	appendRespCh  chan AppendEntriesResponse

	// 控制
	stopCh chan struct{}
	logger func(format string, args ...interface{})
}

// NewRaftNode 创建 Raft 节点
func NewRaftNode(id int, peers []int) *RaftNode {
	node := &RaftNode{
		id:           id,
		peers:        peers,
		currentTerm:  0,
		votedFor:     -1,
		log:          make([]LogEntry, 0),
		role:         Follower,
		commitIndex:  -1,
		lastApplied:  -1,
		nextIndex:    make(map[int]int),
		matchIndex:   make(map[int]int),
		voteReqCh:    make(chan VoteRequest, 100),
		voteRespCh:   make(chan VoteResponse, 100),
		appendReqCh:  make(chan AppendEntriesRequest, 100),
		appendRespCh: make(chan AppendEntriesResponse, 100),
		stopCh:       make(chan struct{}),
		logger: func(format string, args ...interface{}) {
			prefix := fmt.Sprintf("[Node %d] ", id)
			fmt.Printf(prefix+format+"\n", args...)
		},
	}
	return node
}

// lastLogInfo 获取最后一条日志的索引和任期
func (n *RaftNode) lastLogInfo() (int, int) {
	if len(n.log) == 0 {
		return -1, -1
	}
	last := n.log[len(n.log)-1]
	return last.Index, last.Term
}

// ============================================================
// Raft 集群（协调多个节点通信）
// ============================================================

// RaftCluster Raft 集群
type RaftCluster struct {
	nodes map[int]*RaftNode
	mu    sync.RWMutex
}

// NewRaftCluster 创建 Raft 集群
func NewRaftCluster(nodeCount int) *RaftCluster {
	cluster := &RaftCluster{
		nodes: make(map[int]*RaftNode),
	}

	// 构建节点 ID 列表
	ids := make([]int, nodeCount)
	for i := 0; i < nodeCount; i++ {
		ids[i] = i
	}

	// 创建节点
	for i := 0; i < nodeCount; i++ {
		peers := make([]int, 0, nodeCount-1)
		for _, id := range ids {
			if id != i {
				peers = append(peers, id)
			}
		}
		cluster.nodes[i] = NewRaftNode(i, peers)
	}

	return cluster
}

// sendVoteRequest 发送投票请求到目标节点
func (c *RaftCluster) sendVoteRequest(targetID int, req VoteRequest) {
	c.mu.RLock()
	node, ok := c.nodes[targetID]
	c.mu.RUnlock()
	if ok {
		select {
		case node.voteReqCh <- req:
		default:
			// 通道满，丢弃（模拟网络丢包）
		}
	}
}

// sendVoteResponse 发送投票响应到目标节点
func (c *RaftCluster) sendVoteResponse(targetID int, resp VoteResponse) {
	c.mu.RLock()
	node, ok := c.nodes[targetID]
	c.mu.RUnlock()
	if ok {
		select {
		case node.voteRespCh <- resp:
		default:
		}
	}
}

// sendAppendEntries 发送日志追加请求到目标节点
func (c *RaftCluster) sendAppendEntries(targetID int, req AppendEntriesRequest) {
	c.mu.RLock()
	node, ok := c.nodes[targetID]
	c.mu.RUnlock()
	if ok {
		select {
		case node.appendReqCh <- req:
		default:
		}
	}
}

// sendAppendResponse 发送日志追加响应到目标节点
func (c *RaftCluster) sendAppendResponse(targetID int, resp AppendEntriesResponse) {
	c.mu.RLock()
	node, ok := c.nodes[targetID]
	c.mu.RUnlock()
	if ok {
		select {
		case node.appendRespCh <- resp:
		default:
		}
	}
}

// Run 启动集群中所有节点的事件循环
func (c *RaftCluster) Run(duration time.Duration) {
	var wg sync.WaitGroup

	for _, node := range c.nodes {
		wg.Add(1)
		go func(n *RaftNode) {
			defer wg.Done()
			c.runNode(n)
		}(node)
	}

	// 运行指定时间后停止
	time.Sleep(duration)
	for _, node := range c.nodes {
		close(node.stopCh)
	}
	wg.Wait()
}

// runNode 运行单个节点的事件循环
func (c *RaftCluster) runNode(n *RaftNode) {
	// 随机选举超时（150ms ~ 300ms）
	electionTimeout := func() time.Duration {
		return time.Duration(150+rand.Intn(150)) * time.Millisecond
	}

	heartbeatInterval := 50 * time.Millisecond
	electionTimer := time.NewTimer(electionTimeout())
	heartbeatTimer := time.NewTimer(heartbeatInterval)
	heartbeatTimer.Stop() // 初始不是 Leader，不发心跳

	for {
		select {
		case <-n.stopCh:
			return

		// 选举超时：Follower/Candidate 发起选举
		case <-electionTimer.C:
			n.mu.Lock()
			if n.role != Leader {
				n.startElection(c)
			}
			n.mu.Unlock()
			electionTimer.Reset(electionTimeout())

		// 心跳定时器：Leader 发送心跳
		case <-heartbeatTimer.C:
			n.mu.Lock()
			if n.role == Leader {
				n.sendHeartbeats(c)
			}
			n.mu.Unlock()
			heartbeatTimer.Reset(heartbeatInterval)

		// 处理投票请求
		case req := <-n.voteReqCh:
			n.mu.Lock()
			resp := n.handleVoteRequest(req)
			n.mu.Unlock()
			c.sendVoteResponse(req.CandidateID, resp)
			if resp.VoteGranted {
				electionTimer.Reset(electionTimeout())
			}

		// 处理投票响应
		case resp := <-n.voteRespCh:
			n.mu.Lock()
			becameLeader := n.handleVoteResponse(resp)
			if becameLeader {
				heartbeatTimer.Reset(0) // 立即发送心跳
				electionTimer.Stop()
			}
			n.mu.Unlock()

		// 处理日志追加请求（心跳/日志复制）
		case req := <-n.appendReqCh:
			n.mu.Lock()
			resp := n.handleAppendEntries(req)
			n.mu.Unlock()
			c.sendAppendResponse(req.LeaderID, resp)
			electionTimer.Reset(electionTimeout()) // 收到心跳，重置选举超时

		// 处理日志追加响应
		case resp := <-n.appendRespCh:
			n.mu.Lock()
			n.handleAppendResponse(resp)
			n.mu.Unlock()
		}
	}
}

// ============================================================
// Raft 核心算法实现
// ============================================================

// startElection 发起选举
func (n *RaftNode) startElection(c *RaftCluster) {
	n.currentTerm++
	n.role = Candidate
	n.votedFor = n.id
	n.logger("发起选举，Term=%d", n.currentTerm)

	lastIdx, lastTerm := n.lastLogInfo()

	// 向所有 peer 发送投票请求
	for _, peerID := range n.peers {
		go c.sendVoteRequest(peerID, VoteRequest{
			Term:         n.currentTerm,
			CandidateID:  n.id,
			LastLogIndex: lastIdx,
			LastLogTerm:  lastTerm,
		})
	}
}

// handleVoteRequest 处理投票请求
func (n *RaftNode) handleVoteRequest(req VoteRequest) VoteResponse {
	// 如果请求的 Term 更大，更新自己的 Term 并转为 Follower
	if req.Term > n.currentTerm {
		n.currentTerm = req.Term
		n.role = Follower
		n.votedFor = -1
	}

	// 投票条件：
	// 1. 请求的 Term >= 自己的 Term
	// 2. 本任期内未投票或已投给该候选者
	// 3. 候选者的日志至少和自己一样新
	granted := false
	if req.Term >= n.currentTerm &&
		(n.votedFor == -1 || n.votedFor == req.CandidateID) {
		myLastIdx, myLastTerm := n.lastLogInfo()
		// 日志新旧比较：先比 Term，再比 Index
		if req.LastLogTerm > myLastTerm ||
			(req.LastLogTerm == myLastTerm && req.LastLogIndex >= myLastIdx) {
			granted = true
			n.votedFor = req.CandidateID
		}
	}

	return VoteResponse{
		Term:        n.currentTerm,
		VoteGranted: granted,
		VoterID:     n.id,
	}
}

// voteCount 用于追踪投票计数（简化实现，存储在节点中）
var voteCountMap = sync.Map{}

// handleVoteResponse 处理投票响应
// 返回是否成为 Leader
func (n *RaftNode) handleVoteResponse(resp VoteResponse) bool {
	if n.role != Candidate {
		return false
	}

	if resp.Term > n.currentTerm {
		n.currentTerm = resp.Term
		n.role = Follower
		n.votedFor = -1
		return false
	}

	if resp.VoteGranted {
		// 使用 sync.Map 追踪投票计数
		key := fmt.Sprintf("%d-%d", n.id, n.currentTerm)
		val, _ := voteCountMap.LoadOrStore(key, new(int32))
		count := val.(*int32)
		*count++

		// 加上自己的一票
		totalVotes := int(*count) + 1
		majority := (len(n.peers)+1)/2 + 1

		if totalVotes >= majority {
			n.role = Leader
			n.logger("🎉 成为 Leader！Term=%d，获得 %d/%d 票",
				n.currentTerm, totalVotes, len(n.peers)+1)

			// 初始化 Leader 状态
			for _, peerID := range n.peers {
				n.nextIndex[peerID] = len(n.log)
				n.matchIndex[peerID] = -1
			}
			return true
		}
	}
	return false
}

// sendHeartbeats Leader 发送心跳（空的 AppendEntries）
func (n *RaftNode) sendHeartbeats(c *RaftCluster) {
	for _, peerID := range n.peers {
		prevIdx := -1
		prevTerm := -1
		if ni, ok := n.nextIndex[peerID]; ok && ni > 0 && ni-1 < len(n.log) {
			prevIdx = n.log[ni-1].Index
			prevTerm = n.log[ni-1].Term
		}

		// 检查是否有新日志需要复制
		var entries []LogEntry
		if ni, ok := n.nextIndex[peerID]; ok && ni < len(n.log) {
			entries = n.log[ni:]
		}

		go c.sendAppendEntries(peerID, AppendEntriesRequest{
			Term:         n.currentTerm,
			LeaderID:     n.id,
			PrevLogIndex: prevIdx,
			PrevLogTerm:  prevTerm,
			Entries:      entries,
			LeaderCommit: n.commitIndex,
		})
	}
}

// handleAppendEntries 处理日志追加请求
func (n *RaftNode) handleAppendEntries(req AppendEntriesRequest) AppendEntriesResponse {
	// 如果请求的 Term 更大，更新自己
	if req.Term > n.currentTerm {
		n.currentTerm = req.Term
		n.role = Follower
		n.votedFor = -1
	}

	// 拒绝旧 Term 的请求
	if req.Term < n.currentTerm {
		return AppendEntriesResponse{Term: n.currentTerm, Success: false, NodeID: n.id}
	}

	// 收到合法 Leader 的请求，转为 Follower
	n.role = Follower

	// 追加日志条目
	if len(req.Entries) > 0 {
		for _, entry := range req.Entries {
			if entry.Index < len(n.log) {
				// 覆盖冲突的日志
				n.log[entry.Index] = entry
			} else {
				n.log = append(n.log, entry)
			}
		}
		n.logger("复制了 %d 条日志，来自 Leader %d", len(req.Entries), req.LeaderID)
	}

	// 更新 commitIndex
	if req.LeaderCommit > n.commitIndex {
		lastIdx := len(n.log) - 1
		if req.LeaderCommit < lastIdx {
			n.commitIndex = req.LeaderCommit
		} else {
			n.commitIndex = lastIdx
		}
	}

	return AppendEntriesResponse{Term: n.currentTerm, Success: true, NodeID: n.id}
}

// handleAppendResponse Leader 处理日志追加响应
func (n *RaftNode) handleAppendResponse(resp AppendEntriesResponse) {
	if n.role != Leader {
		return
	}

	if resp.Term > n.currentTerm {
		n.currentTerm = resp.Term
		n.role = Follower
		n.votedFor = -1
		return
	}

	if resp.Success {
		// 更新 nextIndex 和 matchIndex
		n.matchIndex[resp.NodeID] = len(n.log) - 1
		n.nextIndex[resp.NodeID] = len(n.log)

		// 检查是否可以提交新的日志
		n.tryCommit()
	}
}

// tryCommit Leader 尝试提交日志
// 如果多数节点已复制某条日志，则提交该日志
func (n *RaftNode) tryCommit() {
	for i := len(n.log) - 1; i > n.commitIndex; i-- {
		if n.log[i].Term != n.currentTerm {
			continue // 只提交当前 Term 的日志
		}

		replicaCount := 1 // 算上自己
		for _, peerID := range n.peers {
			if n.matchIndex[peerID] >= i {
				replicaCount++
			}
		}

		majority := (len(n.peers)+1)/2 + 1
		if replicaCount >= majority {
			n.commitIndex = i
			n.logger("提交日志 Index=%d, Command=%q（%d/%d 节点确认）",
				i, n.log[i].Command, replicaCount, len(n.peers)+1)
			break
		}
	}
}

// ProposeCommand Leader 提议一个新命令
func (n *RaftNode) ProposeCommand(command string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.role != Leader {
		return false
	}

	entry := LogEntry{
		Term:    n.currentTerm,
		Index:   len(n.log),
		Command: command,
	}
	n.log = append(n.log, entry)
	n.logger("追加日志: Index=%d, Command=%q", entry.Index, command)
	return true
}

// ============================================================
// 演示
// ============================================================

func main() {
	fmt.Println("=== Raft 一致性算法模拟 ===")
	fmt.Println()

	// 场景一：Leader 选举
	demoLeaderElection()

	// 场景二：日志复制
	demoLogReplication()
}

// demoLeaderElection 演示 Leader 选举过程
func demoLeaderElection() {
	fmt.Println("--- 场景一：Leader 选举 ---")
	fmt.Println("3 个节点的集群，观察选举过程")
	fmt.Println()

	cluster := NewRaftCluster(3)

	// 运行 2 秒，观察选举
	cluster.Run(2 * time.Second)

	// 打印最终状态
	fmt.Println()
	fmt.Println("  最终状态:")
	for id, node := range cluster.nodes {
		node.mu.Lock()
		fmt.Printf("    Node %d: Role=%s, Term=%d, VotedFor=%d, LogLen=%d\n",
			id, node.role, node.currentTerm, node.votedFor, len(node.log))
		node.mu.Unlock()
	}
	fmt.Println()
}

// demoLogReplication 演示日志复制过程
func demoLogReplication() {
	fmt.Println("--- 场景二：日志复制 ---")
	fmt.Println("Leader 接收写请求，复制到 Follower")
	fmt.Println()

	cluster := NewRaftCluster(3)

	// 先运行 1 秒完成选举
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		cluster.Run(3 * time.Second)
	}()

	// 等待选举完成
	time.Sleep(1 * time.Second)

	// 找到 Leader 并提交命令
	for _, node := range cluster.nodes {
		node.mu.Lock()
		isLeader := node.role == Leader
		node.mu.Unlock()

		if isLeader {
			fmt.Printf("  找到 Leader: Node %d\n", node.id)

			// 提交几个命令
			commands := []string{"SET x=1", "SET y=2", "SET z=3"}
			for _, cmd := range commands {
				if node.ProposeCommand(cmd) {
					fmt.Printf("  提交命令: %s\n", cmd)
				}
				time.Sleep(200 * time.Millisecond)
			}
			break
		}
	}

	wg.Wait()

	// 打印最终状态
	fmt.Println()
	fmt.Println("  最终状态:")
	for id, node := range cluster.nodes {
		node.mu.Lock()
		fmt.Printf("    Node %d: Role=%s, Term=%d, LogLen=%d, CommitIndex=%d\n",
			id, node.role, node.currentTerm, len(node.log), node.commitIndex)
		if len(node.log) > 0 {
			fmt.Printf("      日志: ")
			for _, entry := range node.log {
				fmt.Printf("[%d:%s] ", entry.Index, entry.Command)
			}
			fmt.Println()
		}
		node.mu.Unlock()
	}
	fmt.Println()
}
