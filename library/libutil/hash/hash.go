package hash

func JenkinsOneAtTimeHash(s string) (hash uint32) {
	b := []byte(s)
	for i := range b {
		hash += uint32(b[i])
		hash += hash << 10
		hash ^= hash >> 6
	}
	hash += hash << 3
	hash ^= hash >> 11
	hash += hash << 15
	return
}
