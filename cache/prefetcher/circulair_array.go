package prefetcher

type circulairArray struct {
	data [MAX_TRACE_SIZE + 1]string
	/* index of the first entry */
	start int
	/* index of the first free spot */
	end   int
	empty bool
}

func newCirculairArray() circulairArray {
	return circulairArray{start: 0, end: 0, empty: true}
}

func getNewPos(currPos int, dist int) int {
	return (currPos + dist) % MAX_TRACE_SIZE
}

func (ca *circulairArray) get(index int) string {
	/* empty array or index out of range */
	if ca.empty || index >= MAX_TRACE_SIZE {
		return ""
	}
	return ca.data[getNewPos(ca.start, index)]
}

func (ca *circulairArray) removeFront() {
	if ca.empty {
		return
	}

	ca.start = getNewPos(ca.start, 1)

	if ca.start == ca.end {
		ca.empty = true
	}
}

func (ca *circulairArray) pushBack(el string) {
	/* array is full, remove first element */
	if !ca.empty && ca.end == ca.start {
		ca.removeFront()
	}

	ca.data[ca.end] = el
	ca.end = getNewPos(ca.end, 1)

	ca.empty = false
}

// TODO return as pointer and dont append to array? this might be slow
func (ca *circulairArray) getDataAsSlice() []string {

	d := []string{}

	for i := 0; i < len(ca.data); i++ {
		if ca.get(i) == "" {
			break
		}
		d = append(d, ca.get(i))
	}
	return d
}

// TODO is this used?
func (ca *circulairArray) len() int {
	if ca.empty {
		return 0
	}
	if ca.end > ca.start {
		return ca.end - ca.end
	}

	return MAX_TRACE_SIZE - ca.start + ca.end
}

func (ca *circulairArray) flush() {
	ca.start = 0
	ca.end = 0
	ca.empty = true
}

// TODO is flush neccesary?
func (ca *circulairArray) writeData(data *[]string) {
	ca.flush()
	for _, d := range *data {
		ca.pushBack(d)
	}
}
