package numbers

func BytesToInt(bytes []byte) int {
    result := 0
    for i := 0; i < len(bytes); i++ {
        result = result << 8
        result += int(bytes[i])

    }

    return result
}