package prefetcher

import "fmt"

const MAX_ENTRIES int = 5

type circulairArray struct {
	data [MAX_ENTRIES]string
	/* index of the first entry */
	start int
	/* index of the first free spot */
	end int
}

func newCirculairArray() circulairArray {
	return circulairArray{start: 0, end: 0}
}

func getNewPos(currPos int, dist int) int {
	return (currPos + dist) % MAX_ENTRIES
}

func (ca *circulairArray) get(index int) string {
	/* empty array or index out of range */
	if ca.start == ca.end || index >= MAX_ENTRIES {
		fmt.Println("END")
		return ""
	}
	return ca.data[getNewPos(ca.start, index)]
}

func (ca *circulairArray) removeFront() {
	/* empty array */
	if ca.start == ca.end {
		return
	}

	ca.start = getNewPos(ca.start, 1)
}

func (ca *circulairArray) pushBack(el string) {
	/* array is full, remove first element */
	if ca.end == ca.start {
		ca.removeFront()
	}

	fmt.Println(ca.end)
	ca.data[ca.end] = el
	ca.end = getNewPos(ca.end, 1)
	fmt.Println(ca.end)

}

// TODO return as pointer and dont append to array? this might be slow
func (ca *circulairArray) getAll() []string {

	fmt.Println("data: ", ca.data)
	d := []string{}

	for i := 0; i < MAX_ENTRIES; i++ {
		if ca.get(i) == "" {
			break
		}
		d = append(d, ca.get(i))
	}
	return d
}
