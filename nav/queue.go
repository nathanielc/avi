package nav

// Queue is a basic FIFO queue based on a circular list that resizes as needed.
type queue struct {
	nodes []*Waypoint
	head  int
	tail  int
	count int
}

// Push adds a node to the queue.
func (q *queue) Push(n *Waypoint) {
	if q.head == q.tail && q.count > 0 {
		nodes := make([]*Waypoint, len(q.nodes)*2)
		copy(nodes, q.nodes[q.head:])
		copy(nodes[len(q.nodes)-q.head:], q.nodes[:q.head])
		q.head = 0
		q.tail = len(q.nodes)
		q.nodes = nodes
	}
	q.nodes[q.tail] = n
	q.tail = (q.tail + 1) % len(q.nodes)
	q.count++
}

// Pop removes and returns a node from the queue in first to last order.
func (q *queue) Pop() *Waypoint {
	if q.count == 0 {
		return nil
	}
	node := q.nodes[q.head]
	q.head = (q.head + 1) % len(q.nodes)
	q.count--
	return node
}
